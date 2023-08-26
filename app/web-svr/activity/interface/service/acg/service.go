package acg

import (
	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Service ...
type Service struct {
	c     *conf.Config
	redis *redis.Pool
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		redis: redis.NewPool(c.Redis.Config),
	}
	return s
}

// Close ...
func (s *Service) Close() {
}
