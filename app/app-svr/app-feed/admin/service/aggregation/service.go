package aggregation

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/aggregation"
	"go-gateway/app/app-svr/app-feed/admin/dao/archive"
)

// Service is egg service
type Service struct {
	c      *conf.Config
	dao    *aggregation.Dao
	arcDao *archive.Dao
}

// New new a egg service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		dao:    aggregation.New(c),
		arcDao: archive.New(c),
	}
	return
}
