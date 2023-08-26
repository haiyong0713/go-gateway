package like

import (
	"context"
	"encoding/json"
	"time"

	accountAPI "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	actAPI "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/model/dynamic"
	l "go-gateway/app/web-svr/activity/interface/model/like"

	"go-common/library/sync/errgroup.v2"
)

var (
	waitDone = int32(0)
	pass     = int32(1)
	unpass   = int32(2)
	normal   = 0
	auditing = 1
	ongoing  = 2
)

func (s *Service) UpLaunchCheck(c context.Context, mid int64) (res *l.UpCheck, err error) {
	up := new(l.ActUp)
	if up, err = s.dao.ActUp(c, mid); err != nil {
		log.Error("s.dao.ActUp(%d) error(%v)", mid, err)
		return
	}
	res = &l.UpCheck{Up: up}
	if up == nil {
		res.Status = normal
		return
	}
	var status int
	if up.State == waitDone {
		status = auditing
	}
	nowT := time.Now().Unix()
	if up.State == pass && ((up.Stime <= xtime.Time(nowT) && up.Etime >= xtime.Time(nowT)) || up.Stime > xtime.Time(nowT)) {
		status = ongoing
	}
	if up.Offline == 1 || up.State == unpass {
		status = normal
	}
	res.Status = status
	return
}

func (s *Service) UpCheck(c context.Context, mid int64) (res int, err error) {
	for _, v := range s.c.UpAct.WhiteMid {
		if mid == v {
			res = 1
			break
		}
	}
	return
}

func (s *Service) UpLaunch(c context.Context, title, statement string, stime, etime xtime.Time, aid, mid int64) (err error) {
	up, err := s.UpLaunchCheck(c, mid)
	if err != nil || up.Status != normal {
		err = ecode.ActivityIsINOrCheck
		return
	}
	if err = s.dao.InsertUpAct(c, title, statement, stime, etime, aid, mid); err != nil {
		log.Error("s.dao.InsertUpAct(%d) error(%v)", mid, err)
	}
	return
}

func (s *Service) UpArchiveList(c context.Context, mid int64) (res *api.UpArcsReply, err error) {
	if mid == 0 {
		return
	}
	resReply := &api.UpArcsReply{}
	resReply.Arcs = make([]*api.Arc, 0)
	if res, err = client.ArchiveClient.UpArcs(c, &api.UpArcsRequest{Mid: mid, Ps: 20, Pn: 1}); err != nil {
		log.Error("s.arcClient.UpArcs(%d) error(%v)", mid, err)
	}
	if res != nil {
		for _, v := range res.Arcs {
			HideArcAttribute(v)
			resReply.Arcs = append(resReply.Arcs, v)
		}

	}

	return resReply, err
}

func (s *Service) UpActPage(c context.Context, mid, sid int64) (res *l.ActUpPage, err error) {
	up := new(l.ActUp)
	if up, err = s.dao.ActUpBySid(c, sid); err != nil {
		log.Error("s.dao.ActUpBySid(%d) error(%v)", sid, err)
		return
	}
	if up == nil || up.Offline == 1 || up.State != pass {
		err = ecode.ActivityNotExist
		return
	}
	eg := errgroup.WithContext(c)
	var dyReply *dynamic.DyReply
	if up.Title != "" {
		eg.Go(func(ctx context.Context) (err error) {
			if dyReply, err = s.dynamicDao.FetchDynamics(ctx, 0, mid, 20, "2", up.Title, 1); err != nil {
				log.Error("s.dao.FetchDynamics(%s) error(%v)", up.Title, err)
			}
			return
		})
	}
	var arc *api.ArcReply
	if up.Aid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if arc, err = client.ArchiveClient.Arc(ctx, &api.ArcRequest{Aid: up.Aid}); err != nil {
				log.Error("s.arcClient.Arc(%d) error(%v)", up.Aid, err)
			}
			return
		})
	}
	var acc *accountAPI.InfoReply
	eg.Go(func(ctx context.Context) (err error) {
		if acc, err = s.accClient.Info3(ctx, &accountAPI.MidReq{Mid: up.Mid}); err != nil {
			log.Error("s.accClient.Info3(%d) error(%v)", up.Mid, err)
		}
		return
	})
	eg.Wait()
	act := &l.ActUpReply{ActUp: up, Image: s.c.UpAct.Image}
	if acc != nil && acc.Info != nil {
		act.Name = acc.Info.Name
		act.Face = acc.Info.Face
	}
	res = &l.ActUpPage{Act: act, Archive: arc}
	if dyReply != nil {
		res.Dynamic = dyReply.Cards
	}
	return
}

func (s *Service) UpActDo(c context.Context, sid, mid, totalTime int64, matchedPercent float32) (res int64, err error) {
	eg := errgroup.WithContext(c)
	up := new(l.ActUp)
	eg.Go(func(ctx context.Context) (err error) {
		if up, err = s.dao.ActUpBySid(ctx, sid); err != nil {
			log.Error("s.dao.ActUpBySid(%d) error(%v)", sid, err)
		}
		return
	})
	var days float64
	eg.Go(func(ctx context.Context) (err error) {
		if days, err = s.dao.CacheUpUserDays(ctx, sid, mid); err != nil {
			log.Error("s.dao.CacheUpUserDays(%d,%d) error(%v)", sid, mid, err)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res = int64(days)
	if up == nil {
		err = ecode.ActivityNotExist
		return
	}
	nowT := time.Now().Unix()
	if up.Offline == 1 || up.State == unpass || up.Etime < xtime.Time(nowT) {
		err = ecode.ActivityHasOffLine
		return
	}
	if up.Stime > xtime.Time(nowT) {
		err = ecode.ActivityNotStart
		return
	}
	attendResult := &l.AttendResult{TotalTime: totalTime, MatchedPercent: matchedPercent}
	var bs []byte
	if bs, err = json.Marshal(attendResult); err != nil {
		log.Error("UpActDo json.Marshal() error(%v)", err)
		return
	}
	result := string(bs)
	var upUserState map[string]*l.UpActUserState
	if upUserState, err = s.dao.UpActUserState(c, up, mid); err != nil {
		log.Error("UpActDo s.dao.UpActUserState mid(%d) arg(%v) error(%v)", mid, up, err)
		return
	}
	round := up.UpRound(nowT)
	userState, ok := upUserState[like.RoundMapKey(round)]
	if ok && userState != nil {
		if userState.Finish == l.HasFinish {
			return
		}
	}
	// add log
	if err = s.dao.AddUserLog(c, sid, mid, l.BusinessID, round); err != nil {
		log.Error("s.dao.AddUserLog sid(%d) mid(%d) round(%d) error(%v)", sid, mid, round, err)
		return
	}
	var count, finish int64
	if userState != nil {
		count = userState.Times + 1
		if up.FinishCount == count {
			finish = l.HasFinish
			res = res + 1
		}
		if err = s.dao.UpUserState(c, sid, mid, l.BusinessID, round, finish, count, up.Suffix); err != nil {
			log.Error("s.dao.UpUserState sid(%d) mid(%d) round(%d) count(%d) finish(%d) error(%v)", sid, mid, round, count, finish, err)
			return
		}
	} else {
		count = 1
		if up.FinishCount == count {
			res = res + 1
			finish = l.HasFinish
		}
		if err = s.dao.AddUserState(c, sid, mid, l.BusinessID, round, finish, count, up.Suffix, result); err != nil {
			log.Error("s.dao.AddUserState sid(%d) mid(%d) round(%d) count(%d) finish(%d) error(%v)", sid, mid, round, count, finish, err)
			return
		}
	}
	s.cache.Do(c, func(ctx context.Context) {
		upUser := &l.UpActUserState{
			Sid:    sid,
			Mid:    mid,
			Bid:    l.BusinessID,
			Round:  round,
			Times:  count,
			Finish: finish,
			Result: result,
			Ctime:  xtime.Time(nowT),
		}
		s.dao.AddUpUsersRank(ctx, sid, mid, nowT, res)
		s.dao.SetCacheUpActUserState(ctx, up, upUser, mid, round)
	})
	return
}

func (s *Service) UpActInfo(c context.Context, aid int64) (res *actAPI.UpActInfo, err error) {
	up := new(l.ActUp)
	if up, err = s.dao.ActUpByAid(c, aid); err != nil {
		log.Error("s.dao.UpActInfo(%d) error(%v)", aid, err)
		return
	}
	if up == nil {
		return
	}
	var acc *accountAPI.InfoReply
	if acc, err = s.accClient.Info3(c, &accountAPI.MidReq{Mid: up.Mid}); err != nil {
		log.Error("s.accClient.Info3(%d) error(%v)", up.Mid, err)
		return
	}
	var name string
	if acc != nil && acc.Info != nil {
		name = acc.Info.Name
	}
	res = &actAPI.UpActInfo{
		ID:        up.ID,
		Mid:       up.Mid,
		Title:     up.Title,
		Statement: up.Statement,
		Stime:     up.Stime,
		Etime:     up.Etime,
		Ctime:     up.Ctime,
		Mtime:     up.Mtime,
		State:     up.State,
		Offline:   up.Offline,
		Aid:       up.Aid,
		Image:     s.c.UpAct.Image,
		Name:      name,
	}
	return
}

func (s *Service) UpActRank(c context.Context, sid, mid int64) (res *l.RankList, err error) {
	list := make([]*l.RankUserDays, 0)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if list, err = s.dao.CacheUpUsersRank(ctx, sid, 0, 10); err != nil {
			log.Error("GradeShow s.dao.CacheUsersRank error(%v)", err)
		}
		return
	})
	var days float64
	eg.Go(func(ctx context.Context) (err error) {
		if days, err = s.dao.CacheUpUserDays(ctx, sid, mid); err != nil {
			log.Error("s.dao.CacheUpUserDays(%d,%d) error(%v)", sid, mid, err)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res = &l.RankList{SelfDays: int64(days)}
	mids := make([]int64, 0)
	for _, v := range list {
		if v == nil || v.Mid <= 0 {
			continue
		}
		mids = append(mids, v.Mid)
	}
	cards := make(map[int64]*accountAPI.Card)
	if cards, err = s.accCards(c, mids); err != nil {
		log.Error("UpActRank s.accCards(%v) error(%v)", mids, err)
		return
	}
	i := 1
	for _, v := range list {
		if i > 10 {
			break
		}
		if v == nil || v.Mid <= 0 {
			continue
		}
		if _, k := cards[v.Mid]; !k {
			continue
		}
		tmp := &l.RankUserInfo{
			Mid:  v.Mid,
			Days: int64(v.Days),
			Name: cards[v.Mid].Name,
			Face: cards[v.Mid].Face,
		}
		res.List = append(res.List, tmp)
		i++
	}
	return
}

// accCards .
func (s *Service) accCards(c context.Context, mids []int64) (ac map[int64]*accountAPI.Card, err error) {
	var (
		arg       = &accountAPI.MidsReq{Mids: mids}
		tempReply *accountAPI.CardsReply
	)
	if len(mids) == 0 {
		return
	}
	if tempReply, err = s.accClient.Cards3(c, arg); err != nil {
		log.Error("s.accRPC.Cards3(%d) error(%v)", mids, err)
		err = ecode.ActivityServerTimeout
		return
	}
	ac = make(map[int64]*accountAPI.Card)
	for k, v := range tempReply.Cards {
		ac[k] = v
	}
	return
}
