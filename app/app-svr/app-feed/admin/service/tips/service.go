package tips

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/manager"
)

// Service is tianma service
type Service struct {
	dao *manager.Dao
}

// New new a tianma service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: manager.New(c),
	}

	//nolint:biligowordcheck
	go s.PublishMonitor()

	return
}
