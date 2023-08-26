package v1

import (
	"context"

	"go-common/library/log"

	baikegrpc "git.bilibili.co/bapis/bapis-go/community/interface/baike"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	mediagrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	reviewgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/review"

	"github.com/pkg/errors"
)

func (d *dao) Details(c context.Context, tids []int64) (res map[int64]*channelgrpc.ChannelCard, err error) {
	args := &channelgrpc.SimpleChannelDetailReq{Cids: tids}
	var details *channelgrpc.SimpleChannelDetailReply
	if details, err = d.channelClient.SimpleChannelDetail(c, args); err != nil {
		log.Error("%+v")
		return
	}
	res = details.GetChannelMap()
	return
}

func (d *dao) SearchChannel(c context.Context, mid int64, channelIDs []int64) (res map[int64]*channelgrpc.SearchChannel, err error) {
	var (
		args   = &channelgrpc.SearchChannelReq{Mid: mid, Cids: channelIDs}
		resTmp *channelgrpc.SearchChannelReply
	)
	if resTmp, err = d.channelClient.SearchChannel(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetChannelMap()
	return
}

func (d *dao) SearchChannelsInfo(c context.Context, mid int64, channelIDs []int64) (res map[int64]*channelgrpc.SearchChannelCard, err error) {
	var (
		args   = &channelgrpc.SearchChannelsInfoReq{Mid: mid, Cids: channelIDs}
		resTmp *channelgrpc.SearchChannelsInfoReply
	)
	if resTmp, err = d.channelClient.SearchChannelsInfo(c, args); err != nil {
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

func (d *dao) RelativeChannel(c context.Context, mid int64, channelIDs []int64) (res []*channelgrpc.RelativeChannel, err error) {
	var (
		args   = &channelgrpc.RelativeChannelReq{Mid: mid, Cids: channelIDs}
		resTmp *channelgrpc.RelativeChannelReply
	)
	if resTmp, err = d.channelClient.RelativeChannel(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCards()
	return
}

func (d *dao) ChannelList(c context.Context, mid int64, ctype int32, offset string) (res *channelgrpc.ChannelListReply, err error) {
	var arg = &channelgrpc.ChannelListReq{Mid: mid, CategoryType: ctype, Offset: offset}
	if res, err = d.channelClient.ChannelList(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *dao) SearchChannelInHome(c context.Context, channelIDs []int64) (res *channelgrpc.SearchChannelInHomeReply, err error) {
	var args = &channelgrpc.SearchChannelInHomeReq{Cids: channelIDs}
	if res, err = d.channelClient.SearchChannelInHome(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *dao) ChannelInfos(ctx context.Context, channelIDs []int64) (map[int64]*channelgrpc.Channel, error) {
	reply, err := d.channelClient.Infos(ctx, &channelgrpc.InfosReq{Cids: channelIDs})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.GetCidMap(), nil
}

func (d *dao) ChannelFav(ctx context.Context, mid, ps int64, offset string) (*channelgrpc.SubChannelReply, error) {
	reply, err := d.channelClient.SubChannel(ctx, &channelgrpc.SubChannelReq{Mid: mid, Ps: int32(ps), Offset: offset})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *dao) ChannelDetail(ctx context.Context, arg *channelgrpc.ChannelDetailReq) (*channelgrpc.ChannelDetailReply, error) {
	return d.channelClient.ChannelDetail(ctx, arg)
}

func (d *dao) ChannelFeed(ctx context.Context, arg *baikegrpc.ChannelFeedReq) (*baikegrpc.ChannelFeedReply, error) {
	return d.baikeClient.ChannelFeed(ctx, arg)
}

func (d *dao) GetMediaBizInfoByMediaBizId(ctx context.Context, mediaId int64) (*mediagrpc.MediaBizInfoGetReply, error) {
	reply, err := d.ogvMediaClient.GetMediaBizInfoByMediaBizId(ctx, &mediagrpc.MediaBizInfoGetReq{MediaBizId: mediaId})
	if err != nil {
		return nil, errors.Wrapf(err, "GetMediaBizInfoByMediaBizId mediaId=%d", mediaId)
	}
	return reply, nil
}

func (d *dao) GetMediaReviewInfo(ctx context.Context, mediaId int64) (*reviewgrpc.ReviewInfoReply, error) {
	reply, err := d.ogvReviewClient.ReviewInfo(ctx, &reviewgrpc.ReviewInfoReq{MediaId: mediaId})
	if err != nil {
		return nil, errors.Wrapf(err, "GetMediaReviewInfo mediaId=%d", mediaId)
	}
	return reply, nil
}

func (d *dao) GetMediaAllowReview(c context.Context, mediaBizId int32) (*reviewgrpc.AllowReviewReply, error) {
	return d.ogvReviewClient.AllowReview(c, &reviewgrpc.AllowReviewReq{MediaId: mediaBizId})
}
