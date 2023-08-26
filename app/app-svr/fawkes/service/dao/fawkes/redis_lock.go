package fawkes

import (
	"context"

	"go-common/library/cache/redis"
)

// TryLock ...
func (d *Dao) TryLock(ctx context.Context, key string, timeout int32) (bool, error) {
	reply, err := redis.String(d.redis.Do(ctx, "SET", key, 1, "EX", timeout, "NX"))
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
func (d *Dao) UnLock(ctx context.Context, key string) (err error) {
	_, err = d.redis.Do(ctx, "DEL", key)
	return
}
