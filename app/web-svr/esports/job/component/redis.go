package component

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/esports/job/conf"
)

var (
	GlobalAutoSubCache *redis.Pool
	GlobalMC           *memcache.Memcache
)

func InitRedis() {
	GlobalAutoSubCache = redis.NewPool(conf.Conf.AutoSubCache)
}

func InitCache() {
	GlobalMC = memcache.New(conf.Conf.Memcache)
}
