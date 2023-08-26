package search

import (
	"context"

	"go-gateway/app/app-svr/app-feed/admin/model/manager"
)

func (s *Service) AppMogulLogList(ctx context.Context, param *manager.AppMogulLogParam) (*manager.AppMogulLogReply, error) {
	return s.managerDao.AppMogulLogList(ctx, param)
}
