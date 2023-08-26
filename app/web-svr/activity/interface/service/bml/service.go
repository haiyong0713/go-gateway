package bml

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bml"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
)

// Service struct
type Service struct {
	c      *conf.Config
	dao    *bml.Dao
	bwsdao *bwsonline.Dao
	cache  *fanout.Fanout
}

// New Service
func New(c *conf.Config) *Service {
	return &Service{
		c:      c,
		dao:    bml.New(c),
		bwsdao: bwsonline.New(c),
		cache:  fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
	}
}
