package dynamic

import (
	"context"

	"go-common/library/log"
)

func (s *Service) load() {
	ctx := context.Background()
	list, err := s.rcmdDao.SchoolRcmd(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if err := s.dynDao.AddSchoolCache(ctx, list); err != nil {
		log.Error("%+v", err)
		return
	}
}
