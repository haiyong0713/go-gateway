package pgc

import (
	"context"
	"fmt"
	"math"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/util"

	inlinegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	epgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// CardsInfoReply pgc grpc
func (d *Dao) CardsInfoReply2(c context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	// ogv侧season.Cards接口要求每次请求的id最多50个，否则报400错误（/pgc.service.season.season.v1.Season/Cards）
	const MAX_LEN = 50
	var (
		tmp map[int32]*seasongrpc.CardInfoProto
	)

	pages := int(math.Ceil(float64(len(seasonIds)) / float64(MAX_LEN)))
	for pn := 1; pn <= pages; pn++ {
		start, end := util.PaginateSlice(pn, MAX_LEN, len(seasonIds))
		if tmp, err = d.CardsInfoReply(c, seasonIds[start:end]); err != nil {
			return
		}
		res = d.mergeCards(res, tmp)
	}
	return
}

func (d *Dao) mergeCards(a map[int32]*seasongrpc.CardInfoProto, b map[int32]*seasongrpc.CardInfoProto) (ret map[int32]*seasongrpc.CardInfoProto) {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	ret = a
	for k, v := range b {
		ret[k] = v
	}
	return
}

// CardsInfoReply pgc grpc
func (d *Dao) CardsInfoReply(c context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {

	arg := &seasongrpc.SeasonInfoReq{
		SeasonIds: seasonIds,
	}
	info, err := d.rpcClient.Cards(c, arg)
	if err != nil {
		err = fmt.Errorf(util.ErrorRpcFmts, err.Error(), d.userFeed.Pgc, "SeasonClient.Cards")
		log.Error("CardsInfoReply req(%v) error(%v)", seasonIds, err)
		return nil, err
	}
	if len(info.Cards) == 0 {
		err = fmt.Errorf(util.ErrorNullFmts, util.ErrorDataNull, d.userFeed.Pgc, "SeasonClient.Cards")
		log.Error("CardsInfoReply req(%v) res(%v)", seasonIds, info)
		return nil, err
	}
	res = info.Cards
	return
}

// CardsEpInfoReply get pgc ep cards values by epid
func (d *Dao) CardsEpInfoReply(c context.Context, epIds []int32) (res map[int32]*epgrpc.EpisodeCardsProto, err error) {
	var epInfo *epgrpc.EpisodeCardsReply
	arg := &epgrpc.EpReq{
		Epids:   epIds,
		NeedAll: 1,
	}
	epInfo, err = d.epClient.Cards(c, arg)
	if err != nil {
		err = fmt.Errorf(util.ErrorRpcFmts, err.Error(), d.userFeed.Pgc, "EpisodeClient.Cards")
		log.Error("CardsEpInfoReply req(%v) error(%v)", epIds, err)
		return nil, err
	}
	if epInfo == nil || len(epInfo.Cards) == 0 {
		return nil, fmt.Errorf("无效EpID(%v)"+util.ErrorPersonFmt, epIds, d.userFeed.Pgc)
	}
	res = epInfo.Cards
	return
}

func (d *Dao) InlineCardEpInfoReply(c context.Context, epIds []int32) (res map[int32]*inlinegrpc.EpisodeCard, err error) {
	var epInfo *inlinegrpc.EpisodeCardReply
	arg := &inlinegrpc.EpReq{
		EpIds: epIds,
	}
	if epInfo, err = d.inlineClient.EpCard(c, arg); err != nil {
		err = fmt.Errorf(util.ErrorRpcFmts, err.Error(), d.userFeed.Pgc, "EpisodeClient.Cards")
		log.Error("CardsEpInfoReply req(%v) error(%v)", epIds, err)
		return nil, err
	}
	if epInfo == nil || len(epInfo.Infos) == 0 {
		return nil, fmt.Errorf("无效EpID(%v)"+util.ErrorPersonFmt, epIds, d.userFeed.Pgc)
	}
	res = epInfo.Infos
	return
}
