package college

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/college"
)

// CollegeAidInsertOrUpdate ...
func (s *Service) CollegeAidInsertOrUpdate(c context.Context, data *college.AIDList) (err error) {
	if data != nil {
		err := s.college.BacthInsertOrUpdateAidList(c, data)
		if err != nil {
			log.Errorc(c, "s.college.BacthInsertOrUpdateAidList")
			return err
		}
	}
	return nil
}
