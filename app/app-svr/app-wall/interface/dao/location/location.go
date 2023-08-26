package location

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/conf"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

type Dao struct {
	c       *conf.Config
	locGRPC locgrpc.LocationClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) InfoGRPC(c context.Context, ipaddr string) (info *locgrpc.InfoReply, err error) {
	if info, err = d.locGRPC.Info(c, &locgrpc.InfoReq{Addr: ipaddr}); err != nil {
		log.Error("%v", err)
	}
	return
}
