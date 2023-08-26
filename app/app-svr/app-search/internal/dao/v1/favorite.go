package v1

import (
	"context"

	favclient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

func (d *dao) IsFavVideos(ctx context.Context, mid int64, aids []int64) (map[int64]int8, error) {
	const _typeVideo = 2

	reply, err := d.favClient.IsFavoreds(ctx, &favclient.IsFavoredsReq{
		Typ:  _typeVideo,
		Mid:  mid,
		Oids: aids,
	})
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
