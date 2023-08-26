package service

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"

	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_prefixSnStatPb = "ss_"
	_statLock       = "stat_lock_%d"
)

func snStatPBKey(sid int64) string {
	return _prefixSnStatPb + strconv.FormatInt(sid, 10)
}
func statWatchLock(aid int64) string {
	return fmt.Sprintf(_statLock, aid)
}

// updateSnCache purge stat info in cache
func (s *Service) updateSnCache(c context.Context, st *api.Stat) error {
	if st == nil {
		return nil
	}
	err := retry.WithAttempts(c, "updateSnCache", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		bs, err := st.Marshal()
		if err != nil {
			log.Error("st.Marshal error(%+v)", err)
			return err
		}
		conn := s.redis.Get(c)
		defer conn.Close()
		if _, err = conn.Do("SET", snStatPBKey(st.SeasonID), bs); err != nil {
			log.Error("conn.Do error(%+v)", err)
			return err
		}
		log.Info("SeasonStat update cache season(%d) stat(%+v) success", st.SeasonID, st)
		return nil
	})
	return err
}

// GetStCache get a season stat from cache.
func (s *Service) GetStCache(c context.Context, sid int64) (ss *api.Stat, err error) {
	var (
		key  = snStatPBKey(sid)
		conn = s.redis.Get(c)
	)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("conn.Do(GET,%s) error(%+v)", key, err)
		return nil, err
	}
	ss = new(api.Stat)
	if err = ss.Unmarshal(bs); err != nil {
		log.Error("Stat.Unmarshal error(%+v)", err)
		return nil, err
	}
	return ss, nil
}

// 加锁
func (s *Service) TryLock(ctx context.Context, key, value string, timeout int32) (bool, error) {
	var (
		conn = s.redis.Get(ctx)
	)
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", key, value, "EX", timeout, "NX"))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	if reply != "OK" {
		return false, nil
	}
	return true, nil
}

// UnLock 解锁
func (s *Service) UnLock(ctx context.Context, key string, value string) bool {
	var (
		conn = s.redis.Get(ctx)
	)
	defer conn.Close()
	msg, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Error("GetLock Msg key(%+v) value(%+v) err(%+v)", key, msg, err)
	}
	if msg == value {
		msg, _ := redis.Int64(conn.Do("DEL", key))
		if msg == 1 || msg == 0 {
			return true
		}
	}
	return false
}
