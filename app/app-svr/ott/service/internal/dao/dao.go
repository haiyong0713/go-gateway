package dao

import (
	"context"

	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/ott/service/conf"

	arccli "git.bilibili.co/bapis/bapis-go/archive/service"
)

type Dao struct {
	c         *conf.Config
	arcClient arccli.ArchiveClient
	client    *httpx.Client
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
		// mysql
		client: httpx.NewClient(c.HTTPClient),
	}
	var err error
	if dao.arcClient, err = arccli.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	return nil
}
