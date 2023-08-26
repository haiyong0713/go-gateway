package channel

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	baikegrpc "git.bilibili.co/bapis/bapis-go/community/interface/baike"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

// Dao is archive dao.
type Dao struct {
	c           *conf.Config
	grpcClient  channelgrpc.ChannelRPCClient
	baikeClient baikegrpc.BaikeClient
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
	if d.baikeClient, err = baikegrpc.NewClientBaike(c.ChannelGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) Details(c context.Context, tids []int64) (res map[int64]*channelgrpc.ChannelCard, err error) {
	args := &channelgrpc.SimpleChannelDetailReq{Cids: tids}
	var details *channelgrpc.SimpleChannelDetailReply
	if details, err = d.grpcClient.SimpleChannelDetail(c, args); err != nil {
		log.Error("%v")
		return
	}
	res = details.GetChannelMap()
	return
}

func (d *Dao) SearchChannel(c context.Context, mid int64, channelIDs []int64) (res map[int64]*channelgrpc.SearchChannel, err error) {
	var (
		args   = &channelgrpc.SearchChannelReq{Mid: mid, Cids: channelIDs}
		resTmp *channelgrpc.SearchChannelReply
	)
	if resTmp, err = d.grpcClient.SearchChannel(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetChannelMap()
	return
}

func (d *Dao) SearchChannelsInfo(c context.Context, mid int64, channelIDs []int64) (res map[int64]*channelgrpc.SearchChannelCard, err error) {
	var (
		args   = &channelgrpc.SearchChannelsInfoReq{Mid: mid, Cids: channelIDs}
		resTmp *channelgrpc.SearchChannelsInfoReply
	)
	if resTmp, err = d.grpcClient.SearchChannelsInfo(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[int64]*channelgrpc.SearchChannelCard)
	for _, channel := range resTmp.GetCards() {
		if channel == nil {
			continue
		}
		res[channel.Cid] = channel
	}
	return
}

func (d *Dao) RelativeChannel(c context.Context, mid int64, channelIDs []int64) (res []*channelgrpc.RelativeChannel, err error) {
	var (
		args   = &channelgrpc.RelativeChannelReq{Mid: mid, Cids: channelIDs}
		resTmp *channelgrpc.RelativeChannelReply
	)
	if resTmp, err = d.grpcClient.RelativeChannel(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCards()
	return
}

func (d *Dao) ChannelList(c context.Context, mid int64, ctype int32, offset string) (res *channelgrpc.ChannelListReply, err error) {
	var arg = &channelgrpc.ChannelListReq{Mid: mid, CategoryType: ctype, Offset: offset}
	if res, err = d.grpcClient.ChannelList(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) SearchChannelInHome(c context.Context, channelIDs []int64) (res *channelgrpc.SearchChannelInHomeReply, err error) {
	var args = &channelgrpc.SearchChannelInHomeReq{Cids: channelIDs}
	if res, err = d.grpcClient.SearchChannelInHome(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) ChannelInfos(ctx context.Context, channelIDs []int64) (map[int64]*channelgrpc.Channel, error) {
	reply, err := d.grpcClient.Infos(ctx, &channelgrpc.InfosReq{Cids: channelIDs})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.GetCidMap(), nil
}

func (d *Dao) ChannelFav(ctx context.Context, mid, ps int64, offset string) (*channelgrpc.SubChannelReply, error) {
	reply, err := d.grpcClient.SubChannel(ctx, &channelgrpc.SubChannelReq{Mid: mid, Ps: int32(ps), Offset: offset})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) ChannelDetail(ctx context.Context, arg *channelgrpc.ChannelDetailReq) (*channelgrpc.ChannelDetailReply, error) {
	return d.grpcClient.ChannelDetail(ctx, arg)
}

func (d *Dao) ChannelFeed(ctx context.Context, arg *baikegrpc.ChannelFeedReq) (*baikegrpc.ChannelFeedReply, error) {
	return d.baikeClient.ChannelFeed(ctx, arg)
}
