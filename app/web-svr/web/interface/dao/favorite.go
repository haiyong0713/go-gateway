package dao

import (
	"context"

	"go-common/library/log"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

func (d *Dao) IsFavoreds(c context.Context, mid int64, oids []int64) (res map[int64]bool, err error) {
	var (
		args  = &favgrpc.IsFavoredsReq{Typ: 2, Mid: mid, Oids: oids}
		isFav *favgrpc.IsFavoredsReply
	)
	if isFav, err = d.favClient.IsFavoreds(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = isFav.GetFaveds()
	return
}
