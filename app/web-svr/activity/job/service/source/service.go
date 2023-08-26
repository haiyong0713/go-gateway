package source

import (
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/like"
	rankdao "go-gateway/app/web-svr/activity/job/dao/rank_v2"

	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

// Service service
type Service struct {
	c                 *conf.Config
	dao               *like.Dao
	rankDao           rankdao.Dao
	arcClient         arcapi.ArchiveClient
	flowcontrolClient flowcontrolapi.FlowControlClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     like.New(c),
		rankDao: rankdao.New(c),
	}
	var err error
	if s.arcClient, err = arcapi.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.flowcontrolClient, err = flowcontrolapi.NewClient(c.FlowControlClient); err != nil {
		panic(err)
	}

	return s
}
