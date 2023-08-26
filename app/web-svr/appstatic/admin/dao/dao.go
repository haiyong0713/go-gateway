package dao

import (
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/boss"
	"go-common/library/database/orm"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/appstatic/admin/conf"

	"github.com/jinzhu/gorm"
)

// Dao .
type Dao struct {
	DB              *gorm.DB
	GWDB            *gorm.DB
	c               *conf.Config
	client          *httpx.Client
	redis           *redis.Pool
	playerRedis     []*redis.Pool
	redisPushExpire int32
	boss            *boss.Boss
	host            *conf.Host
}

// New new a instance
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// db
		DB:              orm.NewMySQL(c.ORM),
		GWDB:            orm.NewMySQL(c.GWDB),
		c:               c,
		client:          httpx.NewClient(c.HTTPClient),
		redis:           redis.NewPool(c.Redis.Config),
		redisPushExpire: int32(time.Duration(c.Cfg.Push.Expire) / time.Second),
		boss:            boss.New(c.Boss),
		host:            c.Host,
	}
	for _, v := range c.PlayerRedis { // 初始化app-player使用的shjd & ylf的redis
		d.playerRedis = append(d.playerRedis, redis.NewPool(v.Config))
	}
	d.DB.LogMode(true)
	return
}

// Close close connection of db , mc.
func (d *Dao) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
