package search_whitelist

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/search_whitelist"
)

// Service is rank service
type Service struct {
	c   *conf.Config
	dao *search_whitelist.Dao
}

// New new a rank service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: search_whitelist.New(c),
	}

	//nolint:biligowordcheck
	go s.JobUpdateState()

	return
}
