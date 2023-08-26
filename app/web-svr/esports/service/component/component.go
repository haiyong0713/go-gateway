package component

import (
	"github.com/jinzhu/gorm"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/conf"
)

var (
	GlobalMemcached *memcache.Memcache
	GlobalRedis     *redis.Redis
	GlobalDB        *sql.DB
	GlobalOrm       *gorm.DB
)

func InitMemcached(conf *conf.Config) error {
	GlobalMemcached = memcache.New(conf.Memcached)
	if GlobalMemcached == nil {
		log.Error("InitMemcache Error, GlobalMemcached:%+v", GlobalMemcached)
		return xecode.Errorf(xecode.ServerErr, "Memcache初始化失败")
	}
	return nil
}

func InitRedis(conf *conf.Config) error {
	GlobalRedis = redis.NewRedis(conf.CommonRedis)
	if GlobalRedis == nil {
		log.Error("InitRedis Error, GlobalRedis:%+v", GlobalRedis)
		return xecode.Errorf(xecode.ServerErr, "Redis初始化失败")
	}
	return nil
}

func InitDB(conf *conf.Config) error {
	GlobalDB = sql.NewMySQL(conf.SqlCfg)
	if GlobalDB == nil {
		log.Error("InitDB Error, GlobalDB:%+v", GlobalDB)
		return xecode.Errorf(xecode.ServerErr, "DB初始化失败")
	}
	GlobalOrm = orm.NewMySQL(conf.OrmCfg)
	return nil
}

func InitByCfg() error {
	if err := InitMemcached(conf.Conf); err != nil {
		return err
	}
	if err := InitRedis(conf.Conf); err != nil {
		return err
	}
	if err := InitDB(conf.Conf); err != nil {
		return err
	}
	return nil
}

func Close() {
	_ = GlobalDB.Close()
	_ = GlobalMemcached.Close()
	_ = GlobalRedis.Close()
}
