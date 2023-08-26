package service

import (
	"go-gateway/app/app-svr/ugc-season/service/conf"
	"go-gateway/app/app-svr/ugc-season/service/dao"
)

// Service is
type Service struct {
	c *conf.Config
	d *dao.Dao
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		d: dao.New(c),
	}
	return
}

// Close resource.
func (s *Service) Close() {
	s.d.Close()
}
