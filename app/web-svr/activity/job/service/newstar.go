package service

import (
	"context"
	"time"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/model/like"

	relationmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"go-common/library/sync/errgroup.v2"
)

const (
	_maxPn    = 1000
	_ps       = 100
	_maxArcPn = 10
	_original = 1
	_unDo     = 0
	_checking = 1
	_finish   = 2
)

func (s *Service) NewstarArchiveTask() {
	var (
		pn  int64
		ctx = context.Background()
	)
	for i := 0; i < _maxPn; i++ {
		vUsers, err := s.dao.BigVUsers(ctx, pn)
		if err != nil {
			log.Error("NewstarArchiveTask i(%d) error(%+v)", i, err)
			time.Sleep(time.Second)
			continue
		}
		if len(vUsers) == 0 {
			log.Warn("NewstarArchiveTask success i(%d)", i)
			break
		}
		arcStatusMap := make(map[int64]int64, 100)
		for _, vUser := range vUsers {
			arcStatus := s.arcTaskFinish(ctx, vUser)
			arcStatusMap[vUser.ID] = arcStatus
			pn = vUser.ID
			time.Sleep(10 * time.Millisecond)
		}
		if _, err = s.dao.UpdateArcStatus(ctx, arcStatusMap); err != nil {
			log.Error("NewstarArchiveTask s.dao.UpdateArcStatus arcStatusMap(%+v) error(%+v)", arcStatusMap, err)
			time.Sleep(time.Second)
			continue
		}
		for _, vUser := range vUsers {
			s.dao.DelNewstar(ctx, vUser.Mid, vUser.InviterMid, s.c.Rule.NewstarName)
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (s *Service) arcTaskFinish(ctx context.Context, vUser *like.BigVUser) int64 {
	for i := 1; i <= _maxArcPn; i++ {
		arcsRly, err := s.arcClient.UpArcs(ctx, &api.UpArcsRequest{Mid: vUser.Mid, Pn: int32(i), Ps: _ps})
		if err != nil || arcsRly == nil {
			log.Error("arcTaskFinish s.arcClient.UpArcs mid(%d) pn(%d) error(%+v)", vUser.Mid, i, err)
			if ecode.EqualError(ecode.NothingFound, err) {
				return 0
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if len(arcsRly.Arcs) == 0 {
			return 0
		}
		for _, arc := range arcsRly.Arcs {
			if arc.Copyright == _original && arc.IsNormal() {
				for _, tp := range s.c.Rule.NewstarArcTypes {
					if tp == arc.TypeID {
						return 1
					}
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return 0
}

func (s *Service) FinishNewstar() {
	var (
		pn  int64
		ctx = context.Background()
	)
	for i := 0; i < _maxPn; i++ {
		time.Sleep(time.Second)
		vUsers, err := s.dao.BigVUsers(ctx, pn)
		if err != nil {
			log.Error("FinishNewstar i(%d) error(%+v)", i, err)
			continue
		}
		if len(vUsers) == 0 {
			log.Warn("FinishNewstar success i(%d)", i)
			break
		}
		for _, vUser := range vUsers {
			pn = vUser.ID
			// 结算参加活动达到30(可配制天数)天用户
			if s.remainingDays(vUser.Ctime) > 0 {
				continue
			}
			s.finishBigV(ctx, vUser)
		}
	}
}

func (s *Service) remainingDays(ctime xtime.Time) int64 {
	startTime := time.Unix(ctime.Time().Unix(), 0)
	endTime := startTime.Add(time.Duration(s.c.Rule.NewstarDays))
	endDate := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
	subTime := endDate.Sub(time.Now()).Hours()
	rs := int64(subTime / 24)
	if rs < 0 {
		return 0
	}
	return rs
}

func (s *Service) finishBigV(c context.Context, vUser *like.BigVUser) {
	var (
		arcStatus, follower, isMobile, isName, isIdentity int64
		user                                              *accmdl.Profile
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		arcStatus = s.arcTaskFinish(c, vUser)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		account, accErr := s.accClient.Profile3(c, &accmdl.MidReq{Mid: vUser.Mid})
		if accErr != nil {
			log.Error("s.accClient.Profile3(%d) error(%v)", vUser.Mid, accErr)
			return accErr
		}
		user = account.Profile
		return nil
	})
	group.Go(func(ctx context.Context) error {
		statReply, relErr := s.relationClient.Stat(ctx, &relationmdl.MidReq{Mid: vUser.Mid})
		if relErr != nil || statReply == nil {
			log.Error("s.relationClient.Stat(%d) error(%v)", vUser.Mid, relErr)
			return relErr
		}
		follower = statReply.Follower
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("finishBigV group.Wait mid(%d) error(%+v)", vUser.Mid, err)
		return
	}
	group2 := errgroup.WithContext(c)
	group2.Go(func(ctx context.Context) error {
		if user.Identification == 0 {
			realRes, realErr := s.memberClient.RealnameApplyStatus(c, &memberAPI.MemberMidReq{Mid: vUser.Mid})
			if realErr != nil {
				log.Error("finishBigV s.memberClient.RealnameApplyStatus(%d) error(%+v)", vUser.Mid, realErr)
				isName = _unDo
				return nil
			}
			switch realRes.Status {
			case 0:
				isName = _checking
			case 1:
				isName = _finish
			default:
				isName = _unDo
			}
		} else {
			isName = _finish
		}
		return nil
	})
	group2.Go(func(ctx context.Context) error {
		if user.Official.Role == 0 {
			officialRes, officialErr := s.memberClient.OfficialDoc(c, &memberAPI.MidReq{Mid: vUser.Mid})
			if officialErr != nil {
				log.Error("finishBigV s.memberClient.OfficialDoc(%d) error(%+v)", vUser.Mid, officialErr)
				isIdentity = _unDo
				return nil
			}
			switch officialRes.State {
			case 0:
				isIdentity = _checking
			case 1:
				isIdentity = _finish
			default:
				isIdentity = _unDo
			}
		} else {
			isIdentity = _finish
		}
		return nil
	})
	if err := group2.Wait(); err != nil {
		log.Error("finishBigV group2.Wait mid(%d) error(%+v)", vUser.Mid, err)
		return
	}
	if user.TelStatus == 1 {
		isMobile = _finish
	}
	s.dao.FinishBigV(c, vUser.ID, isName, isMobile, isIdentity, follower, arcStatus)
	return

}

// IdentityChecking .
func (s *Service) IdentityChecking() {
	var (
		pn  int64
		ctx = context.Background()
	)
	for i := 0; i < _maxPn; i++ {
		time.Sleep(time.Second)
		vUsers, err := s.dao.BigVUsers(ctx, pn)
		if err != nil {
			log.Error("NewstarIdentity i(%d) error(%+v)", i, err)
			continue
		}
		if len(vUsers) == 0 {
			log.Warn("NewstarIdentity success i(%d)", i)
			break
		}
		idMap := make(map[int64]int64, 100)
		for _, vUser := range vUsers {
			time.Sleep(100 * time.Millisecond)
			pn = vUser.ID
			officialRes, officialErr := s.memberClient.OfficialDoc(ctx, &memberAPI.MidReq{Mid: vUser.Mid})
			if officialErr != nil {
				log.Error("NewstarIdentity s.memberClient.OfficialDoc(%d) error(%+v)", vUser.Mid, officialErr)
				continue
			}
			switch officialRes.State {
			case 0:
				idMap[vUser.ID] = _checking
			default:
				// 除了审核中都更新为_unDo,finishBigV 会更新最终状态
				if vUser.IsIdentity != _unDo {
					idMap[vUser.ID] = _unDo
				}
			}
		}
		if len(idMap) > 0 {
			if _, err = s.dao.UpIdentity(ctx, idMap); err != nil {
				log.Errorc(ctx, "IdentityChecking s.dao.UpIdentity idMap(%+v) error(%+v)", idMap, err)
			}
		}
	}
}
