package dao

import (
	"context"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

type channelDao struct {
	client channelgrpc.ChannelRPCClient
}

func (d *channelDao) Infos(c context.Context, req *channelgrpc.InfosReq) (*channelgrpc.InfosReply, error) {
	rly, err := d.client.Infos(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
