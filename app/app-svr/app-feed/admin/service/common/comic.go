package common

import (
	"context"

	"go-gateway/app/app-svr/app-feed/admin/model/comic"
)

func (s *Service) Comic(c context.Context, id int64) (data []*comic.ComicInfo, err error) {
	return s.comic.ComicInfo(c, id)
}
