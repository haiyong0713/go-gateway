package service

import (
	"context"

	"go-common/library/log"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
)

func (s *Service) WebReplyWall(ctx context.Context, mid int64) (res *v1.GetReplyWallListResponse, err error) {
	arg := &v1.GetReplyWallListReq{Mid: mid}
	if res, err = s.esportsServiceClient.GetReplyWallList(ctx, arg); err != nil {
		log.Errorc(ctx, "ContestReplyWall s.esportsServiceClient.GetReplyWallList() mid(%d) error(%+v)", mid, err)
		return
	}
	if res.Contest == nil {
		res.Contest = &v1.ContestDetail{}
	}
	if res.ReplyList == nil {
		res.ReplyList = make([]*v1.ReplyWallInfo, 0)
	}
	return
}
