package hmt_channel

import (
	"context"

	chagrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/native-page/interface/conf"
)

// Dao is rpc dao.
type Dao struct {
	chaClient chagrpc.ChannelRPCClient
	conf      *conf.Config
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.chaClient, err = chagrpc.NewClient(c.HmtChannelClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) ChannelFeed(c context.Context, cid, mid int64, buvid string, offset, ps int32) (*chagrpc.ChannelFeedReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	return d.chaClient.ChannelFeed(c, &chagrpc.ChannelFeedReq{Cid: cid, Mid: mid, Buvid: buvid, Offset: offset, Ps: ps, Context: &chagrpc.ChannelContext{Ip: ip}})
}
