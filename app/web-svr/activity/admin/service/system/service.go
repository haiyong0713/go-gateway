package system

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/system"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *system.Dao
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
		dao: system.New(c),
	}
	return
}
