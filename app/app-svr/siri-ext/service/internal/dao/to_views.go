package dao

import (
	"context"
	"go-common/library/log"

	toviewsgrpc "git.bilibili.co/bapis/bapis-go/community/service/toview"
)

func (d *dao) UserToViewsIsEmpty(ctx context.Context, mid int64) bool {
	reply, err := d.toviews.UserToViews(ctx, &toviewsgrpc.UserToViewsReq{
		Mid:        mid,
		BusinessId: 1,
		Pn:         1,
		Ps:         10,
	})
	if err != nil {
		log.Error("Failed to get user to views: %d: %+v", mid, err)
		return true // 认为稍后再看就是空的
	}
	return reply.Count == 0
}
