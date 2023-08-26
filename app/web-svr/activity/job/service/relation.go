package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
)

func (s *Service) CallInternalSyncActRelationInfoDB2Cache() {
	ctx := context.Background()
	req := &api.InternalSyncActRelationInfoDB2CacheReq{
		From: "JOB",
	}
	if _, err := s.actGRPC.InternalSyncActRelationInfoDB2Cache(ctx, req); err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]CallInternalSyncActRelationInfoDB2Cache Err(%v)", err)
		return
	}
	return
}
