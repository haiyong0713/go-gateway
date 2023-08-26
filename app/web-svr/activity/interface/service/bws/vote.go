package bws

import (
	"context"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

func (s *Service) AddVote(ctx context.Context, mid, pid, result, ts int64) error {
	nowTs := time.Now().Unix()
	if nowTs-ts > s.c.Bws.Bws2020VoteTs {
		return ecode.BwsPageOverTime
	}
	bid := s.c.Bws.Bws2020Bid
	userToken, err := s.midToKey(ctx, bid, mid)
	if err != nil {
		return err
	}
	points, err := s.dao.BwsPoints(ctx, []int64{pid})
	if err != nil {
		log.Error("AddVote s.dao.BwsPoints pid:%d error(%v)", pid, err)
		return ecode.ActivityPointFail
	}
	point, ok := points[pid]
	if !ok || point == nil || point.Bid != bid {
		return ecode.ActivityIDNotExists
	}
	logID, err := s.dao.UserUnFinishVoteID(ctx, userToken, pid)
	if err != nil {
		log.Error("AddVote UserUnFinishVoteID userToken:%s pid:%d error:%v", userToken, pid, err)
		return ecode.BwsHasVote
	}
	if logID > 0 {
		return ecode.BwsHasVote
	}
	if _, err = s.dao.AddVoteLog(ctx, userToken, pid, result); err != nil {
		log.Error("AddVote s.dao.AddVoteLog userToken:%s pid:%d error:%v", userToken, pid, err)
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserUnFinishVoteID(ctx, userToken, pid)
		})
	})
	return nil
}

func (s *Service) VoteClear(ctx context.Context, mid, pid, result int64) error {
	if pid != s.c.Bws.Bws2020GamePid {
		return xecode.RequestErr
	}
	points, err := s.dao.BwsPoints(ctx, []int64{pid})
	if err != nil {
		log.Error("VoteClear s.dao.BwsPoints pid:%d error(%v)", pid, err)
		return ecode.ActivityPointFail
	}
	point, ok := points[pid]
	if !ok || point == nil || point.Bid != s.c.Bws.Bws2020Bid {
		return ecode.ActivityIDNotExists
	}
	if point.Ower != mid && !s.isAdmin(mid) {
		return ecode.ActivityNotOwner
	}
	voteLogs, err := s.dao.UnFinishVoteLog(ctx, point.ID)
	if err != nil {
		log.Error("VoteClear UnFinishVoteLog error:%v", err)
		return err
	}
	for _, v := range voteLogs {
		if v != nil {
			unlockUser := v.UserToken
			s.cache.Do(ctx, func(ctx context.Context) {
				if v.Result == result {
					if unlockErr := s.Unlock2020(ctx, mid, true, &bwsmdl.ParamUnlock20{Bid: s.c.Bws.Bws2020Bid, Pid: point.ID, Key: unlockUser}); unlockErr != nil {
						log.Error("VoteClear Unlock2020 userToken:%s error:%v", unlockUser, unlockErr)
					}
				}
				retry(func() error {
					return s.dao.DelCacheUserUnFinishVoteID(ctx, unlockUser, pid)
				})
			})
		}
	}
	affected, err := s.dao.FinishVoteLog(ctx)
	if err != nil {
		log.Error("VoteClear error:%v", err)
		return err
	}
	log.Warn("VoteClear finish count:%d", affected)
	return nil
}
