package fit

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/favorite"
	"go-gateway/app/web-svr/activity/interface/dao/fit"
	"go-gateway/app/web-svr/activity/interface/service/archive"
)

// Service ...
type Service struct {
	c       *conf.Config
	fitDao  fit.Dao
	favDao  *favorite.Dao
	archive *archive.Service
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		fitDao:  fit.New(c),
		favDao:  favorite.New(c),
		archive: archive.New(c),
	}
	return s
}
