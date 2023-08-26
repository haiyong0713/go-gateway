package page

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/page"
)

type Service struct {
	c   *conf.Config
	dao *page.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: page.New(c),
	}
	return s
}
