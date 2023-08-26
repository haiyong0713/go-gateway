package dao

import (
	"context"

	"go-common/library/log"

	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
)

func (d *Dao) TagOGV(c context.Context, tagIDs []int64) (res map[int64]*appCardgrpc.SeasonCards, err error) {
	var (
		arg    = &appCardgrpc.TagCardsReq{TagIds: tagIDs}
		resTmp *appCardgrpc.TagCardsReply
	)
	if resTmp, err = d.bangumiCardClient.TagCards(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetSeasonInfo()
	return
}

func (d *Dao) ShareMessage(c context.Context, epids []int32) ([]*pgcShareGrpc.ShareMessageResBody, error) {
	reply, err := d.pgcShareGRPC.QueryShareMessageInfo(c, &pgcShareGrpc.ShareMessageReq{EpId: epids})
	if err != nil {
		return nil, err
	}
	return reply.GetBodys(), nil
}
