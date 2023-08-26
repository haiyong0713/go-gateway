package dao

import (
	"go-gateway/app/web-svr/activity/job/conf"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
)

var (
	GlobalDB          *xsql.DB
	GlobalReadDB      *xsql.DB
	S10PointCostMC    *memcache.Memcache
	S10PointShopRedis *redis.Pool
)

func New(c *conf.Config) {
	S10PointShopRedis = redis.NewPool(c.S10PointShopRedis)
	GlobalDB = xsql.NewMySQL(c.S10MySQL)
	GlobalReadDB = xsql.NewMySQL(c.MySQL.Read)
	S10PointCostMC = memcache.New(c.S10PointCostMC)

}

func Close() {
	GlobalDB.Close()
	GlobalReadDB.Close()
	S10PointShopRedis.Close()
	S10PointCostMC.Close()
}
