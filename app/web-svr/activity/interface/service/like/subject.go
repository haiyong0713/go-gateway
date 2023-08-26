package like

import (
	"context"
	"encoding/json"
	"fmt"
	tagService "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
	errgroup2 "go-common/library/sync/errgroup.v2"
	api "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	xecode "go-gateway/app/web-svr/activity/ecode"
	ldao "go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/task"
)

// SubjectInitialize act_subject data initialize .
func (s *Service) SubjectInitialize(c context.Context, minSid int64) (err error) {
	if minSid < 0 {
		minSid = 0
	}
	var actSub []*like.SubjectItem
	for {
		if actSub, err = s.dao.SubjectListMoreSid(c, minSid); err != nil {
			log.Error("dao.subjectListMoreSid(%d) error(%+v)", minSid, err)
			break
		}
		// empty slice or nil
		if len(actSub) == 0 {
			log.Info("SubjectInitialize end success")
			break
		}
		for _, sub := range actSub {
			item := sub
			if minSid < item.ID {
				minSid = item.ID
			}
			id := item.ID
			//the activity offline is stored with empty data
			if item.State != ldao.SubjectValidState {
				item = &like.SubjectItem{}
			}
			s.cache.Do(c, func(c context.Context) {
				s.dao.AddCacheActSubject(c, id, item)
			})
		}
	}
	s.cache.Do(c, func(c context.Context) {
		s.SubjectMaxIDInitialize(c)
	})
	return
}

// SubjectMaxIDInitialize Initialize act_subject max id data .
func (s *Service) SubjectMaxIDInitialize(c context.Context) (err error) {
	var actSub *like.SubjectItem
	if actSub, err = s.dao.SubjectMaxID(c); err != nil {
		log.Error(" s.dao.SubjectMaxID() error(%+v)", err)
		return
	}
	if actSub.ID >= 0 {
		if err = s.dao.AddCacheActSubjectMaxID(c, actSub.ID); err != nil {
			log.Error("s.dao.AddCacheActSubjectMaxID(%d) error(%v)", actSub.ID, err)
		}
	}
	return
}

// SubjectUp up act_subject cahce info .
func (s *Service) SubjectUp(c context.Context, sid int64) (err error) {
	var (
		actSub   *like.SubjectItem
		maxSubID int64
	)
	group, ctx := errgroup.WithContext(c)
	group.Go(func() (e error) {
		if actSub, e = s.dao.RawActSubject(ctx, sid); e != nil {
			log.Error("dao.RawActSubject(%d) error(%+v)", sid, e)
		}
		return
	})
	group.Go(func() (e error) {
		if maxSubID, e = s.dao.CacheActSubjectMaxID(ctx); e != nil {
			log.Error("dao.RawActSubject(%d) error(%v)", sid, e)
		}
		return
	})
	if err = group.Wait(); err != nil {
		log.Error("SubjectUp error(%v)", err)
		return
	}
	if actSub.ID == 0 || actSub.State != ldao.SubjectValidState {
		actSub = &like.SubjectItem{}
	}
	if maxSubID < sid {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheActSubjectMaxID(context.Background(), sid)
		})
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.AddCacheActSubject(context.Background(), sid, actSub)
	})
	return
}

// SubjectLikeListInitialize Initialize likes list .
func (s *Service) SubjectLikeListInitialize(c context.Context, sid int64) (err error) {
	var (
		actSub *like.SubjectItem
		items  []*like.Item
		lid    = int64(0)
	)
	if actSub, err = s.dao.RawActSubject(c, sid); err != nil {
		log.Error("dao.RawActSubject(%d) error(%+v)", sid, err)
		return
	}
	if actSub.ID == 0 {
		log.Info("SubjectSLikeListInitialize end success")
		return
	}
	for {
		if items, err = s.dao.LikesBySid(c, lid, sid); err != nil {
			log.Error("dao.LikesBySid(%d,%d) error(%+v)", lid, sid, err)
			break
		}
		// empty slice or nil
		if len(items) == 0 {
			log.Info("SubjectSLikeListInitialize end success")
			break
		}
		//Initialize likes ctime cache
		cItems := items
		s.cache.Do(c, func(c context.Context) {
			s.dao.LikeListCtime(c, sid, cItems)
		})
		for _, val := range items {
			if lid < val.ID {
				lid = val.ID
			}
		}
	}
	return
}

// LikeActCountInitialize Initialize like_action cache data .
func (s *Service) LikeActCountInitialize(c context.Context, sid int64) (err error) {
	var (
		actSub  *like.SubjectItem
		items   []*like.Item
		lid     = int64(0)
		types   = make(map[int64]int64)
		lidLike map[int64]int64
	)
	if actSub, err = s.dao.RawActSubject(c, sid); err != nil {
		log.Error("dao.RawActSubject(%d) error(%+v)", sid, err)
		return
	}
	if actSub.ID == 0 {
		log.Info("SubjectSLikeListInitialize end success")
		return
	}
	for {
		if items, err = s.dao.LikesBySid(c, lid, sid); err != nil {
			log.Error("dao.LikesBySid(%d,%d) error(%+v)", lid, sid, err)
			break
		}
		if len(items) == 0 {
			log.Info("SubjectSLikeListInitialize end success")
			break
		}
		lidList := make([]int64, 0, len(items))
		for _, val := range items {
			if lid < val.ID {
				lid = val.ID
			}
			lidList = append(lidList, val.ID)
			types[val.ID] = val.Type
		}
		if lidLike, err = s.dao.LikeActSums(c, lidList); err != nil {
			log.Error(" s.dao.LikeActSums(%d,%v) error(%+v)", sid, lidList, err)
			return
		}
		rlyLike := make(map[int64]int64, 0)
		for _, v := range lidList {
			rlyLike[v] = 0
			if _, ok := lidLike[v]; ok {
				rlyLike[v] = lidLike[v]
			}
		}
		if len(rlyLike) == 0 {
			continue
		}
		if err = s.dao.SetInitializeLikeCache(c, sid, rlyLike, types); err != nil {
			log.Error("LikeActCountInitialize:eg.Wait() error(%+v)", err)
			return
		}
	}
	return
}

// ActSubjectWithAid ...
func (s *Service) ActSubjectWithAid(c context.Context, sid, aid int64) (res interface{}, err error) {
	var subject *like.SubjectItem
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		return
	}
	if subject == nil || subject.ID == 0 {
		err = ecode.NothingFound
	}
	var lid int64
	if lid, err = s.GetLidByWid(c, aid); err != nil {
		return
	}
	if lid == 0 {
		log.Warnc(c, "ActSubjectWithAid s.GetLidByWid(c, %d) return zero lid", aid)
	}
	return struct {
		*like.PubSub
		Lid int64 `json:"lid"`
	}{
		PubSub: subject.PublicData(),
		Lid:    lid,
	}, nil
}

// ActSubject .
func (s *Service) ActSubject(c context.Context, sid int64) (res *like.SubjectItem, err error) {
	if res, err = s.dao.ActSubject(c, sid); err != nil {
		return
	}
	if res == nil || res.ID == 0 {
		err = ecode.NothingFound
		return
	}
	return
}

// ActSubjects batch get subject.
func (s *Service) ActSubjects(c context.Context, sids []int64) (list map[int64]*like.SubjectItem, err error) {
	var (
		res map[int64]*like.SubjectItem
	)
	if res, err = s.dao.ActSubjects(c, sids); err != nil {
		return
	}
	list = make(map[int64]*like.SubjectItem, len(res))
	for _, v := range res {
		if v.ID > 0 {
			list[v.ID] = v
		}
	}
	return
}

// ActProtocol .
func (s *Service) ActProtocol(c context.Context, a *like.ArgActProtocol) (res *like.SubProtocol, err error) {
	res = new(like.SubProtocol)
	if res.SubjectItem, err = s.dao.ActSubject(c, a.Sid); err != nil {
		log.Error("s.dao.ActSubject() error(%+v)", err)
		return
	}
	if res.SubjectItem.ID == 0 {
		err = ecode.NothingFound
		return
	}
	now := time.Now().Unix()
	if int64(res.SubjectItem.Stime) <= now && int64(res.SubjectItem.Etime) >= now {
		if res.ActSubjectProtocol, err = s.dao.ActSubjectProtocol(c, a.Sid); err != nil {
			log.Error("s.dao.ActSubjectProtocol(%d) error(%+v)", a.Sid, err)
			return
		}
	}
	if res.SubjectItem.Type == like.CLOCKIN {
		res.Rules, err = s.dao.RawSubjectRulesBySid(c, a.Sid)
		if err != nil {
			log.Error("s.dao.RawSubjectRulesBySid(%d) error(%+v)", a.Sid, err)
			return
		}
	}
	return
}

func (s *Service) ActLikeCount(c context.Context, sid int64) (total int64, err error) {
	subject, err := s.dao.ActSubject(c, sid)
	if err != nil {
		log.Error("ActLikeCount s.dao.ActSubject sid:%d error(%v)", sid, err)
		return
	}
	if subject.ID == 0 {
		err = xecode.ActivityHasOffLine
		return
	}
	if subject.Type == like.CLOCKIN {
		var (
			taskIDs   []int64
			taskStats map[int64]int64
		)
		taskIDs, err = s.taskDao.TaskIDs(c, task.BusinessAct, sid)
		if err != nil {
			log.Error("s.taskDao.TaskIDs sid(%d) error(%v)", sid, err)
			return
		}
		taskStats, err = s.taskDao.TaskStats(c, taskIDs, sid, task.BusinessAct)
		if err != nil {
			log.Error("s.taskDao.TaskIDs sid(%d) error(%v)", sid, err)
			return
		}
		for _, v := range taskStats {
			total += v
		}
	} else {
		total, err = s.dao.LikeCount(c, sid, 0)
	}
	if err != nil {
		log.Error("s.dao.EsTotal or LikeCount (%d) error(%v)", sid, err)
	}
	return
}

// Protocols .
func (s *Service) Protocols(c context.Context, sids []int64) (*like.ProtocolReply, error) {
	infos, err := s.dao.ActSubjectProtocols(c, sids)
	if err != nil {
		log.Error("s.dao.ActSubjectProtocol(%d) error(%+v)", sids, err)
		return nil, err
	}
	rly := &like.ProtocolReply{}
	rly.List = make(map[int64]*like.PubicProto)
	for _, v := range infos {
		if v == nil {
			continue
		}
		rly.List[v.Sid] = &like.PubicProto{
			Tags:     v.Tags,
			Types:    v.Types,
			Sid:      v.Sid,
			BgmID:    v.BgmID,
			PasterID: v.PasterID,
			InstepID: v.InstepID,
			Oids:     v.Oids,
			Award:    v.Award,
			AwardURL: v.AwardURL,
		}
	}
	return rly, nil
}

// 定时将数据库中热门数据IDs集合写入cache中 目前只写了部分预约活动ids
func (s *Service) InternalSyncActSubjectInfoDB2Cache(ctx context.Context, req *api.InternalSyncActSubjectInfoDB2CacheReq) (reply *api.InternalSyncActSubjectInfoDB2CacheReply, err error) {
	var (
		b   = []byte("")
		IDs []int64
	)
	reply = new(api.InternalSyncActSubjectInfoDB2CacheReply)
	IDs, err = s.dao.HotGetActSubjectInfo(ctx, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActSubjectInfoDB2Cache->HotGetActSubjectInfo Source(%v) Err(%v)", req.From, err)
		return
	}
	if len(IDs) > 0 {
		b, err = json.Marshal(IDs)
	}
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActSubjectInfoDB2Cache->json.Marshal Source(%v) Err(%v)", req.From, err)
		return
	}
	err = s.dao.HotAddActSubjectInfoSet(ctx, string(b))
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActSubjectInfoDB2Cache->HotAddActSubjectInfoSet Source(%v) Err(%v)", req.From, err)
		return
	}
	log.Infoc(ctx, "[HOT-DATA-SUCC]InternalSyncActSubjectInfoDB2Cache Res(%v)", string(b))
	return
}

// 定时将有效的预约类型活动数据加载到内存中
func (s *Service) InternalGetActSubjectInfoFromCacheSetInfoMemory() {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			res, err := s.DeserializeActSubjectInfoFromCache(context.Background())
			if err != nil {
				log.Error("[HOT-DATA-FAIL]InternalGetActSubjectInfoFromCacheSetInfoMemory->DeserializeActSubjectInfoFromCache Err(%v)", err)
				continue
			}
			s.HotActSubjectInfoStore = res
		}
	}
}

// 定时将redis中有效id集合拿出来进行反序列化 在通过循环，将所有id信息拿出来进行返回
func (s *Service) DeserializeActSubjectInfoFromCache(ctx context.Context) (map[int64]*like.SubjectItem, error) {
	res := make(map[int64]*like.SubjectItem, 0)

	str, err := s.dao.HotGetActSubjectInfoSet(ctx)
	if err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActSubjectInfoFromCache->HotGetActSubjectInfoSet Err(%v)", err)
		return res, err
	}

	if str == "" {
		return res, nil
	}

	var IDs []int64
	if err := json.Unmarshal([]byte(str), &IDs); err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActSubjectInfoFromCache->json.Unmarshal([]byte(str), &IDs) Err(%v)", err)
		return res, err
	}

	// 循环获取所有id基本信息
	for _, id := range IDs {
		item, err := s.dao.ActSubject(ctx, id)
		if err != nil {
			err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActSubjectInfoFromCache->s.dao.ActSubject Err(%v) ID(%v)", err, id)
			return res, err
		}
		res[item.ID] = item
	}

	return res, nil
}

// 服务首次加载 将有效的id信息写入缓存再反序列化加载到内存中 失败了需要panic
func (s *Service) initInternalGetActSubjectInfoFromCacheSetInfoMemory() error {
	ctx := context.Background()
	var (
		err error
		res map[int64]*like.SubjectItem
	)
	// 初始化需要自己查询DB写入缓存 否则首次上线因为缓存没有数据会panic
	_, err = s.InternalSyncActSubjectInfoDB2Cache(ctx, &api.InternalSyncActSubjectInfoDB2CacheReq{
		From: "init",
	})
	if err != nil {
		return err
	}
	// 首次加载 先初始化 避免nil panic
	s.HotActSubjectInfoStore = make(map[int64]*like.SubjectItem, 0)
	// 内存反序列化数据
	res, err = s.DeserializeActSubjectInfoFromCache(ctx)
	if err != nil {
		return err
	}
	s.HotActSubjectInfoStore = res
	return nil
}

// 编辑更新admin会进行强刷单条 cache key 失败会阻止提交表单 缓存过期时间为 凌晨3-5点后的随机时间
func (s *Service) InternalGetActSubjectItemFromDB2FlushIntoCache(ctx context.Context, id int64) error {
	item, err := s.dao.RawActSubject(ctx, id)
	if err != nil {
		return fmt.Errorf("InternalGetActSubjectInfoFromDB2FlushIntoCache->s.dao.RawActSubject ID(%v) Err(%v)", id, err)
	}
	if item.ID <= 0 {
		return fmt.Errorf("InternalGetActSubjectInfoFromDB2FlushIntoCache->s.dao.RawActSubject No Result ID(%v)", item.ID)
	}
	for i := 0; i < 3; i++ {
		err = s.dao.AddCacheGetActSubjectInfo(ctx, id, item)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("InternalGetActSubjectItemFromDB2FlushIntoCache->AddCacheGetActSubjectInfo ID(%v) Err(%v)", id, err)
	}
	return nil
}

// 获取活动基本信息（性能优化版）内存找不到会回源 目前内存只包含部分热门预约活动ids
func (s *Service) GetActSubjectInfoByOptimization(ctx context.Context, id int64) (res *like.SubjectItem, err error) {
	res = new(like.SubjectItem)
	if v, ok := s.HotActSubjectInfoStore[id]; ok {
		res = v
		return
	}
	var subject *like.SubjectItem
	subject, err = s.dao.ActSubjectWithState(ctx, id)
	if err != nil {
		log.Errorc(ctx, "GetActSubjectInfoByOptimization->s.dao.ActSubject Err(%v)", err)
		return
	}
	if subject != nil {
		res = subject
	}
	return
}

// 获取活动基本信息（性能优化版）内存找不到会回源 目前内存只包含部分热门预约活动ids
func (s *Service) GetActSubjectsInfoByOptimization(ctx context.Context, ids []int64) (res map[int64]*like.SubjectItem, err error) {
	res = make(map[int64]*like.SubjectItem)
	var noMemIDs []int64

	if len(ids) > 0 {
		for _, id := range ids {
			if v, ok := s.HotActSubjectInfoStore[id]; ok {
				res[id] = v
			} else {
				noMemIDs = append(noMemIDs, id)
			}
		}
		if len(noMemIDs) > 0 {
			var subjects map[int64]*like.SubjectItem
			subjects, err = s.dao.ActSubjectsWithState(ctx, ids)
			if err != nil {
				log.Errorc(ctx, "GetActSubjectsInfoByOptimization->s.dao.ActSubjects Err(%v)", err)
				return
			}
			if subjects != nil {
				for sid, subject := range subjects {
					res[sid] = subject
				}
			}
		}
	}
	return
}

// 定时将数据库中预约数据IDs集合写入cache中
func (s *Service) InternalSyncActSubjectReserveIDsInfoDB2Cache(ctx context.Context, req *api.InternalSyncActSubjectReserveIDsInfoDB2CacheReq) (reply *api.InternalSyncActSubjectReserveIDsInfoDB2CacheReply, err error) {
	var (
		b       = []byte("")
		IDs     []int64
		nowTime string
	)
	reply = new(api.InternalSyncActSubjectReserveIDsInfoDB2CacheReply)
	nowTime = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	IDs, err = s.dao.HotGetActSubjectReserveIDsInfo(ctx, nowTime)
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActSubjectReserveIDsInfoDB2Cache->HotGetActSubjectReserveIDsInfo Source(%v) Err(%v)", req.From, err)
		return
	}
	if len(IDs) > 0 {
		b, err = json.Marshal(IDs)
	}
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActSubjectReserveIDsInfoDB2Cache->json.Marshal Source(%v) Err(%v)", req.From, err)
		return
	}
	err = s.dao.HotAddActSubjectReserveIDsInfoSet(ctx, string(b))
	if err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]InternalSyncActSubjectReserveIDsInfoDB2Cache->HotAddActSubjectReserveIDsInfoSet Source(%v) Err(%v)", req.From, err)
		return
	}
	log.Infoc(ctx, "[HOT-DATA-SUCC]InternalSyncActSubjectReserveIDsInfoDB2Cache Res(%v)", string(b))
	return
}
func (s *Service) InternalSyncActSubjectRuleIntoMemory() {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	doAction := func() {
		sids := make([]int64, 0, len(s.HotActSubjectInfoStore))
		for id, subject := range s.HotActSubjectInfoStore {
			if subject.Type == like.CLOCKIN || subject.Type == like.USERACTIONSTAT {
				sids = append(sids, id)
			}
		}
		if len(sids) == 0 {
			return
		}
		err := s.dao.AddMemorySubjectRulesBySid(context.Background(), sids)
		if err != nil {
			log.Error("[HOT-DATA-FAIL]InternalSyncActSubjectRuleIntoMemory->AddMemorySubjectRulesBySid Err(%v)", err)
		}
	}

	doAction()
	for {
		select {
		case <-ticker.C:
			doAction()
		}
	}
}

// 定时将有效的预约类型活动数据加载到内存中
func (s *Service) InternalGetActSubjectReserveIDsInfoFromCacheSetInfoMemory() {

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			res, err := s.DeserializeActSubjectReserveIDsInfoFromCache(context.Background())
			if err != nil {
				log.Error("[HOT-DATA-FAIL]InternalGetActSubjectReserveIDsInfoFromCacheSetInfoMemory->DeserializeActSubjectReserveIDsInfoFromCache Err(%v)", err)
				continue
			}
			s.HotActSubjectReserveIDsInfoStore = res
		}
	}
}

// 定时将redis中有效id集合拿出来进行反序列化 在通过循环，将所有id信息拿出来进行返回
func (s *Service) DeserializeActSubjectReserveIDsInfoFromCache(ctx context.Context) (map[int64]int64, error) {
	res := make(map[int64]int64)

	str, err := s.dao.HotGetActSubjectReserveIDsInfoSet(ctx)
	if err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActSubjectReserveIDsInfoFromCache->HotGetActSubjectReserveIDsInfoSet Err(%v)", err)
		return res, err
	}

	if str == "" {
		return res, nil
	}

	var IDs []int64
	if err := json.Unmarshal([]byte(str), &IDs); err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActSubjectReserveIDsInfoFromCache->json.Unmarshal([]byte(str), &IDs) Err(%v)", err)
		return res, err
	}

	// 获取所有id基本信息
	item := make(map[int64]int64)
	item, err = s.dao.ReservesTotal(ctx, IDs)
	if err != nil {
		err = fmt.Errorf("[HOT-DATA-FAIL]DeserializeActSubjectReserveIDsInfoFromCache->s.dao.ReservesTotal Err(%v) ID(%v)", err, IDs)
		return res, err
	}

	for sid, total := range item {
		res[sid] = total
	}

	return res, nil
}

// 服务首次加载 将有效的id信息写入缓存再反序列化加载到内存中 失败了需要panic
func (s *Service) initInternalGetActSubjectReserveIDsInfoFromCacheSetInfoMemory() error {
	ctx := context.Background()
	var (
		err error
		res map[int64]int64
	)
	// 初始化需要自己查询DB写入缓存 否则首次上线因为缓存没有数据会panic
	_, err = s.InternalSyncActSubjectReserveIDsInfoDB2Cache(ctx, &api.InternalSyncActSubjectReserveIDsInfoDB2CacheReq{
		From: "init",
	})
	if err != nil {
		return err
	}
	// 首次加载 先初始化 避免nil panic
	s.HotActSubjectReserveIDsInfoStore = make(map[int64]int64)
	// 内存反序列化数据
	res, err = s.DeserializeActSubjectReserveIDsInfoFromCache(ctx)
	if err != nil {
		return err
	}
	s.HotActSubjectReserveIDsInfoStore = res
	return nil
}

// 获取预约人数信息（性能优化版）内存找不到会回源 目前内存只包含部分热门预约活动ids
func (s *Service) GetActSubjectReserveIDInfoByOptimization(ctx context.Context, id int64) (res int64, err error) {
	res = int64(0)
	if v, ok := s.HotActSubjectReserveIDsInfoStore[id]; ok {
		res = v
		return
	}

	var ids []int64
	ids = append(ids, id)

	var totals map[int64]int64
	totals, err = s.dao.ReservesTotal(ctx, ids)
	if err != nil {
		log.Errorc(ctx, "GetActSubjectReserveIDInfoByOptimization->s.dao.ReservesTotal Err(%v)", err)
		return
	}
	if v, ok := totals[id]; ok {
		res = v
	}
	return
}

// 获取预约人数信息（性能优化版）内存找不到会回源 目前内存只包含部分热门预约活动ids
func (s *Service) GetActSubjectsReserveIDsFollowTotalByOptimization(ctx context.Context, ids []int64) (res map[int64]int64, err error) {
	res = make(map[int64]int64)
	var noMemIDs []int64

	if len(ids) > 0 {
		for _, id := range ids {
			if v, ok := s.HotActSubjectReserveIDsInfoStore[id]; ok {
				res[id] = v
			} else {
				noMemIDs = append(noMemIDs, id)
			}
		}
		if len(noMemIDs) > 0 {
			var totals map[int64]int64
			totals, err = s.dao.ReservesTotal(ctx, ids)
			if err != nil {
				err = errors.Wrap(err, "s.dao.ReservesTotal err")
				return
			}
			for _, v := range noMemIDs {
				if num, ok := totals[v]; ok {
					res[v] = num
				} else {
					res[v] = 0
				}
			}
		}
	}
	return
}

func (s *Service) initGetActReservedMapVideoSourceTags() error {
	res, err := s.GetActReservedMapVideoSourceTagsData(context.Background())
	if err != nil {
		return err
	}
	s.reserveVideoSourceTags = res
	return nil
}

func (s *Service) GetActReservedMapVideoSourceTags(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			res, err := s.GetActReservedMapVideoSourceTagsData(ctx)
			if err != nil {
				log.Error("GetActReservedMapVideoSourceTags->GetActReservedMapVideoSourceTagsData Err(%v)", err)
				continue
			}
			s.reserveVideoSourceTags = res
		}
	}
}

func (s *Service) GetActReservedMapVideoSourceTagsData(ctx context.Context) (res map[int64]*like.ActVideoSourceRelationReserve, err error) {
	res = make(map[int64]*like.ActVideoSourceRelationReserve)
	// 找出未开始或进行中的视频数据源活动
	ts := time.Now().Unix()
	subjectData, err := s.dao.RawSubjectsBeforeOrOnGoing(ctx, []int64{like.VIDEO2, like.VIDEO, like.VIDEOLIKE, like.PHONEVIDEO, like.SMALLVIDEO}, ts)
	if err != nil {
		return
	}
	// 找出视频数据源关联预约数据源的活动
	relationData := make(map[int64]*like.SubjectItem)
	collectionVideoSourceSubjectID := make([]int64, 0)
	for _, item := range subjectData {
		if item.RelationID != 0 {
			relationData[item.RelationID] = item
			collectionVideoSourceSubjectID = append(collectionVideoSourceSubjectID, item.ID)
		}
	}
	if len(collectionVideoSourceSubjectID) == 0 {
		return
	}
	// 获取关联预约数据源 的视频数据源分区和tag信息
	protocolData, err := s.dao.RawActSubjectProtocols(ctx, collectionVideoSourceSubjectID)
	if err != nil {
		return
	}
	if len(relationData) == 0 {
		return
	}
	// 整合数据
	for reserveID, data := range relationData {
		if protocol, ok := protocolData[data.ID]; ok {
			res[reserveID] = &like.ActVideoSourceRelationReserve{
				Sid:   data.ID,
				Stime: data.Stime.Time().Unix(),
				Etime: data.Etime.Time().Unix(),
				Types: protocol.Types,
				Tags:  protocol.Tags,
			}
		}
	}
	log.Infoc(ctx, "GetActReservedMapVideoSourceTagsData Succ data(%+v)", res)
	return
}

func (s *Service) GetActReserveTag(ctx context.Context, req *api.ActReserveTagReq) (reply *api.ActReserveTagReply, err error) {
	groupNum := 5
	reply = &api.ActReserveTagReply{
		List: make([]*api.ActReserveTagItem, 0),
	}
	source := s.reserveVideoSourceTags
	if len(source) == 0 {
		return
	}
	effectReserveIDs := make([]int64, 0)
	ts := time.Now().Unix()
	// 先过滤掉视频数据源活动未开始或已结束
	for reserveID, videoSourceData := range source {
		if ts >= videoSourceData.Stime && ts <= videoSourceData.Etime {
			effectReserveIDs = append(effectReserveIDs, reserveID)
		}
	}
	if len(effectReserveIDs) == 0 {
		return
	}
	// 预约id分组
	group := make([][]int64, 0)
	// 初始化 防止越界
	max := math.Ceil(float64(len(effectReserveIDs)) / float64(groupNum))
	for i := 0; i < int(max); i++ {
		group = append(group, make([]int64, 0))
	}
	index := 0
	// 分组并发
	for _, reserveID := range effectReserveIDs {
		group[index] = append(group[index], reserveID)
		if len(group[index]) == groupNum {
			index++
		}
	}
	// 映射关系 sid => 预约时间
	reserveTotal := like.ReservesTime{}
	for _, eachGroup := range group {
		reserveIDs := eachGroup
		eg := errgroup2.WithContext(ctx)
		for _, reserveID := range reserveIDs {
			sid := reserveID
			eg.Go(func(ctx context.Context) error {
				reserveRst, err := s.dao.ReserveOnly(ctx, sid, req.Mid)
				if err != nil {
					log.Errorc(ctx, "s.dao.ReserveOnly(%v,%d) error(%v)", sid, req.Mid, err)
					return nil
				}
				if reserveRst != nil && reserveRst.State == 1 {
					reserveTotal = append(reserveTotal, like.ReserveTime{Sid: sid, Mtime: reserveRst.Mtime.Time().Unix()})
				}
				return nil
			})
		}
		if err = eg.Wait(); err != nil {
			log.Errorc(ctx, "eg.Wait() error(%v)", err)
			return
		}
	}

	// 通过reserveTotal里面数据查询最后预约时间
	if len(reserveTotal) == 0 {
		return
	}

	// 按照预约时间做排序
	sort.Sort(reserveTotal)

	// 整合数据
	for _, v := range reserveTotal {
		if detail, ok := source[v.Sid]; ok {
			reply.List = append(reply.List, &api.ActReserveTagItem{
				Sid:   detail.Sid,
				Tag:   detail.Tags,
				Types: detail.Types,
			})
		}
	}

	return
}

func (s *Service) GetUpActReserveWhiteList(ctx context.Context) error {
	// 初始化
	s.dao.DynamicArc = make(map[int64]bool)
	s.dao.DynamicLive = make(map[int64]bool)

	// 单次请求限制
	limit := 5000
	// 获取最大id
	lastID, err := s.dao.GetUpActReserveWhiteListCount(ctx)
	if err != nil {
		return err
	}
	// 循环次数
	num := (lastID / limit) + 1
	// 取数据
	for i := 0; i < num; i++ {
		data, err := s.dao.GetUpActReserveWhiteList(ctx, int64(i*limit), int64((i+1)*limit))
		if err != nil {
			return err
		}
		if dynamicArcMids, ok := data[like.DynamicArc]; ok {
			for _, mid := range dynamicArcMids {
				s.dao.DynamicArc[mid] = true
			}
		}
		if dynamicLiveMids, ok := data[like.DynamicLive]; ok {
			for _, mid := range dynamicLiveMids {
				s.dao.DynamicLive[mid] = true
			}
		}
	}
	return nil
}

func (s *Service) GetTagConvert(ctx context.Context, tagIDs string) (res []*like.TagConvertItem, err error) {
	res = make([]*like.TagConvertItem, 0)

	IDs := strings.Split(tagIDs, ",")
	if len(IDs) <= 0 {
		return nil, ecode.Errorf(ecode.RequestErr, "empty tagIDs")
	}

	convertIDs := make([]int64, 0)
	for _, v := range IDs {
		var convertID int64
		convertID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			err = errors.Wrapf(err, "illegal convert to int64 tagID:(%v)", v)
			return
		}
		convertIDs = append(convertIDs, convertID)
	}

	var reply *tagService.TagsReply
	if len(convertIDs) > 0 {
		reply, err = client.TagClient.Tags(ctx, &tagService.TagsReq{Tids: convertIDs})
		if err != nil {
			return nil, ecode.Errorf(ecode.RequestErr, "client.TagClient.Tags err tids(%+v)", convertIDs)
		}
	}

	if reply == nil {
		return nil, ecode.Errorf(ecode.RequestErr, "client.TagClient.Tags reply nil tids(%+v)", convertIDs)
	}

	tags := reply.Tags
	for _, v := range convertIDs {
		if _, ok := tags[v]; !ok {
			return nil, ecode.Errorf(ecode.RequestErr, "can`t find tag info tid(%+v)", v)
		}
		if tags[v].Name == "" {
			return nil, ecode.Errorf(ecode.RequestErr, "tag name rmpty tid(%+v)", v)
		}
		res = append(res, &like.TagConvertItem{
			TagID:   v,
			TagName: tags[v].Name,
		})
	}

	return
}
