package component

import (
	"go-common/library/cache/memcache"

	"go-gateway/app/web-svr/esports/interface/conf"
)

var (
	GlobalMemcached           *memcache.Memcache
	GlobalMemcached4UserGuess *memcache.Memcache
)

func initMemcahced(cfg *conf.Config) {
	GlobalMemcached = memcache.New(cfg.Memcached)
	GlobalMemcached4UserGuess = memcache.New(cfg.Memcached4UserGuess)
}
