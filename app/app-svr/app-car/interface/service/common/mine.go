package common

import (
	"context"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
)

func (s *Service) MineTabs(c context.Context, req *commonmdl.MineTabsReq) (resp []*commonmdl.MineTab) {
	for k, v := range commonmdl.Tabs {
		var isDefault bool
		if k == commonmdl.DefaultTab {
			isDefault = true
		}
		resp = append(resp, &commonmdl.MineTab{
			Type:      k,
			Name:      v,
			IsDefault: isDefault,
		})
	}
	return
}
