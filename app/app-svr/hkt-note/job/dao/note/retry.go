package note

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

func (d *Dao) AddCacheRetry(c context.Context, key string, val string, score int64) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("ZADD", key, score, val); err != nil {
		log.Error("retryError AddCacheRetry key(%s) val(%s) score(%d) err(%+v)", key, val, score, err)
	}
}

func (d *Dao) CacheRetry(c context.Context, key string) (string, error) {
	var (
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	res, err := redis.Values(conn.Do("ZRANGE", key, 0, 0, "WITHSCORES"))
	if err != nil {
		return "", err
	}
	if len(res) == 0 {
		return "", nil
	}
	var (
		score int64
		val   string
	)
	if res, err = redis.Scan(res, &val, &score); err != nil {
		log.Error("redis.Scan(%v) error(%v)", res, err)
		return "", err
	}
	if score > time.Now().Unix() {
		log.Info("retryInfo key(%s) val(%s) score(%d) still need to wait,skip", key, val, score)
		return "", nil
	}
	return val, nil
}

func (d *Dao) RemCacheRetry(c context.Context, key string, val string) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("ZREM", key, val); err != nil {
		log.Error("retryError RemCacheRetry key(%s) val(%s) err(%+v)", key, val, err)
	}
}
