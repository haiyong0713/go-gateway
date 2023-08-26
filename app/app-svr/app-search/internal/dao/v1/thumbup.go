package v1

import (
	"context"

	"go-common/library/net/metadata"

	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

func (d *dao) HasLike(ctx context.Context, buvid string, mid int64, messageIDs []int64) (map[int64]thumbupgrpc.State, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	out := make(map[int64]thumbupgrpc.State)
	if mid > 0 {
		arg := &thumbupgrpc.HasLikeReq{
			Business:   "archive",
			MessageIds: messageIDs,
			Mid:        mid,
			IP:         ip,
		}
		reply, err := d.thumbupClient.HasLike(ctx, arg)
		if err != nil {
			return nil, err
		}
		for k, v := range reply.States {
			out[k] = v.State
		}
		return out, nil
	}
	arg := &thumbupgrpc.BuvidHasLikeReq{
		Business:   "archive",
		MessageIds: messageIDs,
		Buvid:      buvid,
		IP:         ip,
	}
	reply, err := d.thumbupClient.BuvidHasLike(ctx, arg)
	if err != nil {
		return nil, err
	}
	for k, v := range reply.States {
		out[k] = v.State
	}
	return out, nil
}
