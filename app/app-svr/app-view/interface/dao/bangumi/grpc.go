package bangumi

import (
	"context"

	ogvgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"github.com/pkg/errors"
)

// CardsInfoReply pgc cards info
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

func (d *Dao) EpCardsInfo(c context.Context, req *ogvgrpc.EpCardsReq) (map[int32]*ogvgrpc.EpisodeCard, error) {
	res, err := d.ogvRpcClient.EpCards(c, req)
	if err != nil {
		return nil, err
	}
	return res.GetCards(), nil
}
