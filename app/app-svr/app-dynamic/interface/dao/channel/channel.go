package channel

import (
	"context"

	"go-common/library/log"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

func (d *Dao) SubscribedChannel(c context.Context, mid int64) (res *channelgrpc.SubscribeReply, err error) {
	res, err = d.grpcClient.Subscribe(c, &channelgrpc.SubscribeReq{Mid: mid, SubVersion: 1})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) ChannelSort(c context.Context, mid int64, action int32, stick, normal string) (err error) {
	arg := &channelgrpc.UpdateSubscribeReq{Mid: mid, Action: action, Tops: stick, Cids: normal}
	if _, err = d.grpcClient.UpdateSubscribe(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) SearchChannelsInfo(c context.Context, mid int64, channelIDs []int64) (map[int64]*channelgrpc.SearchChannelCard, error) {
	var (
		args = &channelgrpc.SearchChannelsInfoReq{Mid: mid, Cids: channelIDs}
	)
	reply, err := d.grpcClient.SearchChannelsInfo(c, args)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	res := make(map[int64]*channelgrpc.SearchChannelCard)
	for _, channel := range reply.GetCards() {
		if channel == nil {
			continue
		}
		res[channel.Cid] = channel
	}
	return res, nil
}

func (d *Dao) RelativeChannel(c context.Context, mid int64, channelIDs []int64) ([]*channelgrpc.RelativeChannel, error) {
	var (
		args = &channelgrpc.RelativeChannelReq{Mid: mid, Cids: channelIDs}
	)
	reply, err := d.grpcClient.RelativeChannel(c, args)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.GetCards(), nil
}

func (d *Dao) ChannelList(c context.Context, mid int64, ctype int32, offset string) (*channelgrpc.ChannelListReply, error) {
	var arg = &channelgrpc.ChannelListReq{Mid: mid, CategoryType: ctype, Offset: offset}
	reply, err := d.grpcClient.ChannelList(c, arg)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply, nil
}
