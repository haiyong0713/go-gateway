package cpc100

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/like"
)

// Service ...
type Service struct {
	c   *conf.Config
	dao *like.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: like.New(c),
	}
	return s
}

func (s *Service) Close() {

}
