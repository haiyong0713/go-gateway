package service

import (
	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/archive-extra/job/conf"
	"go-gateway/app/app-svr/archive-extra/job/dao"
)

// Service is
type Service struct {
	c               *conf.Config
	d               *dao.Dao
	redis           *redis.Redis
	ArchiveExtraRgs []*Railgun
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		d:     dao.New(c),
		redis: redis.NewRedis(c.Redis),
	}
	s.initArchiveExtraRg()
	return
}
