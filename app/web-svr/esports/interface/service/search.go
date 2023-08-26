package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/pkg/idsafe/bvid"
)

// Search  search video list.
func (s *Service) Search(c context.Context, mid int64, p *model.ParamSearch, buvid string) (rs *model.SearchEsp, err error) {
	if rs, err = s.dao.Search(c, mid, p, buvid); err != nil {
		return
	}
	for _, v := range rs.Result {
		if v.Bvid, err = bvid.AvToBv(v.ID); err != nil {
			log.Error("Search AvToBv id(%d) error(%v)", v.ID, err)
			err = nil
			continue
		}
	}
	return
}
