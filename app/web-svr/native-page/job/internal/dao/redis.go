package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/credis"

	"go-gateway/app/web-svr/native-page/job/internal/model"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
)

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -struct_name=progressDao -key=pageProgressParamsKey -expire=d.cfg.PageProgressParamsExpire -encode=json
	AddCachePageProgressParams(c context.Context, pageID int64, data []*model.ProgressParam) error
}

func pageProgressParamsKey(pageID int64) string {
	return fmt.Sprintf("page_prog_params_%d", pageID)
}

func sponsoredUpKey(mid int64) string {
	return fmt.Sprintf("sponsored_up_%d", mid)
}

func NewRedis() (r credis.Redis, cf func(), err error) {
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
	r = credis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func (d *dao) AddCacheSponsoredUp(c context.Context, mid int64) error {
	key := sponsoredUpKey(mid)
	if _, err := d.redis.Do(c, "set", key, true); err != nil {
		log.Errorc(c, "Fail to AddCacheSponsoredUp, key=%s error=%+v", key, err)
		return err
	}
	return nil
}
