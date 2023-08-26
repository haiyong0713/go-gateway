package dao

import (
	"context"

	"go-common/library/cache/memcache"
	"go-common/library/conf/paladin"
	"go-common/library/log"
)

//go:generate kratos tool mcgen
type _mc interface {
}

func NewMC() (mc *memcache.Memcache, err error) {
	var cfg struct {
		Client *memcache.Config
	}
	if err = paladin.Get("memcache.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	mc = memcache.New(cfg.Client)
	return
}

func (d *dao) PingMC(ctx context.Context) (err error) {
	if err = d.mc.Set(ctx, &memcache.Item{Key: "ping", Value: []byte("pong"), Expiration: 0}); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}
