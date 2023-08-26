package reward_conf

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/reward_conf"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *reward_conf.Dao
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
		dao: reward_conf.New(c),
	}
	return
}
