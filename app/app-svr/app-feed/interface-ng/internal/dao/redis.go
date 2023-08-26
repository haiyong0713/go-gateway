package dao

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/log"
)

func NewRedis() (*redis.Redis, func(), error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err := paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return nil, nil, err
	}
	if err := ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return nil, nil, err
	}
	r := redis.NewRedis(&cfg)
	closeFn := func() { r.Close() }
	return r, closeFn, nil
}

func (d *dao) PingRedis(ctx context.Context) error {
	if _, err := d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
		return err
	}
	return nil
}
