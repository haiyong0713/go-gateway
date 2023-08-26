package dao

import (
	"context"
	"go-common/library/net/metadata"

	api "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

type thumbupDao struct {
	thumbup api.ThumbupClient
}

// HasLike user has like
func (d *thumbupDao) HasLike(ctx context.Context, buvid string, mid int64, messageIDs []int64) (map[int64]int8, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	out := make(map[int64]int8)
	if mid > 0 {
		arg := &api.HasLikeReq{
			Business:   "archive",
			MessageIds: messageIDs,
			Mid:        mid,
			IP:         ip,
		}
		reply, err := d.thumbup.HasLike(ctx, arg)
		if err != nil {
			return nil, err
		}
		for k, v := range reply.States {
			out[k] = int8(v.State)
		}
		return out, nil
	}
	arg := &api.BuvidHasLikeReq{
		Business:   "archive",
		MessageIds: messageIDs,
		Buvid:      buvid,
		IP:         ip,
	}
	reply, err := d.thumbup.BuvidHasLike(ctx, arg)
	if err != nil {
		return nil, err
	}
	for k, v := range reply.States {
		out[k] = int8(v.State)
	}
	return out, nil
}
