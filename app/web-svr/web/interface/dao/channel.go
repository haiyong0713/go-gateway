package dao

import (
	"context"

	"go-common/library/log"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"

	"github.com/pkg/errors"
)

func (d *Dao) ChannelDetail(c context.Context, mid, channelID int64) (*channelgrpc.ChannelDetailReply, error) {
	arg := &channelgrpc.ChannelDetailReq{Mid: mid, ChannelId: channelID}
	detail, err := d.channelClient.ChannelDetail(c, arg)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return detail, nil
}

func (d *Dao) ResourceList(c context.Context, arg *channelgrpc.ResourceListReq) (res *channelgrpc.ResourceListReply, err error) {
	if res, err = d.channelClient.ResourceList(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) ResourceChannels(c context.Context, mid, aid, typ int64) (*channelgrpc.ResourceChannelsReply, error) {
	var (
		err error
		req = &channelgrpc.ResourceChannelsReq{
			Rid:  aid,
			Mid:  mid,
			Type: typ,
		}
		reply *channelgrpc.ResourceChannelsReply
	)
	if reply, err = d.channelClient.ResourceChannels(c, req); err != nil {
		err = errors.Wrapf(err, "d.channelClient.ResourceChannels(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) RecentChannelWebFeed(ctx context.Context, cid, mid int64, pn, ps int32) (aids []int64, count int32, err error) {
	req := &channelgrpc.RecentChannelWebFeedReq{
		Cid: cid,
		Mid: mid,
		Pn:  pn,
		Ps:  ps,
	}
	reply, err := d.channelClient.RecentChannelWebFeed(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	for _, val := range reply.GetList() {
		aids = append(aids, val.Id)
	}
	return aids, reply.GetTotal(), nil
}
