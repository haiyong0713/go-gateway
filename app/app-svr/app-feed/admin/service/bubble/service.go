package bubble

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	bubbledao "go-gateway/app/app-svr/app-feed/admin/dao/bubble"
)

type Service struct {
	c         *conf.Config
	bubbleDao *bubbledao.Dao
	// cache
	cache *fanout.Fanout
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		bubbleDao: bubbledao.New(c),
		// cache
		cache: fanout.New("cache"),
	}
	return
}
