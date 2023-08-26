package like

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Service) ActRelationInfo(c context.Context, req *api.ActRelationInfoReq) (*api.ActRelationInfoReply, error) {
	GRPCReply := new(api.ActRelationInfoReply)
	HTTPRes, err := s.GetActRelationInfo(c, req.Id, req.Mid, req.Specific)
	if err != nil {
		return nil, err
	}

	GRPCReply.Name = HTTPRes.Name
	GRPCReply.NativeIDs = HTTPRes.NativeIDs
	GRPCReply.H5IDs = HTTPRes.H5IDs
	GRPCReply.WebIDs = HTTPRes.WebIDs
	GRPCReply.LotteryIDs = HTTPRes.LotteryIDs
	GRPCReply.ReserveIDs = HTTPRes.ReserveIDs
	GRPCReply.VideoSourceIDs = HTTPRes.VideoSourceIDs
	GRPCReply.NativeID = HTTPRes.NativeID
	GRPCReply.ReserveID = HTTPRes.ReserveID

	if HTTPRes.ReserveItem != nil {
		item := new(api.ActRelationInfoReserveItem)

		item.Sid = HTTPRes.ReserveItem.Sid
		item.Name = HTTPRes.ReserveItem.Name
		item.Total = HTTPRes.ReserveItem.Total
		item.State = HTTPRes.ReserveItem.State
		item.StartTime = HTTPRes.ReserveItem.StartTime
		item.EndTime = HTTPRes.ReserveItem.EndTime
		item.ActStatus = HTTPRes.ReserveItem.ActStatus

		GRPCReply.ReserveItem = item
	}

	if HTTPRes.ReserveItems != nil {
		items := new(api.ActRelationInfoReserveItems)
		items.Total = HTTPRes.ReserveItems.Total
		items.State = HTTPRes.ReserveItems.State
		if len(HTTPRes.ReserveItems.ReserveList) > 0 {
			items.ReserveList = make([]*api.ActRelationInfoReserveItem, 0)
			for _, v := range HTTPRes.ReserveItems.ReserveList {
				item := new(api.ActRelationInfoReserveItem)

				item.Sid = v.Sid
				item.Total = v.Total
				item.State = v.State
				item.StartTime = v.StartTime
				item.EndTime = v.EndTime
				item.ActStatus = v.ActStatus

				items.ReserveList = append(items.ReserveList, item)
			}
		}
		GRPCReply.ReserveItems = items
	}

	return GRPCReply, nil
}

func (s *Service) ActRelationReserve(c context.Context, req *api.ActRelationReserveReq) (*api.ActRelationReserveReply, error) {
	var (
		failedMsgStore     sync.Map
		hasFailed          = false
		err                error
		actRelationSubject *like.ActRelationInfo
	)
	actRelationSubject, err = s.GetActRelationInfoByOptimization(c, req.Id)
	if err != nil {
		log.Errorc(c, "ActRelationReserve GetActRelationInfoByOptimization Err %v", err)
		return nil, err
	}
	if actRelationSubject.ID <= 0 {
		log.Errorc(c, "ActRelationReserve ActRelationInfo No Exist req:%+v reply:%+v", req, actRelationSubject)
		return nil, ecode.ActivityRelationIDNoExistErr
	}
	if actRelationSubject.ReserveIDs == "" {
		log.Errorc(c, "ActRelationReserve res.ReserveIDs empty req:%+v reply:%+v", req, actRelationSubject)
		return nil, errors.New("no exist reserveIDs")
	}
	reserveIDs := strings.Split(actRelationSubject.ReserveIDs, ",")
	if len(reserveIDs) == 0 {
		log.Errorc(c, "ActRelationReserve res.ReserveIDs split reserveIDs empty req:%+v reply:%+v", req, actRelationSubject)
		return nil, errors.New("split reserveIDs empty")
	}

	storeReserveIDs := make([]int64, 0)
	for _, v := range reserveIDs {
		reserveID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(c, "ActRelationReserve Format reserveIDs Err reserveID:%v err:%v", reserveID, err)
			return nil, errors.New("reserve failed")
		}
		storeReserveIDs = append(storeReserveIDs, reserveID)
	}

	// 并发
	g := errgroup.WithContext(c)
	for _, reserveID := range storeReserveIDs {
		sid := reserveID
		g.Go(func(c context.Context) (e error) {
			err := s.AsyncReserve(c, sid, req.Mid, 1, &like.ReserveReport{
				From:     req.From,
				Typ:      req.Typ,
				Oid:      req.Oid,
				Ip:       req.Ip,
				Platform: req.Platform,
				Mobiapp:  req.Mobiapp,
				Buvid:    req.Buvid,
				Spmid:    req.Spmid,
			})
			if err != ecode.ActivityRepeatSubmit && err != nil {
				// 预定失败记录信息到log
				hasFailed = true
				failedMsgStore.LoadOrStore(sid, err.Error())
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Errorc(c, "Goroutine Do AsyncReserve Err err:%v", err)
		return nil, errors.New("reserve failed")
	}
	// 打印预定失败信息
	if hasFailed == true {
		errLog := make(map[int64]string, 0)
		failedMsgStore.Range(func(key, value interface{}) bool {
			errLog[key.(int64)] = "errno:" + value.(string)
			return true
		})
		log.Errorc(c, "Do AsyncReserve Return Failed %+v", errLog)
		return nil, errors.New("reserve failed")
	}

	return &api.ActRelationReserveReply{
		State: 1,
	}, nil

}

func (s *Service) ActRelationReserveInfo(c context.Context, req *api.ActRelationReserveInfoReq) (*api.ActRelationReserveInfoReply, error) {
	ts := time.Now().Unix()
	var (
		actRelationSubject *like.ActRelationInfo
		err                error
	)
	actRelationSubject, err = s.GetActRelationInfoByOptimization(c, req.Id)
	if err != nil {
		log.Errorc(c, "ActRelationReserveInfo GetActRelationInfoByOptimization Err %v", err)
		return nil, err
	}
	if actRelationSubject.ID <= 0 {
		log.Errorc(c, "ActRelationReserveInfo ActRelationInfo No Exist req:%+v reply:%+v", req, actRelationSubject)
		return nil, ecode.ActivityRelationIDNoExistErr
	}
	if actRelationSubject.ReserveIDs == "" {
		log.Errorc(c, "ActRelationReserveInfo res.ReserveIDs empty req:%+v reply:%+v", req, actRelationSubject)
		return nil, errors.New("no exist reserveIDs")
	}
	reserveIDs := strings.Split(actRelationSubject.ReserveIDs, ",")
	if len(reserveIDs) == 0 {
		log.Errorc(c, "ActRelationReserveInfo res.ReserveIDs split reserveIDs empty req:%+v reply:%+v", req, actRelationSubject)
		return nil, errors.New("split reserveIDs empty")
	}

	storeReserveIDs := make([]int64, 0)
	for _, v := range reserveIDs {
		reserveID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(c, "ActRelationReserveInfo Format reserveIDs Err reserveID:%v err:%v", reserveID, err)
			return nil, errors.New("get info failed")
		}
		storeReserveIDs = append(storeReserveIDs, reserveID)
	}

	// 获取预约sid信息
	var actSubjects map[int64]*like.SubjectItem
	actSubjects, err = s.GetActSubjectsInfoByOptimization(c, storeReserveIDs)
	if err != nil {
		log.Errorc(c, "ActRelationReserveInfo GetActSubjectsInfoByOptimization Err %v", err)
		return nil, err
	}

	var data map[int64]*like.ActFollowingReply
	data, err = s.ReserveFollowings(c, storeReserveIDs, req.Mid)
	if err != nil {
		log.Errorc(c, "ActRelationReserveInfo ReserveFollowings Err reserveID:%v mid:%v err:%v", storeReserveIDs, req.Mid, err)
		return nil, errors.New("get info failed")
	}

	var res *api.ActRelationReserveInfoReply
	if len(data) > 0 {
		res = &api.ActRelationReserveInfoReply{
			State: 1, // 默认预约成功
		}
		for _, sid := range storeReserveIDs {
			if v, ok := data[sid]; ok {
				// 从subject中获取活动信息
				item := new(api.ActRelationReserveItem)
				if sub, ok := actSubjects[sid]; ok {
					item.Sid = sid
					item.StartTime = sub.Stime.Time().Unix()
					item.EndTime = sub.Etime.Time().Unix()
					item.Total = v.Total

					item.ActStatus = 0 // 0 活动未开始 1 活动已开始 2 活动已结束
					if ts >= item.StartTime && ts <= item.EndTime {
						item.ActStatus = 1
					}
					if ts > item.EndTime {
						item.ActStatus = 2
					}

					item.State = 1 // 默认已预约
					if !v.IsFollowing {
						item.State = int64(0) // 未预约
						res.State = 0         // 有一个未预约 最外层就是未预约
					}

					res.List = append(res.List, item)
				}
			}

			// 去除第一个数据 来填充到最外层
			if len(res.List) > 0 {
				outSide := (res.List)[0]
				res.Sid = outSide.Sid
				res.StartTime = outSide.StartTime
				res.EndTime = outSide.EndTime
				res.Total = outSide.Total
				res.ActStatus = outSide.ActStatus
			}
		}

	}

	return res, nil
}

func (s *Service) GRPCDoRelation(c context.Context, req *api.GRPCDoRelationReq) (*api.NoReply, error) {
	report := &like.ReserveReport{
		From:     req.From,
		Typ:      req.Typ,
		Oid:      req.Oid,
		Ip:       req.Ip,
		Platform: req.Platform,
		Mobiapp:  req.Mobiapp,
		Buvid:    req.Buvid,
		Spmid:    req.Spmid,
	}
	if req.Ip == "" {
		req.Ip = metadata.String(c, metadata.RemoteIP)
	}

	if res, err := s.DoRelation(c, req.Id, req.Mid, report); err != nil || res != 1 {
		log.Errorc(c, "GRPCDoRelation Error req:%+v err:%v", req, err)
		return new(api.NoReply), ecode.ActivityDoRelationErr
	}

	return new(api.NoReply), nil
}

func (s *Service) RelationReserveCancel(c context.Context, req *api.RelationReserveCancelReq) (*api.NoReply, error) {
	report := &like.ReserveReport{
		From:     req.From,
		Typ:      req.Typ,
		Oid:      req.Oid,
		Ip:       req.Ip,
		Platform: req.Platform,
		Mobiapp:  req.Mobiapp,
		Buvid:    req.Buvid,
		Spmid:    req.Spmid,
	}
	if req.Ip == "" {
		req.Ip = metadata.String(c, metadata.RemoteIP)
	}

	if err := s.DoRelationReserveCancel(c, req.Id, req.Mid, report); err != nil {
		log.Errorc(c, "DoRelationReserveCancel Error req:%+v err:%v", req, err)
		return new(api.NoReply), ecode.RelationReserveCancelErr
	}

	return new(api.NoReply), nil
}

// 定时将数据库中热门数据IDs集合写入cache中
func (s *Service) InternalSyncActRelationInfoDB2Cache(ctx context.Context, req *api.InternalSyncActRelationInfoDB2CacheReq) (reply *api.InternalSyncActRelationInfoDB2CacheReply, err error) {
	var (
		IDs []int64
		b   = []byte("")
	)
	reply = new(api.InternalSyncActRelationInfoDB2CacheReply)
	IDs, err = s.dao.HotGetActRelationInfo(ctx)
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActRelationInfoDB2Cache->HotGetActRelationInfo Source(%v) Err(%v)", req.From, err)
		return
	}
	if len(IDs) > 0 {
		b, err = json.Marshal(IDs)
	}
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActRelationInfoDB2Cache->json.Marshal Source(%v) Err(%v)", req.From, err)
		return
	}
	err = s.dao.HotAddActRelationInfoSet(ctx, string(b))
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActRelationInfoDB2Cache->HotAddActRelationInfoKey Source(%v) Err(%v)", req.From, err)
		return
	}
	log.Infoc(ctx, "[HOT-DATA-SUCC]InternalSyncActRelationInfoDB2Cache Res(%v)", string(b))
	return
}

// 定时将有效的id信息加载到内存中
func (s *Service) InternalGetActRelationInfoFromCacheSetInfoMemory() {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			res, err := s.DeserializeActRelationInfoFromCache(context.Background())
			if err != nil {
				log.Error("[HOT-DATA-FAIL]InternalGetActRelationInfoFromCacheSetInfoMemory->DeserializeActRelationInfoFromCache Err(%v)", err)
				continue
			}
			s.HotActRelationInfoStore = res
		}
	}
}

// 服务首次加载 将有效的id信息加载到内存中 失败了需要panic
func (s *Service) initInternalGetActRelationInfoFromCacheSetInfoMemory() error {
	ctx := context.Background()
	var (
		err error
		res map[int64]*like.ActRelationInfo
	)
	// 初始化需要自己查询DB写入缓存 否则首次上线因为缓存没有数据会panic
	_, err = s.InternalSyncActRelationInfoDB2Cache(ctx, &api.InternalSyncActRelationInfoDB2CacheReq{
		From: "init",
	})
	if err != nil {
		return err
	}
	// 首次加载 先初始化 避免nil panic
	s.HotActRelationInfoStore = make(map[int64]*like.ActRelationInfo, 0)
	// 内存反序列化数据
	res, err = s.DeserializeActRelationInfoFromCache(ctx)
	if err != nil {
		return err
	}
	s.HotActRelationInfoStore = res
	return nil
}

// 定时将redis中有效id集合拿出来进行反序列化 在通过循环，将所有id信息拿出来进行返回
func (s *Service) DeserializeActRelationInfoFromCache(ctx context.Context) (map[int64]*like.ActRelationInfo, error) {
	res := make(map[int64]*like.ActRelationInfo, 0)

	str, err := s.dao.HotGetActRelationInfoSet(ctx)
	if err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActRelationInfoFromCache->HotGetActRelationInfoSet Err(%v)", err)
		return res, err
	}

	if str == "" {
		return res, nil
	}

	var IDs []int64
	if err := json.Unmarshal([]byte(str), &IDs); err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]InternalGetActRelationInfoFromCacheSetInfoMemory->json.Unmarshal([]byte(str), &IDs) Err(%v)", err)
		return res, err
	}

	// 循环获取所有id基本信息
	for _, id := range IDs {
		item, err := s.dao.GetActRelationInfo(ctx, id)
		if err != nil {
			err = fmt.Errorf("[HOT-DATA-FAIL]InternalGetActRelationInfoFromCacheSetInfoMemory->s.dao.GetActRelationInfo Err(%v) ID(%v)", err, id)
			return res, err
		}
		res[item.ID] = item
	}

	return res, nil
}

// 性能优化版 - 从内存中获取有效的活动信息 获取不到回源
func (s *Service) GetActRelationInfoByOptimization(ctx context.Context, id int64) (res *like.ActRelationInfo, err error) {
	res = new(like.ActRelationInfo)
	if v, ok := s.HotActRelationInfoStore[id]; ok {
		res = v
		return
	}
	var subject *like.ActRelationInfo
	subject, err = s.dao.GetActRelationInfo(ctx, id)
	if err != nil {
		log.Errorc(ctx, "GetActRelationInfoByOptimization->GetActRelationInfo Err(%v)", err)
		return
	}
	if subject != nil {
		res = subject
	}
	return
}

// 总入口 强刷单条cache key 目前有actrelation和actsubject
func (s *Service) InternalUpdateItemDataWithCache(ctx context.Context, req *api.InternalUpdateItemDataWithCacheReq) (reply *api.InternalUpdateItemDataWithCacheReply, err error) {
	if req.Typ == 0 {
		err = fmt.Errorf("InternalUpdateItemDataWithCache Illigel Typ(%v)", req.Typ)
		return
	}
	if req.ActionType == 0 {
		err = fmt.Errorf("InternalUpdateItemDataWithCache Illigel ActionType(%v)", req.ActionType)
		return
	}
	if req.Oid == 0 {
		err = fmt.Errorf("InternalUpdateItemDataWithCache Illigel Oid(%v)", req.Oid)
		return
	}

	reply = new(api.InternalUpdateItemDataWithCacheReply)
	switch req.Typ {
	case like.ActRelationFlushItemInfo2Cache: // 活动聚合平台信息
		switch req.ActionType {
		case like.Update: // 更新缓存
			if err := s.InternalGetActRelationItemFromDB2FlushIntoCache(ctx, req.Oid); err != nil {
				return reply, err
			}
		case like.Delete: // 删除缓存
			if err := s.InternalDeleteActRelationItemCache(ctx, req.Oid); err != nil {
				return reply, err
			}
		case like.Unable: // 失效缓存
			if err := s.InternalUnableActRelationItemCache(ctx, req.Oid); err != nil {
				return reply, err
			}
		}
	case like.ActSubjectFlushItemInfo2Cache: // actSubject
		switch req.ActionType {
		case like.Update: // 更新缓存
			if err := s.InternalGetActSubjectItemFromDB2FlushIntoCache(ctx, req.Oid); err != nil {
				return reply, err
			}
		}
	}

	return reply, nil
}

// 编辑更新admin会进行强刷单条 cache key 失败会阻止提交表单 缓存过期时间为 凌晨3-5点后的随机时间
func (s *Service) InternalGetActRelationItemFromDB2FlushIntoCache(ctx context.Context, id int64) error {
	item, err := s.dao.RawGetActRelationInfo(ctx, id)
	if err != nil {
		return fmt.Errorf("InternalGetActRelationInfoFromDB2FlushIntoCache->s.dao.RawGetActRelationInfo ID(%v) Err(%v)", id, err)
	}
	if item.ID <= 0 {
		return fmt.Errorf("InternalGetActRelationInfoFromDB2FlushIntoCache->s.dao.RawGetActRelationInfo No Result ID(%v)", item.ID)
	}
	for i := 0; i < 3; i++ {
		err = s.dao.AddCacheGetActRelationInfo(ctx, id, item)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("InternalGetActRelationItemFromDB2FlushIntoCache->AddCacheGetActRelationInfo ID(%v) Err(%v)", id, err)
	}
	return err
}

// 添加admin会进行删除缓存
func (s *Service) InternalDeleteActRelationItemCache(ctx context.Context, id int64) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.dao.DelCacheGetActRelationInfo(ctx, id)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("InternalDeleteActRelationItemCache->DelCacheGetActRelationInfo ID(%v) Err(%v)", id, err)
	}
	return err
}

// 删除admin会进行失效缓存
func (s *Service) InternalUnableActRelationItemCache(ctx context.Context, id int64) error {
	var err error
	item := &like.ActRelationInfo{
		ID: 0,
	}
	for i := 0; i < 3; i++ {
		err = s.dao.AddCacheGetActRelationInfo(ctx, id, item)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("InternalUnableActRelationItemCache->AddCacheGetActRelationInfo ID(%v) Item(%v) Err(%v)", id, item, err)
	}
	return err
}
