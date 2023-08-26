package dao

import (
	"context"

	"go-common/library/cache/redis"

	"go-common/library/conf/paladin.v2"
)

func NewRedis() (r *redis.Redis, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

// TryLock ...
func (d *dao) TryLock(ctx context.Context, key string, timeout int32) (bool, error) {
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
func (d *dao) UnLock(ctx context.Context, key string) error {
	_, err := d.redis.Do(ctx, "DEL", key)
	return err
}
