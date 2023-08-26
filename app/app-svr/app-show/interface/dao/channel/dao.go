package channel

import (
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"go-gateway/app/app-svr/app-show/interface/conf"
)

type Dao struct {
	c          *conf.Config
	grpcClient channelgrpc.ChannelRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClient, err = channelgrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(err)
	}
	return
}
