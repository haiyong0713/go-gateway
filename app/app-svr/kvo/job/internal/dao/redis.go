package dao

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/log"
)

func NewRedis() (r *redis.Redis, err error) {
	var cfg struct {
		Redis *redis.Config
	}
	if err = paladin.Get("redis.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(cfg.Redis)
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}
