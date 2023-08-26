package sidebar

import (
	"context"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	resmodel "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"
)

// Dao is sidebar dao
type Dao struct {
	resRPC *resrpc.Service
}

// New initial sidebar dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		resRPC: resrpc.New(c.ResourceRPC),
	}
	return
}

// Sidebars from resource service
func (d *Dao) Sidebars(c context.Context) (res *resmodel.SideBars, err error) {
	res, err = d.resRPC.SideBars(c)
	return
}
