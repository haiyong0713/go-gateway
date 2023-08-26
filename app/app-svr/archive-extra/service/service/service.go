package service

import (
	"go-gateway/app/app-svr/archive-extra/service/conf"
	"go-gateway/app/app-svr/archive-extra/service/dao"
)

// Service is
type Service struct {
	c *conf.Config
	d *dao.Dao

	authorisedCallers map[string]string
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		d: dao.New(c),
	}

	s.loadSteinsCallers()
	return
}

// Close resource.
func (s *Service) Close() {
	s.d.Close()
}

func (s *Service) loadSteinsCallers() {
	tmp := make(map[string]string)
	for k, v := range s.c.Custom.ExtraCallers {
		tmp[k] = v
	}
	s.authorisedCallers = tmp
}
