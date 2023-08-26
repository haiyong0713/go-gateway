package dao

import (
	"context"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

const _typeVideo = 2

type favouriteDao struct {
	favourite favgrpc.FavoriteClient
}

func (d *favouriteDao) IsFavVideos(ctx context.Context, mid int64, aids []int64) (map[int64]int8, error) {
	reply, err := d.favourite.IsFavoreds(ctx, &favgrpc.IsFavoredsReq{Typ: _typeVideo, Mid: mid, Oids: aids})
	if err != nil {
		return nil, err
	}
	res := make(map[int64]int8)
	for k, v := range reply.Faveds {
		if v {
			res[k] = 1
		}
	}
	return res, nil
}
