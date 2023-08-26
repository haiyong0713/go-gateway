package gameholiday

import (
	"git.bilibili.co/bapis/bapis-go/account/service"

	"go-gateway/app/web-svr/activity/interface/conf"
	gameholiday "go-gateway/app/web-svr/activity/interface/dao/game_holiday"
	"go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c           *conf.Config
	like        *like.Service
	gameHoliday gameholiday.Dao
	accClient   api.AccountClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		like:        like.New(c),
		gameHoliday: gameholiday.New(c),
	}
	var err error
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	return s
}

// Close ...
func (s *Service) Close() {
}
