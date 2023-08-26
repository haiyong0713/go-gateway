package taskv2

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
)

// TaskInsertOrUpdate ...
func (s *Service) TaskInsertOrUpdate(c context.Context, data *model.ActTask) (err error) {
	if data != nil {

		err := s.dao.TaskInsertOrUpdate(c, data)
		if err != nil {
			log.Errorc(c, "s.dao.TaskInsertOrUpdate")
			return err
		}
	}
	return nil
}
