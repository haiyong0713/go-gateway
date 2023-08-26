package component

import (
	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/esports/interface/conf"
)

var (
	GlobalAutoSubCache *redis.Redis
)

func InitRedis(cfg *conf.Config) {
	GlobalAutoSubCache = redis.NewRedis(cfg.AutoSubCache)
}
