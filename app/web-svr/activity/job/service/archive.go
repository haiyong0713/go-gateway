package service

import (
	"context"
	"sort"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/job/model/like"
)

const (
	_rankViewPieceSize = 100
	_rankCount         = 50
	_clickOrder        = "click"
)

func (s *Service) subsRankproc() {
	if s.closed {
		return
	}
	var (
		subs []*like.ActSubject
		err  error
	)
	now := time.Now()
	if subs, err = s.dao.SubjectList(context.Background(), []int64{likemdl.PHONEVIDEO, likemdl.SMALLVIDEO, likemdl.VIDEO, likemdl.VIDEOLIKE, likemdl.VIDEO2}, now); err != nil {
		log.Error("viewRankproc s.dao.SubjectList error(%+v)", err)
		return
	}
	if len(subs) == 0 {
		log.Warn("viewRankproc no subjects time(%d)", now.Unix())
		return
	}
	for _, v := range subs {
		if v != nil && v.ID > 0 {
			s.viewRankproc(v.ID)
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Info("subsRankproc success()")
}

func (s *Service) viewRankproc(sid int64) {
	var (
		likeCnt  int
		rankArcs []*api.Arc
		list     []*like.EsItem
		err      error
	)
	if likeCnt, err = s.dao.LikeCnt(context.Background(), sid); err != nil {
		log.Error("viewRankproc s.dao.LikeCnt(sid:%d) error(%v)", sid, err)
		return
	}
	if likeCnt == 0 {
		log.Warn("viewRankproc s.dao.LikeCnt(sid:%d) likeCnt == 0", sid)
		return
	}

	if list, err = s.dao.ListFromEs(context.Background(), &like.EsParams{Sid: sid, State: 1, Order: _clickOrder, Sort: "desc", Ps: _rankViewPieceSize, Pn: 1}); err != nil {
		log.Error("viewRankproc s.dao.ListFromEs(%d,%d) error(%+v)", sid, 1, err)
		return
	}
	var aids []int64
	for _, v := range list {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	log.Info("viewRankproc sid:%d aids(%v)", sid, aids)
	if len(aids) == 0 {
		return
	}
	var arcs map[int64]*api.Arc
	if arcs, err = s.arcs(context.Background(), aids, _retryTimes); err != nil {
		log.Error("viewRankproc s.arcs(%v) error(%v)", aids, err)
		return
	}
	for _, aid := range aids {
		if arc, ok := arcs[aid]; ok && arc.IsNormal() {
			rankArcs = append(rankArcs, arc)
		}
	}
	sort.Slice(rankArcs, func(i, j int) bool {
		return rankArcs[i].Stat.View > rankArcs[j].Stat.View
	})
	if len(rankArcs) > _rankCount {
		rankArcs = rankArcs[:_rankCount]
	}
	if len(rankArcs) > 0 {
		var rankAids []int64
		for _, v := range rankArcs {
			if v != nil && v.Aid > 0 {
				rankAids = append(rankAids, v.Aid)
			}
		}
		if err = s.setViewRank(context.Background(), sid, rankAids, "", _retryTimes); err != nil {
			log.Error("viewRankproc s.setObjectStat(%d,%v) error(%+v)", sid, rankAids, err)
		}
	}
}

func (s *Service) arcs(c context.Context, aids []int64, retryCnt int) (arcs map[int64]*api.Arc, err error) {
	var arcsRly *api.ArcsReply
	for i := 0; i < retryCnt; i++ {
		if arcsRly, err = s.arcClient.Arcs(c, &api.ArcsRequest{Aids: aids}); err == nil {
			arcs = arcsRly.Arcs
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) setViewRank(c context.Context, sid int64, aids []int64, typ string, retryTime int) (err error) {
	for i := 0; i < retryTime; i++ {
		if err = s.dao.SetViewRank(c, sid, aids, typ); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) staffPassTask(c context.Context, aid, mid int64) {
	s.singleDoTask(c, mid, s.c.Staff.PassTaskID)
	reply, err := s.arcClient.Arc(c, &api.ArcRequest{Aid: aid})
	if err != nil {
		log.Error("statThumbupproc s.arcClient.Arc(%d) error(%v)", aid, err)
		return
	}
	for _, v := range reply.Arc.StaffInfo {
		if v == nil || v.Mid == 0 || v.Mid == mid {
			continue
		}
		s.singleDoTask(c, v.Mid, s.c.Staff.PassTaskID)
	}
}
