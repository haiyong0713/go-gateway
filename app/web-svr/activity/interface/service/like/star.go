package like

import (
	"context"

	relaapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"

	"go-common/library/sync/errgroup.v2"
)

const (
	_maxArcCount = 3
	_maxArcStat  = 2000
)

// StarProjectState .
func (s *Service) StarProjectState(c context.Context, mid, sid int64) (data *likemdl.Star, err error) {
	cfg, ok := s.starConf[sid]
	if !ok || cfg == nil {
		cfg = s.c.Star
	}
	var (
		likeActs map[int64]int
		stat     *likemdl.SubjectStat
	)
	if likeActs, err = s.dao.LikeActs(c, cfg.JoinSid, mid, []int64{cfg.JoinLid, cfg.DownLid}); err != nil {
		log.Error("StarProjectState s.dao.LikeActs(subID:%d,actID:%d,mid:%d) error(%+v)", cfg.Sid, cfg.JoinLid, mid, err)
		return
	}
	data = new(likemdl.Star)
	// check if join
	if isJoin, ok := likeActs[cfg.JoinLid]; !ok || isJoin != like.HasLike {
		return
	}
	data.JoinState = 1
	if downCheck, ok := likeActs[cfg.DownLid]; !ok || downCheck != like.HasLike {
		return
	}
	data.DownState = 1
	// get arc count and state
	if stat, err = s.dao.MyListTotalStateFromEs(c, cfg.Sid, mid, 1); err != nil {
		log.Error("StarProjectState s.dao.MyListTotalStateFromEs(%d,%d) error(%v)", cfg.Sid, mid, err)
		err = nil
	} else {
		data.ArchiveCount = stat.Count
		if data.ArchiveCount > _maxArcCount {
			data.ArchiveCount = _maxArcCount
		}
		data.ArchiveStat = stat.Like + stat.Coin
		if data.ArchiveStat > _maxArcStat {
			data.ArchiveStat = _maxArcStat
		}
	}
	return
}

// StarOneArc .
func (s *Service) StarOneArc(c context.Context, mid, sid int64) (arc *api.Arc, err error) {
	cfg, ok := s.starConf[sid]
	if !ok || cfg == nil {
		cfg = s.c.Star
	}
	var (
		res   *likemdl.ListInfo
		reply *api.ArcReply
	)
	if res, err = s.dao.MyListFromEs(c, cfg.Sid, mid, "id", 1, 1, 1); err != nil {
		log.Error("s.dao.MyListFromEs(%d,%d) error(%v)", cfg.Sid, mid, err)
		return
	}
	if res == nil || len(res.List) == 0 {
		return
	}
	aid := res.List[0].Wid
	if aid > 0 {
		if reply, err = client.ArchiveClient.Arc(c, &api.ArcRequest{Aid: aid}); err != nil {
			log.Error("StarOneArc s.arcClient.Arc aid(%d) error(%v)", aid, err)
			err = nil
		} else {
			arc = reply.Arc
		}
	}
	return
}

// StarMoreArc .
func (s *Service) StarMoreArc(c context.Context, mid int64) (arcs []*api.Arc, err error) {
	var (
		res   *likemdl.ListInfo
		reply *api.ArcsReply
		aids  = make([]int64, 0)
	)
	sids := []int64{s.c.Scholarship.Sid}
	sids = append(sids, s.c.Scholarship.OtherSid...)
	if res, err = s.dao.AllListFromEs(c, sids, mid, "ctime", 5, 1, 0); err != nil {
		log.Error("s.dao.MyListFromEs(%d,%d) error(%v)", s.c.Scholarship.Sid, mid, err)
		return
	}
	if res == nil || len(res.List) == 0 {
		return
	}
	for _, val := range res.List {
		aid := val.Wid
		if aid > 0 {
			aids = append(aids, aid)
		}
	}
	if reply, err = client.ArchiveClient.Arcs(c, &api.ArcsRequest{Aids: aids}); err != nil {
		log.Error("StarMoreArc s.arcClient.Arcs aids(%v) error(%v)", aids, err)
		err = nil
	}
	if reply == nil {
		return
	}
	for _, v := range reply.Arcs {
		HideArcAttribute(v)
		arcs = append(arcs, v)
	}
	return
}

func (s *Service) StarSpring(c context.Context, mid int64) (*likemdl.StarSpring, error) {
	reserve, err := s.dao.ReserveOnly(c, s.c.StarSpring.ReserveSid, mid)
	if err != nil {
		log.Error("StarSpring s.dao.ReserveOnly(%d,%d) error(%+v)", s.c.StarSpring.ReserveSid, mid, err)
		return nil, err
	}
	var isFollow bool
	if reserve != nil && reserve.ID > 0 && reserve.State == 1 {
		isFollow = true
	}
	if !isFollow {
		return &likemdl.StarSpring{}, nil
	}
	group := errgroup.WithContext(c)
	res := new(likemdl.StarSpring)
	var otherCount int64
	group.Go(func(ctx context.Context) error {
		stat, err := s.dao.MyListTotalStateFromEs(ctx, s.c.StarSpring.Sid, mid, 0)
		if err != nil {
			log.Error("StarSpring MyListTotalStateFromEs sid:%d mid:%d error(%v)", s.c.StarSpring.Sid, mid, err)
			return nil
		}
		res.ArcCount = stat.Count
		otherCount += stat.Like
		otherCount += stat.Coin
		return nil
	})
	group.Go(func(ctx context.Context) error {
		fanPre, err := s.currDao.CurrencyUser(ctx, mid, s.c.StarSpring.CurrID)
		if err != nil {
			log.Error("StarSpring CurrencyUser sid:%d mid:%d error(%v)", s.c.StarSpring.CurrID, mid, err)
			return nil
		}
		stat, err := s.relClient.Stat(ctx, &relaapi.MidReq{Mid: mid})
		if err != nil {
			log.Error("StarSpring s.relClient.Stat mid:%d error(%v)", mid, err)
			return nil
		}
		if fanPre != nil && stat != nil && (stat.Follower-fanPre.Amount > 0) {
			otherCount += stat.Follower - fanPre.Amount
		}
		return nil
	})
	group.Wait()
	// 须有投稿数才计算其他数据
	if res.ArcCount > 0 {
		res.OtherCount = otherCount
	}
	return res, nil
}
