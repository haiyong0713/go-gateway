package service

import (
	"context"

	"go-gateway/app/web-svr/space/interface/model"
)

// ShopInfo get shop info.
func (s *Service) ShopInfo(c context.Context, mid int64) (data *model.ShopInfo, err error) {
	return s.dao.ShopInfo(c, mid)
}
