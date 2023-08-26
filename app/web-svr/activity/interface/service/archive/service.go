package archive

import (
	"go-gateway/app/web-svr/activity/interface/conf"

	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

// Service ...
type Service struct {
	c                 *conf.Config
	flowcontrolClient flowcontrolapi.FlowControlClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	var err error

	if s.flowcontrolClient, err = flowcontrolapi.NewClient(c.FlowControlClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {
}
