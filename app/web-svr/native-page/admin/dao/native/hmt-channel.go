package native

import (
	"context"

	chagrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	"go-common/library/net/metadata"
)

func (d *Dao) ChannelFeed(c context.Context, cid, mid int64, buvid string, offset, ps int32) (*chagrpc.ChannelFeedReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	return d.chaClient.ChannelFeed(c, &chagrpc.ChannelFeedReq{Cid: cid, Mid: mid, Buvid: buvid, Offset: offset, Ps: ps, Context: &chagrpc.ChannelContext{Ip: ip}})
}
