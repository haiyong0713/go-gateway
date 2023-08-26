package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive/job/model/retry"
	"go-gateway/app/app-svr/archive/service/api"
)

func (s *Service) retryproc(key string) {
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
		bs, err = s.PopFail(c, key)
		if err != nil || bs == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		msg := &retry.Info{}
		if err = json.Unmarshal(bs, msg); err != nil {
			log.Error("json.Unmarshal msg(%s) error(%+v)", bs, err)
			continue
		}
		log.Info("retry %s %s", msg.Action, bs)
		if key == retry.FailList {
			s.limiter.retry.Wait()
		}
		switch msg.Action {
		case retry.FailUpCache:
			s.updateResultCache(&api.Arc{Aid: msg.Data.Aid, State: msg.Data.State}, nil, msg.Data.WithState)
		case retry.FailDatabus:
			s.sendNotify(msg.Data.DatabusMsg)
		case retry.FailUpVideoCache:
			s.upVideoCache(c, msg.Data.Aid)
		case retry.FailDelVideoCache:
			s.delVideoCache(c, msg.Data.Aid, msg.Data.Cids)
		case retry.FailResultAdd:
			s.arcUpdate(msg.Data.Aid, msg.Data.ArcAction, msg.Data.SeasonID)
		case retry.FailVideoShot:
			s.videoShotHandler(c, msg.Data.Cids)
		case retry.FailVideoFF:
			s.videoFFHandler(c, msg.Data.Cid)
		case retry.FailUpInternal:
			s.internalUpdate(msg.Data.Aid)
		case retry.FailInternalCache:
			s.internalCacheHandler(c, msg.Data.Aid)
		default:
			continue
		}
	}
}

// PushFail rpush fail item to redis
func (s *Service) PushFail(c context.Context, a interface{}, key string) {
	var (
		conn = s.redis.Get(c)
		bs   []byte
		err  error
	)
	defer conn.Close()
	if bs, err = json.Marshal(a); err != nil {
		log.Error("json.Marshal(%+v) error(%+v)", a, err)
		return
	}
	var cnt int
	if cnt, err = redis.Int(conn.Do("RPUSH", key, bs)); err != nil {
		log.Error("conn.Do(RPUSH) error(%+v)", err)
		return
	}
	if key == retry.FailVideoshotList && cnt <= s.c.Custom.VsMonitorSize {
		return
	}
	if (key == retry.FailList || key == retry.FailVideoFFList || key == retry.FailInternalList) && cnt <= s.c.Custom.MonitorSize {
		return
	}
	log.Error("日志告警 archive-job retry list is too big key(%s) len(%d)", key, cnt)
}

// PopFail lpop fail item from redis
func (s *Service) PopFail(c context.Context, key string) (bs []byte, err error) {
	var conn = s.redis.Get(c)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("LPOP", key)); err != nil && err != redis.ErrNil {
		log.Error("redis.Bytes(conn.Do(LPOP, %s)) error(%+v)", key, err)
		return
	}
	return
}
