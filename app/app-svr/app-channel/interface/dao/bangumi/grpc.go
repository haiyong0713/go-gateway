package bangumi

import (
	"context"

	"go-common/library/log"

	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"github.com/pkg/errors"
)

func (d *Dao) CardsInfoReply(c context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	arg := &seasongrpc.SeasonInfoReq{SeasonIds: seasonIds}
	info, err := d.rpcClient.Cards(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

func (d *Dao) EpidsCardsInfoReply(c context.Context, episodeIds []int32) (res map[int32]*episodegrpc.EpisodeCardsProto, err error) {
	arg := &episodegrpc.EpReq{Epids: episodeIds}
	info, err := d.rpcEpidsClient.Cards(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

func (d *Dao) TagOGV(c context.Context, tagIDs []int64) (res map[int64]*appCardgrpc.SeasonCards, err error) {
	var (
		arg    = &appCardgrpc.TagCardsReq{TagIds: tagIDs}
		resTmp *appCardgrpc.TagCardsReply
	)
	if resTmp, err = d.appCardClient.TagCards(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetSeasonInfo()
	return
}
