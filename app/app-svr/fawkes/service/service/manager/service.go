package manager

import (
	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
)

// Service struct.
type Service struct {
	c     *conf.Config
	fkDao *fkdao.Dao
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		fkDao: fkdao.New(c),
	}
	return
}
