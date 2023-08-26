package preheat

import (
	"context"

	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/interface/model/preheat"
)

func (s *Service) DownloadInfo(c context.Context, ID int64) (*preheat.DownInfo, error) {
	rs, err := s.dao.GetByID(c, ID)
	if err != nil {
		return nil, err
	}
	if rs == nil || rs.ID == 0 {
		err = xecode.NothingFound
		return nil, err
	}
	return rs, nil
}
