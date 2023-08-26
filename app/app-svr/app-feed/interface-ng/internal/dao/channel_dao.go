package dao

import (
	"context"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

type channelDao struct {
	channel channelgrpc.ChannelRPCClient
}

func (d *channelDao) Details(ctx context.Context, tids []int64) (map[int64]*channelgrpc.ChannelCard, error) {
	req := &channelgrpc.SimpleChannelDetailReq{Cids: tids}
	details, err := d.channel.SimpleChannelDetail(ctx, req)
	if err != nil {
		return nil, err
	}
	return details.GetChannelMap(), nil
}
