package handwrite

import (
	"git.bilibili.co/bapis/bapis-go/account/service"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/handwrite"
	"go-gateway/app/web-svr/activity/interface/dao/rank"
	"go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c         *conf.Config
	handwrite handwrite.Dao
	rank      rank.Dao
	accClient api.AccountClient
	like      *like.Service
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		handwrite: handwrite.New(c),
		rank:      rank.New(c),
		like:      like.New(c),
	}
	var err error
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {
	s.handwrite.Close()
	s.rank.Close()
}
