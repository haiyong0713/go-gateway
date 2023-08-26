package fit

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/fit"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *fit.Dao
}

// Close service
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: fit.New(c),
	}
	return
}
