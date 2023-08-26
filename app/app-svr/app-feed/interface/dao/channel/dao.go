package channel

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"

	"go-common/library/sync/errgroup.v2"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
)

const (
	_videoChannel = 3
)

// Dao is archive dao.
type Dao struct {
	c             *conf.Config
	grpcClient    channelgrpc.ChannelRPCClient
	hmtgrpcClient hmtchannelgrpc.ChannelRPCClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	//grpc
	var err error
	if d.grpcClient, err = channelgrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.hmtgrpcClient, err = hmtchannelgrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Channels(c context.Context, channelIDs []int64, mid int64) (res []*channelgrpc.ChannelCard, err error) {
	var channelTmp *channelgrpc.TianmaDetailReply
	arg := &channelgrpc.TianmaDetailReq{ChannelIds: channelIDs, Mid: mid}
	if channelTmp, err = d.grpcClient.TianmaDetail(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = channelTmp.GetChannels()
	return
}

func (d *Dao) Details(c context.Context, tids []int64) (res map[int64]*channelgrpc.ChannelCard, err error) {
	args := &channelgrpc.SimpleChannelDetailReq{Cids: tids}
	var details *channelgrpc.SimpleChannelDetailReply
	if details, err = d.grpcClient.SimpleChannelDetail(c, args); err != nil {
		log.Error("%+v", err)
		return
	}
	res = details.GetChannelMap()
	return
}

// ResourceChannels .
func (d *Dao) ResourceChannels(c context.Context, aids []int64, mid int64) (res map[int64][]*channelgrpc.Channel, err error) {
	var (
		g     = errgroup.WithContext(c)
		mutex = sync.Mutex{}
	)
	res = map[int64][]*channelgrpc.Channel{}
	for _, v := range aids {
		aid := v
		g.Go(func(ctx context.Context) (err error) {
			reply, err := d.grpcClient.ResourceChannels(ctx, &channelgrpc.ResourceChannelsReq{Rid: aid, Mid: mid, Type: _videoChannel})
			if err != nil {
				log.Error("d.grpcClient.ResourceChannels(%d,%d) error(%v)", aid, mid, err)
				return
			}
			if reply != nil {
				mutex.Lock()
				res[aid] = reply.Channels
				mutex.Unlock()
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (d *Dao) ChannelFeed(ctx context.Context, param *feed.VerticalChannelParam) (*hmtchannelgrpc.ChannelFeedReply, error) {
	return d.hmtgrpcClient.ChannelFeed(ctx, &hmtchannelgrpc.ChannelFeedReq{
		Cid:     param.ChannelID,
		Mid:     param.Mid,
		Buvid:   param.Buvid,
		Context: &hmtchannelgrpc.ChannelContext{Ip: param.Ip},
		Offset:  param.Offset,
		Ps:      param.Ps,
		Tag:     param.Tag,
	})
}

func (d *Dao) ChannelTag(ctx context.Context, param *feed.VerticalTagParam) (*hmtchannelgrpc.TagReply, error) {
	const _tagMax = 20
	return d.hmtgrpcClient.ChannelTag(ctx, &hmtchannelgrpc.TagReq{
		Cid:     param.ChannelID,
		Mid:     param.Mid,
		Buvid:   param.Buvid,
		Context: &hmtchannelgrpc.ChannelContext{Ip: metadata.String(ctx, metadata.RemoteIP)},
		Ps:      _tagMax,
	})
}

func (d *Dao) ChannelWhite(ctx context.Context, param *feed.VerticalChannelParam, id int32) (*hmtchannelgrpc.WhiteReply, error) {
	return d.hmtgrpcClient.White(ctx, &hmtchannelgrpc.WhiteReq{
		Id:    id,
		Cid:   param.ChannelID,
		Mid:   param.Mid,
		Buvid: param.Buvid,
	})
}
