package domain

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/domain"
)

const (
	DomainStatusInit = 0
	DomainStatusSync = 1
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *domain.Dao
	// chan
	cache *fanout.Fanout
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: domain.New(c),
	}
	return
}
