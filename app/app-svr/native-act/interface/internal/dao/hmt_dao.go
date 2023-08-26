package dao

import (
	"context"

	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
)

type hmtDao struct {
	client hmtgrpc.ChannelRPCClient
}

func (d *hmtDao) ChannelFeed(c context.Context, req *hmtgrpc.ChannelFeedReq) (*hmtgrpc.ChannelFeedReply, error) {
	rly, err := d.client.ChannelFeed(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
