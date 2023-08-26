package bangumi

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	arcmid "go-gateway/app/app-svr/archive/middleware"

	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"

	"github.com/pkg/errors"
)

func (d *Dao) CardsInfoReply(c context.Context, episodeIds []int32) (res map[int32]*episodegrpc.EpisodeCardsProto, err error) {
	arg := &episodegrpc.EpReq{Epids: episodeIds}
	info, err := d.rpcClient.Cards(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

func (d *Dao) EpCardsFromPgcByAids(c context.Context, aids []int64) (map[int64]*pgccard.EpisodeCard, error) {
	arg := &pgccard.EpCardsReq{
		Aid:      aids,
		BizScene: pgccard.BizScene_TIANMA_SMALL_CARD,
	}
	info, err := d.pgcCardClient.EpCards(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return info.AidCards, nil
}

func (d *Dao) EpCardsFromPgcByEpids(c context.Context, epid []int32) (map[int32]*pgccard.EpisodeCard, error) {
	arg := &pgccard.EpCardsReq{
		EpId:     epid,
		BizScene: pgccard.BizScene_TIANMA_SMALL_CARD,
	}
	info, err := d.pgcCardClient.EpCards(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return info.Cards, nil
}

func (d *Dao) CardsByAids(c context.Context, aids []int32) (res map[int32]*episodegrpc.EpisodeCardsProto, err error) {
	arg := &episodegrpc.EpAidReq{Aids: aids}
	info, err := d.rpcClient.CardsByAids(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

func (d *Dao) InlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int, mid int64, isStory, needShareCnt, needHe bool, buvid string, heInlineReq []*pgcinline.HeInlineReq) (map[int32]*pgcinline.EpisodeCard, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &pgcinline.EpReq{
		EpIds: epIDs,
		User: &pgcinline.UserReq{
			Mid:      mid,
			MobiApp:  mobiApp,
			Device:   device,
			Platform: platform,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Fnver:    uint32(batchArg.Fnver),
			Fnval:    uint32(batchArg.Fnval),
			Qn:       uint32(batchArg.Qn),
			Build:    int32(build),
			Fourk:    int32(batchArg.Fourk),
			NetType:  pgccard.NetworkType(batchArg.NetType),
			TfType:   pgccard.TFType(batchArg.TfType),
			Buvid:    buvid,
		},
		SceneControl: &pgcinline.SceneControl{
			WasStory: isStory,
		},
		CustomizeReq: &pgcinline.CustomizeReq{
			NeedShareCount: needShareCnt,
			NeedHe:         needHe,
		},
		HeInlineReq: heInlineReq,
	}
	info, err := d.pgcinlineClient.EpCard(c, arg)
	if err != nil {
		log.Error("pgc inline error(%v) arg(%v)", err, arg)
		return nil, err
	}
	return info.Infos, nil
}

func (d *Dao) StatusByMid(c context.Context, mid int64, SeasonIDs []int32) (map[int32]*pgcClient.FollowStatusProto, error) {
	rly, err := d.pgcFollowClient.StatusByMid(c, &pgcClient.FollowStatusByMidReq{Mid: mid, SeasonId: SeasonIDs})
	if err != nil {
		return nil, err
	}
	return rly.GetResult(), nil
}

func (d *Dao) BatchEpMaterial(ctx context.Context, param []*deliverygrpc.EpMaterialReq) (map[int64]*deliverygrpc.EpMaterial, error) {
	reply, err := d.deliveryClient.BatchEpMaterial(ctx, &deliverygrpc.BatchEpMaterialReq{
		Reqs: param,
	})
	if err != nil {
		return nil, err
	}
	return reply.GetMaterialMap(), nil
}
