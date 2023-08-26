package domain

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	bm "go-common/library/net/http/blademaster"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/admin/conf"
)

const (
	domainPrefix = "act_domain_prefix"
	separator    = ":"
)

// Dao struct user of Dao.
type Dao struct {
	c            *conf.Config
	db           *sql.DB
	orm          *gorm.DB
	redis        *redis.Pool
	httpClient   *bm.Client
	listUrl      string
	fawkesAddUrl string
	fawkesGetUrl string
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		db:           sql.NewMySQL(c.MySQL.Lottery),
		orm:          orm.NewMySQL(c.ORM),
		redis:        redis.NewPool(c.Redis.Config),
		httpClient:   bm.NewClient(c.HTTPClient),
		listUrl:      c.ActDomainConf.APIHost + c.ActDomainConf.DomainListUrl,
		fawkesAddUrl: c.ActDomainConf.FawkesConf.Host + c.ActDomainConf.FawkesConf.AddUrl,
		fawkesGetUrl: c.ActDomainConf.FawkesConf.Host + c.ActDomainConf.FawkesConf.GetUrl,
	}
	return
}

// Ping Dao
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return domainPrefix + separator + strings.Join(strArgs, separator)
}
