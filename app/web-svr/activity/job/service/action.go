package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
)

// actionDealProc .
func (s *Service) actionDealProc(i int) {
	defer s.waiter.Done()
	var (
		ch  = s.subActionCh[i]
		err error
	)
	for {
		ms, ok := <-ch
		if !ok {
			log.Warn("actionDealProc s.actionDealProc(%d) quit", i)
			return
		}
		score := ms.Action + ms.ExtraAction
		if err = s.incrLikeExtend(context.Background(), ms.Lid, score); err != nil {
			log.Info("actionDealProc s.incrLikeExtend(%d) lid:%d score:%d error(%v)", i, ms.Lid, score, err)
		} else {
			log.Info("actionDealProc success(%d) lid:%d score:%d", i, ms.Lid, score)
		}
	}
}

// incrLikeExtend batch insert like_extend table.
func (s *Service) incrLikeExtend(c context.Context, lid, score int64) (err error) {
	var (
		lidInfo *like.Extend
		lids    []int64
	)
	// 先查询 避免频繁使用 ON DUPLICATE KEY UPDATE 语句
	if lidInfo, err = s.dao.RawLikeExtend(c, lid); err != nil {
		log.Error(" s.dao.RawLikeExtend(%v) error(%v)", lids, err)
		return
	}
	if lidInfo != nil && lidInfo.ID > 0 {
		err = s.dao.UpExtend(c, lidInfo.Lid, score)
	} else {
		err = s.dao.AddExtend(c, lid, score)
	}
	return
}

// actionProc .
func (s *Service) actionProc(msg json.RawMessage) (err error) {
	var (
		act = new(like.Action)
	)
	if err = json.Unmarshal(msg, act); err != nil {
		log.Error("actionProc json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	s.subActI++
	s.subActionCh[s.subActI%_sharding] <- act //均匀发任务
	if s.subActI > _sharding {
		s.subActI = 0
	}
	return
}
