package account

import (
	"git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/admin/conf"
)

// Service ...
type Service struct {
	c         *conf.Config
	AccClient api.AccountClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	var err error
	if s.AccClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {

}
