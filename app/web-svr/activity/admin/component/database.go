package component

import (
	"go-gateway/app/web-svr/activity/admin/conf"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
)

var (
	GlobalDB          *xsql.DB
	S10PointCostMC    *memcache.Memcache
	S10PointShopRedis *redis.Pool
	ExportDB          *xsql.DB
	RewardsDB         *xsql.DB
)

func New(c *conf.Config) {
	S10PointShopRedis = redis.NewPool(c.S10PointShopRedis)
	GlobalDB = xsql.NewMySQL(c.S10MySQL)
	S10PointCostMC = memcache.New(c.S10PointCostMC)
	ExportDB = xsql.NewMySQL(c.Export.Lottery)
	RewardsDB = xsql.NewMySQL(c.RewardsMySQL)
}

func Close() {
	GlobalDB.Close()
	S10PointShopRedis.Close()
	S10PointCostMC.Close()
	ExportDB.Close()
}
