package component

import (
	"github.com/jinzhu/gorm"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-gateway/app/web-svr/esports/admin/conf"
)

var (
	GlobalMemcached    *memcache.Memcache
	GlobalAutoSubCache *redis.Pool
	GlobalDB           *gorm.DB
)

func InitMemcached(conf *conf.Config) {
	GlobalMemcached = memcache.New(conf.Memcached)
}

func InitRedis(conf *conf.Config) {
	GlobalAutoSubCache = redis.NewPool(conf.AutoSubCache)
}

func InitRelations(conf *conf.Config) {
	GlobalDB = orm.NewMySQL(conf.ORM)
}

func InitByCfg() error {
	InitMemcached(conf.Conf)
	InitRedis(conf.Conf)
	InitRelations(conf.Conf)
	return nil
}

func Close() {
	_ = GlobalDB.Close()
	_ = GlobalMemcached.Close()
	_ = GlobalAutoSubCache.Close()
}
