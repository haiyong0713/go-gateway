package ping

import (
	"context"

	"go-gateway/app/app-svr/app-feed/interface/conf"
)

type Service struct {
}

func New(c *conf.Config) (s *Service) {
	s = &Service{}
	return
}

// Ping is check server ping.
func (s *Service) Ping(c context.Context) (err error) {
	return
}
