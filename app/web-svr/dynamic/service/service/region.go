package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/model"

	"go-common/library/sync/errgroup.v2"
)

const _retry = 3

// addRegion add region redis .
func (s *Service) addRegion(ctx context.Context, arc *model.ArchiveSub) {
	var (
		err  error
		reid int32
	)
	if arc == nil {
		return
	}
	defer func() {
		if err != nil {
			s.addRetryReg(ctx, arc.Aid, model.Insert, 0)
		}
	}()
	pubtime, _ := time.ParseInLocation("2006-01-02 15:04:05", arc.PubTime, time.Local)
	param := &arcmdl.RegionArc{Aid: arc.Aid, Attribute: arc.Attribute, Copyright: int8(arc.Copyright), PubDate: xtime.Time(pubtime.Unix())}
	reid = s.isExistFatherReg(arc.Typeid) // 获取分区的父id
	if err = s.dao.AddRegionArcCache(ctx, arc.Typeid, reid, param); err != nil {
		log.Error("[addRegion]s.dao.AddRegionArcCache aid(%d) error(%v)", arc.Aid, err)
	}
}

// DelArc delete region archive redis .
func (s *Service) delArc(ctx context.Context, rid int32, arc *model.ArchiveSub) {
	var (
		err  error
		reid int32
	)
	if arc == nil {
		return
	}
	defer func() {
		if err != nil {
			s.addRetryReg(ctx, arc.Aid, model.Delete, rid)
		}
	}()
	pubtime, _ := time.ParseInLocation("2006-01-02 15:04:05", arc.PubTime, time.Local)
	param := &arcmdl.RegionArc{Aid: arc.Aid, Attribute: arc.Attribute, Copyright: int8(arc.Copyright), PubDate: xtime.Time(pubtime.Unix())}
	reid = s.isExistFatherReg(rid)
	if err = s.dao.DelArcCache(ctx, rid, reid, param); err != nil {
		log.Error("s.DelArcCache rid(%d) aid(%d) error(%v)", rid, param.Aid, err)
	}
}

// RegAllArc get all archive .
func (s *Service) RegAllArc(ctx context.Context, rid int64, ps, pn int) (arc []*arcmdl.Arc, count int64, err error) {
	if arc, count, err = s.getArc(ctx, ps, pn, fmt.Sprintf("%d_a", rid)); err != nil {
		log.Error("[RegAllArc] s.getArc() rid(%d) error(%v)", rid, err)
		return
	}
	if len(arc) == 0 {
		arc = make([]*arcmdl.Arc, 0)
		count = 0
	}
	return
}

// RegOriginArc get original archive .
func (s *Service) RegOriginArc(ctx context.Context, rid int64, ps, pn int) (arc []*arcmdl.Arc, count int64, err error) {
	if arc, count, err = s.getArc(ctx, ps, pn, fmt.Sprintf("%d_o", rid)); err != nil {
		log.Error("[RegOriginArc] s.getArc() rid(%d) error(%v)", rid, err)
		return
	}
	if len(arc) == 0 {
		arc = make([]*arcmdl.Arc, 0)
		count = 0
	}
	return
}

func (s *Service) getArc(ctx context.Context, ps, pn int, key string) (arc []*arcmdl.Arc, count int64, err error) {
	gp := errgroup.WithCancel(ctx)
	gp.Go(func(ctx context.Context) (err error) {
		var (
			start = (pn - 1) * ps
			end   = start + ps - 1
		)
		if arc, err = s.dealArc(ctx, start, end, key); err != nil {
			log.Error("s.dealArc() error(%v)", err)
		}
		return err
	})
	gp.Go(func(ctx context.Context) (err error) {
		if count, err = s.dao.RegCount(ctx, key); err != nil {
			log.Error("s.dao.GetRegCount() count(%d) key(%s) error(%v)", count, key, err)
		}
		return err
	})
	if err = gp.Wait(); err != nil {
		log.Error("gp.Wait() error(%v)", err)
	}
	return
}

// get need all aid .
func (s *Service) arcAids(ctx context.Context, key string, start, end int) (aids []int64, err error) {
	var (
		keyRes     []*model.ResKey
		res        []*model.AllRegKey
		firstStart bool
	)
	// get count from redis
	if res, err = s.dao.RegionKeyCount(ctx, key); err != nil {
		log.Error("s.dao.GetRegionCount key(%s) error(%v)", key, err)
		return
	}
	if len(res) == 0 {
		return
	}
	for _, v := range res {
		if start >= 0 {
			start = start - int(v.Count)
		}
		if end >= 0 {
			end = end - int(v.Count)
		}
		if start < 0 && !firstStart {
			flag := &model.ResKey{}
			flag.Reskey = v.Key
			flag.Start = start + int(v.Count)
			flag.End = int(v.Count)
			keyRes = append(keyRes, flag)
			firstStart = true
		} else if start < 0 && end > 0 {
			flag := &model.ResKey{}
			flag.Reskey = v.Key
			flag.Start = 0
			flag.End = int(v.Count)
			keyRes = append(keyRes, flag)
		}
		if end < 0 {
			flag := &model.ResKey{}
			flag.Reskey = v.Key
			flag.Start = 0
			flag.End = end + int(v.Count)
			keyRes = append(keyRes, flag)
			break
		}
	}
	if len(keyRes) == 2 && keyRes[0].Reskey == keyRes[1].Reskey {
		tmp := &model.ResKey{
			Reskey: keyRes[0].Reskey,
			Start:  keyRes[0].Start,
			End:    keyRes[1].End,
		}
		keyRes = make([]*model.ResKey, 0, 1)
		keyRes = append(keyRes, tmp)
	}
	if aids, err = s.dao.AllRegion(ctx, keyRes); err != nil {
		log.Error("s.dao.GetAllRegion aids(%s) error(%v)", xstr.JoinInts(aids), err)
	}
	return
}

// fail retry .
func (s *Service) retryReg() {
	defer s.waiter.Done()
	for {
		if s.closeRetry {
			return
		}
		var (
			err error
			bt  []byte
			arc *arcmdl.ArcReply
			ctx = context.Background()
		)
		if bt, err = s.dao.PopFail(ctx); err != nil || bt == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		msg := &model.ActAid{}
		if err = json.Unmarshal(bt, msg); err != nil {
			log.Error("retryReg Unmarshal(%s) error(%v)", bt, err)
			continue
		}
		log.Info("[retryReg] key(%s) data(%v)", model.FailList, bt)
		if arc, err = s.arcClient.Arc(ctx, &arcmdl.ArcRequest{Aid: msg.Aid}); err != nil || arc == nil {
			log.Error("[retryReg] s.arcClient.Arc aid(%d) error(%v)", msg.Aid, err)
			s.addRetryReg(ctx, msg.Aid, msg.Action, msg.TypeID)
			continue
		}
		if arc.Arc == nil {
			continue
		}
		pubtime := time.Unix(int64(arc.Arc.PubDate), 0).Format("2006-01-02 15:04:05")
		param := &model.ArchiveSub{Aid: arc.Arc.Aid, Attribute: arc.Arc.Attribute, Copyright: int8(arc.Arc.Copyright), PubTime: pubtime, Typeid: arc.Arc.TypeID}
		switch msg.Action {
		case model.Insert:
			if param.CanPlay() {
				s.addRegion(context.Background(), param)
			}
		case model.Delete:
			s.delArc(context.Background(), msg.TypeID, param)
		}
	}
}

// RegionArcInit init region archive redis .
func (s *Service) RegionArcInit(ctx context.Context, rid int32) (err error) {
	if s.arcInit {
		err = errors.New("已经在初始化了")
		return
	}
	defer func() {
		s.arcInit = false
	}()
	s.arcInit = true

	var (
		startFlag = s.c.Rule.InitRegStart
		length    = s.c.Rule.AddArcNum
		endFlag   = s.c.Rule.InitRegEnd
		end       = startFlag + length
	)
	if rid > 0 {
		for start := startFlag; start <= endFlag; start += length {
			var resDyArc map[int32][]*arcmdl.RegionArc
			if resDyArc, err = s.dao.ArchiveAll(context.Background(), rid, start, end); err != nil {
				log.Error("[RegionArcInit] s.dao.ArchiveAll() rid(%d) error(%v)", rid, err)
				return
			}
			if arc, ok := resDyArc[rid]; ok {
				if err = s.dao.AddRegionArcCache(context.Background(), rid, 0, arc...); err != nil {
					log.Error("[RegionArcInit] s.dao.AddRegionArcCache() rid(%d) error(%v)", rid, err)
					return
				}
			}
			end += length
			log.Info("[running] start(%d) rid(%d)", start, rid)
			time.Sleep(time.Duration(s.c.Rule.InitArc))
		}
	} else {
		for start := startFlag; start <= endFlag; start += length {
			var resDyArc map[int32][]*arcmdl.RegionArc
			if resDyArc, err = s.dao.ArchiveAll(context.Background(), rid, start, end); err != nil {
				log.Error("[RegionArcInit] s.dao.ArchiveAll() rid(%d) error(%v)", rid, err)
				return
			}
			for typeid, arc := range resDyArc {
				if err = s.dao.AddRegionArcCache(context.Background(), typeid, 0, arc...); err != nil {
					log.Error("[RegionArcInit] s.dao.AddRegionArcCache() rid(%d) error(%v)", typeid, err)
					return
				}
			}
			end += length
			log.Info("[running] start(%d)", start)
			time.Sleep(time.Duration(s.c.Rule.InitArc))
		}
	}
	log.Info("[running] init region success !")
	return
}

// RegionCnt get today region count.
func (s *Service) RegionCnt(ctx context.Context, rids []int32) (res map[int32]int64, err error) {
	var (
		t   = time.Now()
		min = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
		max = t.Unix()
	)
	if res, err = s.dao.RegionCnt(ctx, rids, min, max); err != nil {
		log.Error("[SecondRegCnt] s.dao.SeRegCount() error(%v)", err)
	}
	return
}

// RecentThrdRegArc 最近三天分区稿件
func (s *Service) RecentThrdRegArc(ctx context.Context, rid int32, pn, ps int) (res []*arcmdl.Arc, err error) {
	var (
		start = (pn - 1) * ps
		end   = start + ps - 1
		aids  []int64
	)
	if aids, err = s.dao.RecentRegArc(ctx, rid, start, end); err != nil {
		log.Error("[RecentRegionArc] s.dao.RecentRegArc() rid(%d) pn(%d) ps(%d) error(%v)", rid, pn, ps, err)
		return
	}
	if res, err = s.archives(ctx, aids); err != nil {
		log.Error("[RecentThrdRegArc] s.archives() aids(%d) error(%v)", aids, err)
		return
	}
	if res == nil {
		res = make([]*arcmdl.Arc, 0)
	}
	return
}

// RecentWeeklyArc 最近七天区稿件
func (s *Service) RecentWeeklyArc(ctx context.Context, pn, ps int) (res []*arcmdl.Arc, count int64, err error) {
	var (
		start = (pn - 1) * ps
		end   = start + ps - 1
	)
	gp := errgroup.WithContext(ctx)
	gp.Go(func(ctx context.Context) (err error) {
		if count, err = s.dao.RecentAllRegArcCnt(ctx); err != nil {
			log.Error("[RecentWeeklyArc] s.dao.RecentAllRegArcCnt pn(%d) ps(%d) error(%v)", pn, ps, err)
		}
		return
	})
	gp.Go(func(ctx context.Context) (err error) {
		var aids []int64
		if aids, err = s.dao.RecentRegArc(ctx, 0, start, end); err != nil {
			log.Error("[RecentWeeklyArc] s.dao.RecentRegionArc() error(%v)", err)
			return
		}
		if res, err = s.archives(ctx, aids); err != nil {
			log.Error("[RecentWeeklyArc] s.archives() aids(%d) error(%v)", aids, err)
		}
		return
	})
	if err = gp.Wait(); err != nil {
		return
	}
	if res == nil {
		res = make([]*arcmdl.Arc, 0)
	}
	return
}

func (s *Service) archives(ctx context.Context, aids []int64) (arcs []*arcmdl.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	var arcsReply *arcmdl.ArcsReply
	if arcsReply, err = s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("[archives]s.arcClient.Arcs() aid(%v) error(%v)", aids, err)
		return
	}
	if len(arcsReply.Arcs) > 0 {
		for _, v := range aids {
			if arc, ok := arcsReply.Arcs[v]; ok && arc.IsNormal() {
				arcs = append(arcs, arc)
			}
		}
	}
	return
}

func (s *Service) addRetryReg(ctx context.Context, aid int64, action string, typeID int32) {
	var (
		err error
		msg = &model.ActAid{
			Aid:    aid,
			Action: action,
			TypeID: typeID,
		}
	)
	log.Warn("[addRetryReg] aid(%d) action(%s) TypeID(%d)", msg.Aid, msg.Action, msg.TypeID)
	if err = s.dao.PushFail(ctx, msg); err != nil {
		log.Error("[addRetryReg] addRetryReg aid(%d) action(%s) TypeID(%d) error(%v)", msg.Aid, msg.Action, msg.TypeID, err)
		_ = s.dao.Send(ctx, fmt.Sprintf("dynamic_service 二级分区重试添加失败 aid(%d) action(%s) TypeID(%d)", msg.Aid, msg.Action, msg.TypeID))
	}
}

func (s *Service) isExistFatherReg(rid int32) (fid int32) {
	if types, ok := s.typesMap[rid]; ok && types.Pid != 0 {
		fid = types.Pid
	}
	return
}

func (s *Service) dealArc(ctx context.Context, start, end int, key string) (arc []*arcmdl.Arc, err error) {
	var (
		aids []int64
		res  map[int64]*arcmdl.Arc
	)
	if aids, err = s.arcAids(ctx, key, start, end); err != nil {
		log.Error("s.arcAids() key(%s) error(%v)", key, err)
		return
	}
	if len(aids) == 0 {
		log.Warn("aids is 0 key(%s)", key)
		return
	}
	if res, err = s.archives3(ctx, aids); err != nil {
		log.Error("s.archives3 aids(%s) error(%v)", xstr.JoinInts(aids), err)
		return
	}
	for _, aid := range aids {
		if a, ok := res[aid]; ok && a.IsNormal() {
			arc = append(arc, a)
		} else {
			log.Warn("[abnormalAID]dynamic-service分区中存在不正常稿件(aid:%d) rid(%d) key(%s)", a.Aid, a.TypeID, key)
		}
	}
	return
}
