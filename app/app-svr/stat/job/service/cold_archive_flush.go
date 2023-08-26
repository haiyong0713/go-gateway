/**
 * 冷门稿件定时刷新
 */
package service

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_BabyCnt = 100
)

func (s *Service) babyNum(c context.Context, key string) (length int64, err error) {
	conn := s.statRedis.Get(c)
	defer conn.Close()
	if length, err = redis.Int64(conn.Do("SCARD", key)); err != nil {
		log.Error("babyNum key %s, Err %v", key, err)
		return
	}
	return
}

func (s *Service) popBaby(c context.Context, key string, cnt int64) (aids []int64, err error) {
	conn := s.statRedis.Get(c)
	defer conn.Close()
	if aids, err = redis.Int64s(conn.Do("SPOP", key, cnt)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
	}
	return
}

// HappyBaby can make babies happy for ever
func (s *Service) HappyBaby() {
	btime := time.Now().AddDate(0, 0, -1)
	log.Info("start HappyBaby")
	for i := int64(0); i < _BabyGroup; i++ {
		var (
			key = babyKey(i, btime)
			c   = context.Background()
			err error
			cnt int64
		)
		if cnt, err = s.babyNum(c, key); err != nil {
			log.Info("s.babyNum(%s) error(%v)", key, err)
			continue
		}
		log.Info("happy baby(%s) total(%d)", key, cnt)
		if cnt == 0 {
			continue
		}
		for {
			time.Sleep(time.Duration(s.c.Custom.BabySleepTick))
			var (
				aids []int64
				err  error
			)
			if aids, err = s.popBaby(c, key, _BabyCnt); err != nil {
				log.Error("s.popBaby(%s) error(%v)", key, err)
				continue
			}
			if len(aids) == 0 {
				log.Info("s.popBaby(%s) end", key)
				break
			}
			for _, aid := range aids {
				log.Info("flush cold archive to DB, aid(%d)", aid)
				_ = s.flushRedisToDB(c, aid)
			}
		}
	}
	log.Info("HappyBaby ended.")
}
