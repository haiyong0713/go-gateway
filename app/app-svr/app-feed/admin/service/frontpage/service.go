package frontpage

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/frontpage"
	"go-gateway/app/app-svr/app-feed/admin/dao/location"
)

// Service is tianma service
type Service struct {
	dao         *frontpage.Dao
	locationDAO *location.Dao
	c           *conf.Config
}

// New new a tianma service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:         frontpage.New(c),
		locationDAO: location.New(c),
		c:           c,
	}

	return
}
