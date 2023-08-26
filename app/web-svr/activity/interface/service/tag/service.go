package tag

import (
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Service struct
type Service struct {
	c *conf.Config
}

// Close service
func (s *Service) Close() {

}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	return
}
