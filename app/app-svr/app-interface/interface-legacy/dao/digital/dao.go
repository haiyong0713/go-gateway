package digital

import (
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	digitalgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/digital/service"
)

type Dao struct {
	c             *conf.Config
	digitalClient digitalgrpc.GarbDigitalClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.digitalClient, err = digitalgrpc.NewClientGarbDigital(c.DigitalGRPC); err != nil {
		panic(err)
	}
	return
}
