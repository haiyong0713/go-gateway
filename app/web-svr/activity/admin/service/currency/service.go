package currency

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/currency"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *currency.Dao
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
		dao: currency.New(c),
	}
	return
}
