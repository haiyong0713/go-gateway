package common

import (
	"context"
	"encoding/json"
	"strconv"

	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	dynamicmdl "go-gateway/app/app-svr/app-car/interface/model/dynamic"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	followgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
)

const (
	refreshTypeNew     = 0
	refreshTypeHistory = 1

	_assistBaseline = "20"
)

func (s *Service) Dynamic(c context.Context, req *commonmdl.DynamicReq, mid int64, buvid string) (resp *commonmdl.DynamicResp, err error) { // nolint:gocognit
	// 客户端期望不要有null
	resp = &commonmdl.DynamicResp{
		Uplist:      []*commonmdl.Uplist{},
		DynamicList: []*commonmdl.Item{},
	}
	// 翻页前置逻辑
	var (
		refreshType                   int
		historyOffset, updateBaseline string
		page                          int64
	)
	if req.PageNext == "" {
		refreshType = refreshTypeNew
		updateBaseline = ""
	} else {
		var pageNext = new(commonmdl.DynamicPageNext)
		refreshType = refreshTypeHistory
		if err = json.Unmarshal([]byte(req.PageNext), &pageNext); err != nil {
			log.Warn("Dynamic json.Unmarshal(%v) error(%v)", req.PageNext, err)
		}
		if pageNext != nil {
			historyOffset = pageNext.HistoryOffset
			page = pageNext.Page
		}
	}
	//获取用户关注链信息(关注的up、follow的PGC内容）
	following, pgcFollowing, err := s.followings(c, mid, false)
	if err != nil {
		log.Error("Dynamic followings(%v) error(%+v)", mid, err)
		return
	}
	// 没有关系链的情况不返回动态列表
	if len(following) == 0 && len(pgcFollowing) == 0 {
		log.Error("Dynamic following(%v) pgcFollowing(%v)", following, pgcFollowing)
		return
	}
	attentions := dynamicmdl.GetAttentionsParams(mid, following, pgcFollowing)
	// 获取动态列表
	var (
		tmpDyns   *dynamicmdl.DynVideoListRes
		tmpUplist *dynamicmdl.VdUpListRsp
	)
	eg := errgroup.WithContext(c)
	// 动态服务端
	eg.Go(func(ctx context.Context) error {
		var (
			dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
			errTmp      error
		)
		if req.Vmid == 0 {
			switch refreshType {
			case refreshTypeHistory:
				tmpDyns, errTmp = s.dynDao.DynVideoHistory(c, mid, historyOffset, page, dynTypeList, attentions, req.Build, req.Platform, req.MobiApp, buvid, req.Device)
				if errTmp != nil {
					log.Error("Dynamic DynVideoHistory(%+v) error(%+v)", req, errTmp)
				}
			default:
				tmpDyns, errTmp = s.dynDao.DynVideoList(c, mid, updateBaseline, _assistBaseline, dynTypeList, attentions, req.Build, req.Platform, req.MobiApp, buvid, req.Device)
				if errTmp != nil {
					log.Error("Dynamic DynVideoList(%+v) error(%+v)", req, errTmp)
				}
			}
		} else {
			tmpDyns, errTmp = s.dynDao.DynVideoPersonal(c, req.Vmid, mid, false, historyOffset, strconv.Itoa(req.Build), req.Platform, req.MobiApp, buvid, req.Device, "", "", "", attentions, dynTypeList)
			if errTmp != nil {
				log.Error("Dynamic DynVideoPersonal(%v) error(%+v)", req.Vmid, errTmp)
			}
		}
		return errTmp
	})
	if req.NeedUplist {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			tmpUplist, errTmp = s.dynDao.VdUpList(ctx, 0, mid, buvid)
			if errTmp != nil {
				log.Error("Dynamic VdUpList(%v, %v, %v) error(%+v)", 0, mid, buvid, errTmp)
				return errTmp
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("Dynamic error(%v)", err)
		return
	}
	if tmpDyns == nil || len(tmpDyns.Dynamics) == 0 {
		return
	}
	// 获取物料
	var (
		aidm  = make(map[int64]struct{})
		epidm = make(map[int32]struct{})
	)
	for _, tmpDyn := range tmpDyns.Dynamics {
		if tmpDyn == nil {
			continue
		}
		if commonmdl.DynamicIsUGC(tmpDyn.Type) {
			aidm[tmpDyn.Rid] = struct{}{}
		}
		if commonmdl.DynamicIsOGV(tmpDyn.Type) {
			epidm[int32(tmpDyn.Rid)] = struct{}{}
		}
	}
	var uidm = make(map[int64]struct{})
	if tmpUplist != nil && len(tmpUplist.Items) > 0 {
		for _, tmpUp := range tmpUplist.Items {
			if tmpUp.UID != 0 {
				uidm[tmpUp.UID] = struct{}{}
			}
		}
	}
	// 获取物料
	var materialParams = new(commonmdl.Params)
	if len(aidm) > 0 {
		materialParams.ArchiveReq = new(commonmdl.ArchiveReq)
		for aid := range aidm {
			var playAv = &archivegrpc.PlayAv{Aid: aid}
			materialParams.ArchiveReq.PlayAvs = append(materialParams.ArchiveReq.PlayAvs, playAv)
		}
	}
	if len(epidm) > 0 {
		var epids []int32
		for epid := range epidm {
			epids = append(epids, epid)
		}
		materialParams.EpisodeReq = new(commonmdl.EpisodeReq)
		materialParams.EpisodeReq.Epids = epids
	}
	if len(uidm) > 0 {
		var mids []int64
		for uid := range uidm {
			mids = append(mids, uid)
		}
		materialParams.AccountCardReq = new(commonmdl.AccountCardReq)
		materialParams.AccountCardReq.Mids = mids
	}
	carContext, err := s.material(c, materialParams, req.DeviceInfo)
	if err != nil {
		b, _ := json.Marshal(materialParams)
		log.Error("Dynamic material(%+v) error(%v)", string(b), err)
		return
	}
	// 聚合卡片
	for _, tmpDyn := range tmpDyns.Dynamics {
		carContext.OriginData = &commonmdl.OriginData{Oid: tmpDyn.Rid}
		if commonmdl.DynamicIsUGC(tmpDyn.Type) {
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeUGC
		}
		if commonmdl.DynamicIsOGV(tmpDyn.Type) {
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeOGVEP
		}
		item := s.formItem(carContext, req.DeviceInfo)
		if item != nil {
			resp.DynamicList = append(resp.DynamicList, item)
		}
	}
	// 聚合常看UP列表
	if req.NeedUplist {
		// 处理UP主部分
		if tmpUplist != nil {
			// 置顶 "全部"
			resp.Uplist = append(resp.Uplist, &commonmdl.Uplist{
				HasUpdate: false, Name: "全部",
				Face: "http://i0.hdslb.com/bfs/feed-admin/e39bf1fbc427ad0ff01f53a287186b7ccff4115e.png",
			})
			for _, tmpUp := range tmpUplist.Items {
				accountCard, ok := carContext.AccountCardResp[tmpUp.UID]
				if !ok || accountCard == nil {
					continue
				}
				var hasUpdate bool
				if tmpUp.HasUpdate == 1 {
					hasUpdate = true
				}
				resp.Uplist = append(resp.Uplist, &commonmdl.Uplist{
					HasUpdate: hasUpdate,
					Mid:       tmpUp.UID,
					Name:      accountCard.Name,
					Face:      accountCard.Face,
				})
			}
		}
	}
	// 分页后置逻辑
	resp.PageNext = new(commonmdl.DynamicPageNext)
	resp.PageNext.HistoryOffset = tmpDyns.HistoryOffset
	resp.PageNext.Page++
	resp.HasNext = tmpDyns.HasMore
	return
}

func (s *Service) followings(c context.Context, mid int64, isOGV bool) ([]*relationgrpc.FollowingReply, []*followgrpc.FollowSeasonProto, error) {
	eg := errgroup.WithCancel(c)
	var (
		follow []*relationgrpc.FollowingReply
		pgc    []*followgrpc.FollowSeasonProto
	)
	eg.Go(func(ctx context.Context) error {
		var err error
		follow, err = s.relationDao.Followings(ctx, mid)
		if err != nil {
			return err
		}
		return nil
	})
	if isOGV {
		eg.Go(func(ctx context.Context) error {
			var err error
			pgc, err = s.bangumiDao.MyRelations(ctx, mid)
			if err != nil {
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}
	return follow, pgc, nil
}
