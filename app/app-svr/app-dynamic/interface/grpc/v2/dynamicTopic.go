package v2

import (
	"context"

	"go-common/library/ecode"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

func (s *Server) LegacyTopicFeed(ctx context.Context, req *api.LegacyTopicFeedReq) (*api.LegacyTopicFeedReply, error) {
	if req.TopicId <= 0 && len(req.TopicName) <= 0 {
		return nil, ecode.RequestErr
	}
	return s.dynSvr.LegacyTopicFeed(ctx, mdlv2.NewGeneralParamFromCtx(ctx), req)
}
