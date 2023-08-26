package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/ugc-season/job/model/retry"
)

func (s *Service) retryproc() {
	defer s.waiter.Done()
	for {
		if s.closeRetry {
			return
		}
		var (
			c   = context.TODO()
			bs  []byte
			err error
		)
		bs, err = s.PopRetryItem(c)
		if err != nil || bs == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		msg := &retry.Info{}
		if err = json.Unmarshal(bs, msg); err != nil {
			log.Error("json.Unretry dedeSyncmarshal(%s) error(%v)", bs, err)
			continue
		}
		log.Info("retry %s %s", msg.Action, bs)
		switch msg.Action {
		case retry.FailSeasonAdd:
			s.seasonUpdate(msg.Data.SeasonID)
		case retry.FailUpSeasonCache:
			_ = s.updateSeasonCache(msg.Data.SeasonID, msg.Data.Mid, retry.ActionUp, msg.Data.Ptime)
		case retry.FailDelSeasonCache:
			_ = s.updateSeasonCache(msg.Data.SeasonID, msg.Data.Mid, retry.ActionDel, msg.Data.Ptime)
		case retry.FailUpSeasonStat:
			s.seasonResUpdate(c, msg.Data.SeasonID)
		case retry.FailForPubArchiveDatabus:
			s.SeasonNotify(msg.Data.SeasonID, msg.Data.SeasonWithArchive)
		default:
			continue
		}
	}
}

// PushToRetryList rpush fail item to redis
func (s *Service) PushToRetryList(c context.Context, a interface{}) (err error) {
	var (
		conn = s.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(a); err != nil {
		log.Error("json.Marshal(%v) error(%v)", a, err)
		return
	}
	if _, err = conn.Do("RPUSH", retry.FailList, bs); err != nil {
		log.Error("conn.Do(RPUSH, %s, %s) error(%v)", retry.FailList, bs, err)
	}
	return
}

// PopRetryItem lpop fail item from redis
func (s *Service) PopRetryItem(c context.Context) (bs []byte, err error) {
	var conn = s.redis.Get(c)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("LPOP", retry.FailList)); err != nil && err != redis.ErrNil {
		log.Error("redis.Bytes(conn.Do(LPOP, %s)) error(%v)", retry.FailList, err)
		return
	}
	return
}
