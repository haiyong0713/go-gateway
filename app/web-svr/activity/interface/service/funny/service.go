package funny

import (
	"git.bilibili.co/bapis/bapis-go/account/service"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/funny"
	"go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c         *conf.Config
	funny     funny.Dao
	like      *like.Service
	accClient api.AccountClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		funny: funny.New(c),
		like:  like.New(c),
	}

	var err error
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	return s
}

// Close ...
func (s *Service) Close() {

}
