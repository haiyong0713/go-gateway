package service

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"sync"

	locApi "git.bilibili.co/bapis/bapis-go/community/service/location"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	pb "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/model"
)

// 检查小名单的逻辑
func (s *Service) CheckSmallList(ctx context.Context, req *pb.CheckCommonBWListReq, scene *model.BWListWithGroup) (result bool, err error) {
	// 获取默认值key
	var defaultValue bool
	if scene.DefaultValue == 1 {
		defaultValue = true
	} else if scene.DefaultValue == 0 {
		defaultValue = false
	}

	// 获取地域key
	var areaPolicyId int64
	areaKey := fmt.Sprintf("bw_%s_area_%s", req.Token, req.Oid)

	// 从redis获取loc策略组id
	if areaPolicyId, err = s.show.GetBWListItemFromRedis(ctx, areaKey); err != nil {
		return false, err
	}

	// 如果loc策略组id为0，则直接放过
	if areaPolicyId == -1 {
		return defaultValue, nil
	} else if areaPolicyId == 0 {
		return true, nil
	} else {
		// 从参数获取user ip，如果没配置则从metadata获取
		userIp := req.UserIp
		if userIp == "" {
			userIp = metadata.String(ctx, metadata.RemoteIP)
		}
		// 将ip和策略组期望带给loc服务
		locReq := &locApi.ZoneLimitPoliciesReq{
			UserIp:       userIp,
			DefaultAuths: make(map[int64]*locApi.ZoneLimitAuth),
		}
		locReq.DefaultAuths[areaPolicyId] = &locApi.ZoneLimitAuth{
			Play: locApi.Status_Forbidden,
		}
		var reply *locApi.ZoneLimitPoliciesReply
		if reply, err = s.locGRPC.ZoneLimitPolicies(ctx, locReq); err != nil {
			log.Error("Failed to get ZoneLimitPolicies: %+v, ip: %s", err, userIp)
			return false, nil
		}
		if auth, ok := reply.Auths[areaPolicyId]; ok {
			return auth.Play == locApi.Status_Allow, nil
		}
	}

	return false, nil
}

// 检查大名单的逻辑
func (s *Service) CheckLargeList(ctx context.Context, req *pb.CheckCommonBWListReq, group *model.BWListWithGroup) (bool, error) {
	high := int64(group.High)
	low := int64(group.Low)
	target := int64(0)
	modBase := int64(100)
	var largeOid string

	//log.Error("[bw-list]CheckLargeList before, LargeOid.MID(%+v), LargeOid.BUVID(%+v)", req.LargeOid.Mid, req.LargeOid.Buvid)

	if req.LargeOid == nil {
		return false, ecode.Error(ecode.RequestErr, "large_oid is nil!")
	}
	switch group.LargeOidType {
	case model.OidType_MID:
		{
			if req.LargeOid.Mid == 0 {
				return group.ShowWithoutLogin == 1, nil
			}
			target = req.LargeOid.Mid % modBase
			largeOid = strconv.FormatInt(req.LargeOid.Mid, 10)
		}
	case model.OidType_BUVID:
		{
			if req.LargeOid.Buvid == "" {
				return false, nil
			}
			dataWithSalt := req.LargeOid.Buvid + "-" + req.Token
			target = int64(crc32.ChecksumIEEE([]byte(dataWithSalt)) % uint32(modBase))
			largeOid = req.LargeOid.Buvid
			//log.Error("[bw-list]CheckLargeList buvid, buvid(%s), target(%d)", req.LargeOid.Buvid, target)
		}
	default:
		return false, nil
	}

	// 判断是否在分组白名单中
	for _, oid := range group.WhiteList {
		if largeOid == oid {
			return true, nil
		}
	}
	//log.Error("[bw-list]CheckLargeList, low(%d), target(%d), high(%d)", low, target, high)
	if group.LargeListUrl != "" {
		apiResult, err := s.manager.CheckLargeList(ctx, req.LargeOid, group.LargeListUrl)
		if err != nil {
			log.Error("[bw-list]第三方接口调用错误：%s", err)
			return false, err
		}
		//log.Error("[bw-list]第三方接口调用结果：%+v", apiResult)
		switch group.SpecialOp {
		case model.SpecialOPNoGrayInApi:
			{
				return apiResult, nil
			}
		case model.SpecialOPNoGrayNotInApi:
			{
				return !apiResult, nil
			}
		default:
			{
				return low <= target && target < high && apiResult, nil
			}
		}
	}

	return low <= target && target < high, nil
}

type LargeOidContent struct {
	Mid   string `json:"mid"`
	Bvuid string `json:"buvid"`
}

// 通用黑白名单检查
func (s *Service) CheckCommonBWList(ctx context.Context, req *pb.CheckCommonBWListReq) (rep *pb.CheckCommonBWListRep, err error) {
	rep = new(pb.CheckCommonBWListRep)
	rep.IsInList = false
	errResult := ecode.Error(-77666, "请使用业务兜底逻辑")

	if req.Token == "" || req.Oid == "" {
		return nil, ecode.RequestErr
	}

	// 获取默认值key
	var (
		scene, group             *model.BWListWithGroup
		sOK, gOK                 bool
		smallResult, largeResult bool
	)
	if scene, sOK = s.bwListSceneTokenCache[req.Token]; !sOK {
		//log.Error("[bw-list]Failed to find token(%+v) from cache(%+v)", req.Token, s.bwListSceneTokenCache)
		return nil, ecode.RequestErr
	}

	switch scene.ListType {
	case model.ListType_SMALL:
		// 仅检查小名单
		if smallResult, err = s.CheckSmallList(ctx, req, scene); err != nil {
			//log.Error("[bw-list]Failed to get small list result: %+v", err)
			return nil, errResult
		}
		rep.IsInList = smallResult
	case model.ListType_LARGE:
		// 匹配检查，防止过多计算
		if req.LargeOid == nil || req.LargeToken == "" {
			largeResult = false
			break
		}
		if group, gOK = s.bwListGroupTokenCache[req.LargeToken]; !gOK {
			return nil, ecode.RequestErr
		}
		if group.SceneToken != scene.SceneToken {
			log.Error("[bw-list]s.bwListSceneTokenCache not equal, %s, %s", group.SceneToken, scene.SceneToken)
			return nil, ecode.RequestErr
		}

		// 由于大小名单都可能存在异步请求，所以使用errgroup
		eg := errgroup.WithContext(ctx)

		// 检查小名单
		eg.Go(func(c context.Context) error {
			var e error
			if smallResult, e = s.CheckSmallList(c, req, scene); e != nil {
				return e
			}
			return nil
		})

		// 检查大名单
		eg.Go(func(c context.Context) error {
			var e error
			if largeResult, e = s.CheckLargeList(c, req, group); e != nil {
				return e
			}
			return nil
		})
		if err = eg.Wait(); err != nil {
			//log.Error("[bw-list]Failed to get some list result: %+v", err)
			// 就算有一个命中成功，也返回true，尽可能减少错误造成的问题
			if !smallResult && !largeResult {
				return nil, errResult
			}
		}
		//log.Error("[bw-list]get some list result: %+v, %+v， %+v", req, smallResult, largeResult)

		rep.IsInList = smallResult || largeResult
	}

	// 正常返回整体判断是否取反
	if req.IsReverse {
		rep.IsInList = !rep.IsInList
	}
	return rep, nil

}

// 通用黑白名单检查
func (s *Service) CheckCommonBWListBatch(ctx context.Context, req *pb.CheckCommonBWListBatchReq) (rep *pb.CheckCommonBWListBatchRep, err error) {
	rep = new(pb.CheckCommonBWListBatchRep)
	rep.IsInList = make(map[string]bool)

	if req.Token == "" || len(req.Oids) == 0 {
		return nil, ecode.RequestErr
	}
	eg := errgroup.WithContext(ctx)

	lock := sync.Mutex{}
	for _, oid := range req.Oids {
		singleReq := &pb.CheckCommonBWListReq{
			Oid:       oid,
			Token:     req.Token,
			UserIp:    req.UserIp,
			IsReverse: req.IsReverse,
		}
		// 开协程去调用单个oid的检查接口，和业务方直接开协程调用单个检查接口，区别只在于请求N次的网络消耗
		eg.Go(func(c context.Context) error {
			var IsInList bool
			if singleRep, e := s.CheckCommonBWList(c, singleReq); e != nil {
				return nil
			} else {
				IsInList = singleRep.IsInList
			}
			lock.Lock()
			rep.IsInList[singleReq.Oid] = IsInList
			lock.Unlock()
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		log.Error("CheckCommonBWList batch fail, err: %+v", err)
	}

	return rep, nil
}

// 周期load缓存数据
func (s *Service) loadBWListWithGroupCache() {
	var (
		rows []*model.BWListWithGroup
		err  error
	)
	if rows, err = s.manager.GetBWListWithGroupFromDB(); err != nil {
		log.Error("loadBWListWithGroupCache fail, err: %+v", err)
		return
	}
	bwListSceneTokenCache := make(map[string]*model.BWListWithGroup)
	bwListGroupTokenCache := make(map[string]*model.BWListWithGroup)

	for _, row := range rows {
		bwListSceneTokenCache[row.SceneToken] = row
		if row.IsGroupDeleted == model.IsDeleted_NORMAL {
			bwListGroupTokenCache[row.GroupToken] = row
		}
	}

	s.bwListSceneTokenCache = bwListSceneTokenCache
	s.bwListGroupTokenCache = bwListGroupTokenCache
}
