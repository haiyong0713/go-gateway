package garb

import (
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	api "git.bilibili.co/bapis/bapis-go/garb/service"
	live2dgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/live2d/service"
)

// Dao is space dao
type Dao struct {
	c            *conf.Config
	garbClient   api.GarbClient
	live2dClient live2dgrpc.GarbCharacterClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.garbClient, err = api.NewClient(c.GarbGRPC); err != nil {
		panic(err)
	}
	if d.live2dClient, err = live2dgrpc.NewClient(c.Live2dGRPC); err != nil {
		panic(err)
	}
	return
}
