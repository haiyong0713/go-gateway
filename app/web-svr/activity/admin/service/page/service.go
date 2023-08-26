package page

import "go-gateway/app/web-svr/activity/admin/conf"

type Service struct {
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{}
	return
}
