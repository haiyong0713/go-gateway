package dao

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

const _pushKey = "appstatic_admin_push"

func (d *Dao) PushTime(c context.Context) (res int64, err error) {
	var (
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", _pushKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("PushTime conn.Do(GET,%s) error(%v)", _pushKey, err)
		}
	}
	return
}

// AddPushTime .
func (d *Dao) AddPushTime(c context.Context, value int64) (err error) {
	var (
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("SETEX", _pushKey, d.redisPushExpire, value); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", _pushKey, d.redisPushExpire, value)
		return
	}
	return
}

// TryLock ...
func (d *Dao) TryLock(ctx context.Context, key string, timeout int32) (bool, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", key, 1, "EX", timeout, "NX"))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	if reply == "OK" {
		return true, nil
	}
	return false, nil
}

// UnLock ...
func (d *Dao) UnLock(ctx context.Context, key string) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}
