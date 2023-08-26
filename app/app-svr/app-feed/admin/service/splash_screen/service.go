package splash_screen

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/splash_screen"
)

type Service struct {
	dao *splash_screen.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: splash_screen.New(c),
	}

	// 定时更新失效配置的etime
	//nolint:errcheck,biligowordcheck
	go s.ETimeEditMonitor()

	return
}
