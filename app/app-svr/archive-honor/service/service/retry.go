package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive-honor/service/api"
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
		msg := &api.RetryInfo{}
		if err = json.Unmarshal(bs, msg); err != nil {
			log.Error("json.Unmarshal (%s) error(%v)", bs, err)
			continue
		}
		log.Info("retry action(%s) %s", msg.Action, bs)
		switch msg.Action {
		case api.ActionUpdate:
			s.HonorUpdate(context.Background(), msg.Data.Aid, msg.Data.Type, msg.Data.URL, msg.Data.Desc, msg.Data.NaUrl)
		case api.ActionDel:
			s.HonorDel(context.Background(), msg.Data.Aid, msg.Data.Type)
		default:
			continue
		}
	}
}

// PushToRetryList rpush fail item to redis
func (s *Service) PushToRetryList(c context.Context, a interface{}) {
	var (
		conn = s.redis.Get(c)
		bs   []byte
		err  error
	)
	defer conn.Close()
	if bs, err = json.Marshal(a); err != nil {
		log.Error("json.Marshal(%v) error(%v)", a, err)
		return
	}
	if _, err = conn.Do("RPUSH", api.FailList, bs); err != nil {
		log.Error("conn.Do(RPUSH, %s, %s) error(%v)", api.FailList, bs, err)
	}
}

// PopRetryItem lpop fail item from redis
func (s *Service) PopRetryItem(c context.Context) (bs []byte, err error) {
	var conn = s.redis.Get(c)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("LPOP", api.FailList)); err != nil && err != redis.ErrNil {
		log.Error("redis.Bytes(conn.Do(LPOP, %s)) error(%v)", api.FailList, err)
		return
	}
	return
}
