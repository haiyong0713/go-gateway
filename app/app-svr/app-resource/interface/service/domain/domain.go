package domain

import (
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/domain"
)

// Service domain service.
type Service struct {
	c *conf.Config

	// 	domains []string
}

// New new domain service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	return
}

// Domain get domain all
func (s *Service) Domain() (res *domain.Domain) {
	res = &domain.Domain{
		Domains:      s.c.Domain.Addr,
		ImageDomains: s.c.Domain.ImageAddr,
	}
	return
}
