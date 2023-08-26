package frontpage

import (
	"go-common/library/log"
	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
)

func (s *Service) GetMenus() (res []*model.Menu, err error) {
	if res, err = s.dao.GetMenus(); err != nil {
		log.Error("Service: GetMenus GetMenus error %v", err)
	}
	return
}
