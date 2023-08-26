package dao

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	bm "go-common/library/net/http/blademaster"
)

// Dao dao.
type Dao struct {
	conf              *conf.Config
	client            *bm.Client
	clientLongTimeout *bm.Client
}

func New(cfg *conf.Config) *Dao {
	return &Dao{
		conf:              cfg,
		client:            bm.NewClient(cfg.HTTPClient),
		clientLongTimeout: bm.NewClient(cfg.HTTPClientLongTimeOut),
	}
}

// Close close the resource.
func (d *Dao) Close() {
}

// Ping ping the resource.
func (d *Dao) Ping(_ context.Context) (err error) {
	return nil
}

// 让下游以1开头的page，以0开头
func fixPage(k int) int {
	return k + 1
}
