package channel

import (
	"context"
	"fmt"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	channelmdl "git.bilibili.co/bapis/bapis-go/community/model/channel"
	"google.golang.org/grpc"
)

// Dao is archive dao.
type Dao struct {
	c          *conf.Config
	grpcClient channelgrpc.ChannelRPCClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	//grpc
	var err error
	clientSDK := conf.WardenSDKBuilder.Build("main.channel.channel-interface")
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(clientSDK.UnaryClientInterceptor()),
	}
	if d.grpcClient, err = channelgrpc.NewClient(c.ChannelGRPC, opts...); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) Tabs(c context.Context) (res []*channelgrpc.ChannelCategory, err error) {
	var tabs *channelgrpc.CategoryReply
	if tabs, err = d.grpcClient.Category(c, &channelgrpc.NoReq{}); err != nil {
		log.Error("%v", err)
		return
	}
	res = tabs.GetCategorys()
	return
}

func (d *Dao) ChannelList(c context.Context, mid int64, ctype int32, offset string) (res *channelgrpc.ChannelListReply, err error) {
	var arg = &channelgrpc.ChannelListReq{Mid: mid, CategoryType: ctype, Offset: offset}
	if res, err = d.grpcClient.ChannelList(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) SubscribedChannel(c context.Context, mid int64, isNewSub int32) (res *channelgrpc.SubscribeReply, err error) {
	arg := &channelgrpc.SubscribeReq{Mid: mid, SubVersion: isNewSub}
	if res, err = d.grpcClient.Subscribe(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) ChannelSort(c context.Context, mid int64, action int32, stick, normal string) (err error) {
	arg := &channelgrpc.UpdateSubscribeReq{Mid: mid, Action: action, Tops: stick, Cids: normal}
	if _, err = d.grpcClient.UpdateSubscribe(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) Recent(c context.Context, mid int64) (res []*channelgrpc.ChannelCard, err error) {
	arg := &channelgrpc.RecentChannelReq{Mid: mid}
	var recents *channelgrpc.RecentChannelReply
	if recents, err = d.grpcClient.RecentChannel(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = recents.GetCards()
	return
}

func (d *Dao) MyChannels(c context.Context, mid int64, offset string, isNoFeed bool, subVersion int32) (res *channelgrpc.MyChannelsReply, err error) {
	arg := &channelgrpc.MyChannelsReq{Mid: mid, Offset: offset, PageSize: 10, NotDefaultFeed: isNoFeed, SubVersion: subVersion}
	if res, err = d.grpcClient.MyChannels(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) Detail(c context.Context, mid, channelID int64, meta *channelmdl.MetaDataCtrl) (*channelgrpc.ChannelDetailReply, error) {
	arg := &channelgrpc.ChannelDetailReq{Mid: mid, ChannelId: channelID, Meta: meta}
	return d.grpcClient.ChannelDetail(c, arg)
}

func (d *Dao) ResourceList(c context.Context, arg *channelgrpc.ResourceListReq) (res *channelgrpc.ResourceListReply, err error) {
	if res, err = d.grpcClient.ResourceList(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) RankCards(c context.Context, channelID int64, offset string, ps int32) (res *channelgrpc.RankCardReply, err error) {
	var args = &channelgrpc.RankCardReq{ChannelId: channelID, Ps: ps, Offset: offset}
	if res, err = d.grpcClient.RankCard(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) Infos(c context.Context, channelIDs []int64, mid int64) (res map[int64]*channelgrpc.Channel, err error) {
	var (
		args  = &channelgrpc.InfosReq{Cids: channelIDs, Mid: mid}
		infos *channelgrpc.InfosReply
	)
	if infos, err = d.grpcClient.Infos(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = infos.GetCidMap()
	return
}

func (d *Dao) ScanedChannels(c context.Context, mid int64) (res []*channelgrpc.ChannelFeedCard, err error) {
	var (
		args   = &channelgrpc.ScanedChannelsReq{Mid: mid}
		resTmp *channelgrpc.ScanedChannelsReply
	)
	if resTmp, err = d.grpcClient.ScanedChannels(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetChannels()
	return
}

func (d *Dao) RedPoint(c context.Context, mid int64) (res bool, err error) {
	var (
		args   = &channelgrpc.NewNotifyReq{Mid: mid}
		resTmp *channelgrpc.NewNotifyReply
	)
	if resTmp, err = d.grpcClient.NewNotify(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetNotify()
	return
}

func (d *Dao) Rcmd(c context.Context, mid int64) (res []*channelgrpc.ChannelFeedCard, err error) {
	var (
		args   = &channelgrpc.AmazingChannelsReq{Mid: mid}
		resTmp *channelgrpc.AmazingChannelsReply
	)
	if resTmp, err = d.grpcClient.AmazingChannels(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetChannels()
	return
}

func (d *Dao) ViewChannel(c context.Context, mid int64) (res []*channelgrpc.ViewChannelCard, err error) {
	var (
		args   = &channelgrpc.ViewChannelReq{Mid: mid}
		resTmp *channelgrpc.ViewChannelReply
	)
	if resTmp, err = d.grpcClient.ViewChannel(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCard()
	return
}

func (d *Dao) HotChannel(c context.Context, mid int64, offset string) (res *channelgrpc.HotChannelReply, err error) {
	var args = &channelgrpc.HotChannelReq{Mid: mid, Offset: offset}
	if res, err = d.grpcClient.HotChannel(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) MyChannels2(c context.Context, mid int64, offset string, isNoFeed bool, subVersion int32) (res *channelgrpc.MyChannelsReply, err error) {
	arg := &channelgrpc.MyChannelsReq{Mid: mid, Offset: offset, PageSize: 10, NotDefaultFeed: isNoFeed, SubVersion: subVersion}
	if res, err = d.grpcClient.MyChannels2(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) RedPoint2(c context.Context, mid int64) (res bool, err error) {
	var (
		args   = &channelgrpc.NewNotifyReq{Mid: mid}
		resTmp *channelgrpc.NewNotifyReply
	)
	if resTmp, err = d.grpcClient.NewNotify2(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetNotify()
	return
}
