package push

import (
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	appres "go-gateway/app/app-svr/app-resource/interface/api/v1"
	"go-gateway/app/web-svr/appstatic/job/conf"
)

// Dao .
type Dao struct {
	c         *conf.Config
	db        *xsql.DB
	client    *bm.Client
	appresCli appres.AppResourceClient
}

// New creates a dao instance.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		db:     xsql.NewMySQL(c.MySQL),
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.appresCli, err = appres.NewClient(c.AppresClient); err != nil {
		panic(err)
	}
	return
}
