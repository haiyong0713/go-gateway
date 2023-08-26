package subscription

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// grpc client
	tunnelClient tunnelgrpc.TunnelClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.tunnelClient, err = tunnelgrpc.NewClient(c.TunnelGRPC); err != nil {
		panic(err)
	}
	return
}
