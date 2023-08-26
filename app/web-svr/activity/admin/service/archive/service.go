package archive

import (
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/admin/conf"

	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

// Service ...
type Service struct {
	c                 *conf.Config
	ArcClient         arcapi.ArchiveClient
	flowcontrolClient flowcontrolapi.FlowControlClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	var err error
	if s.ArcClient, err = arcapi.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.flowcontrolClient, err = flowcontrolapi.NewClient(c.FlowControlClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {
}
