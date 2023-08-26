package service

import (
	"context"

	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) AppAudit(c context.Context, arg *api.NoArgRequest) (res *api.AuditReply, err error) {
	list, err := s.show.AppAudits(c)
	if err != nil {
		log.Error("%+v", err)
	}
	res = &api.AuditReply{List: list}
	return
}
