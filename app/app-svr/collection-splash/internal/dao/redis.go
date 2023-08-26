package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	pb "go-gateway/app/app-svr/collection-splash/api"

	"github.com/pkg/errors"
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

const _collectionSplashCacheKey = "collection-splash:list"

func (d *dao) CacheSplashList(ctx context.Context) ([]*pb.Splash, bool, error) {
	bt, err := redis.Bytes(d.redis.Do(ctx, "GET", _collectionSplashCacheKey))
	if err == redis.ErrNil {
		return nil, true, nil
	}
	if err != nil {
		return nil, true, err
	}
	var list []*pb.Splash
	if err = json.Unmarshal(bt, &list); err != nil {
		return nil, false, errors.WithStack(err)
	}
	return list, false, nil
}

const _collection_splash_exp = 20

func (d *dao) AddCacheSplashList(ctx context.Context, list []*pb.Splash) error {
	bt, err := json.Marshal(list)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = d.redis.Do(ctx, "SETEX", _collectionSplashCacheKey, _collection_splash_exp, bt); err != nil {
		return err
	}
	return nil
}
