package wechat

import (
	"go-gateway/app/web-svr/web-goblin/interface/conf"
	"go-gateway/app/web-svr/web-goblin/interface/dao/wechat"
)

// Service struct .
type Service struct {
	c   *conf.Config
	dao *wechat.Dao
}

// New init wechat service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: wechat.New(c),
	}
	return s
}
