package display

import (
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	locdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/location"
)

// Service is zone service.
type Service struct {
	// ip
	loc *locdao.Dao
}

// New initial display service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		loc: locdao.New(c),
	}
	return
}
