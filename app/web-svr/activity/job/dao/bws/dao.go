package bws

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao  dao
type Dao struct {
	c             *conf.Config
	db            *sql.DB
	redis         *redis.Pool
	httpClient    *blademaster.Client
	addAchieveURL string
	rechargeURL   string
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:             c,
		db:            sql.NewMySQL(c.MySQL.Like),
		redis:         redis.NewPool(c.Redis.Config),
		httpClient:    blademaster.NewClient(c.HTTPClient),
		addAchieveURL: c.Host.APICo + _addAchieveURI,
		rechargeURL:   c.Host.APICo + _rechargeAwardURI,
	}
}
