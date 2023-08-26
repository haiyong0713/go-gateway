package show

import (
	"context"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/banner"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
)

func (s *Service) Banner(c context.Context, mid int64, plat int8, buvid string, param *banner.ShowBannerParam) []cardm.Handler {
	return s.feed(c, plat, mid, buvid, param, model.BannerV1)
}
