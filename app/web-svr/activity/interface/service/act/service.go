package act

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	act "go-gateway/app/web-svr/activity/interface/dao/actplat"
)

// Service ...
type Service struct {
	c      *conf.Config
	actDao *act.Dao
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{}

	return s
}
