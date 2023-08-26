package taskv2

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/taskv2"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *taskv2.Dao
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
		dao: taskv2.New(c),
	}
	return
}
