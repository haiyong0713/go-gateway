package redis_lock

import (
	"context"

	"go-common/library/cache/redis"
)

// TryLock ...
func (d *Dao) TryLock(c context.Context, key string, timeout int32) (bool, error) {
	var conn = d.redis.Get(c)
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", key, "1", "EX", timeout, "NX"))
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

// UnLock ...
func (d *Dao) UnLock(c context.Context, key string) (err error) {
	var conn = d.redis.Get(c)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}
