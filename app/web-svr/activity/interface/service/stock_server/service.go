package stockserver

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/stock"
)

// Service struct
type Service struct {
	c     *conf.Config
	dao   *stock.Dao
	cache *fanout.Fanout
}

var localS *Service

// New Service
func New(c *conf.Config) *Service {
	if localS != nil {
		return localS
	}
	s := &Service{
		c:     c,
		dao:   stock.New(c),
		cache: fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
	}
	return s
}
