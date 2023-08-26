package aggregation_v2

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	aggdao2 "go-gateway/app/app-svr/app-feed/admin/dao/aggregation_v2"
	"go-gateway/app/app-svr/app-feed/admin/dao/archive"
)

type Service struct {
	c       *conf.Config
	aggDao2 *aggdao2.Dao

	// 	aggCh  chan func()
	arcDao *archive.Dao

	// 	cache *fanout.Fanout
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		aggDao2: aggdao2.New(c),
		arcDao:  archive.New(c),
	}
	return
}
