package information

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	showdao "go-gateway/app/app-svr/app-feed/admin/dao/show"
)

// Service is search service
type Service struct {
	showDao *showdao.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao: showdao.New(c),
	}
	return
}
