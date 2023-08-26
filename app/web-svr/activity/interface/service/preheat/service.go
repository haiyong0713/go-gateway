package preheat

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/preheat"
)

type Service struct {
	c   *conf.Config
	dao *preheat.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: preheat.New(c),
	}
	return s
}
