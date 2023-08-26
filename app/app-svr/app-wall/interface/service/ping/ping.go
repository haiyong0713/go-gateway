package ping

import (
	"context"

	"go-gateway/app/app-svr/app-wall/interface/conf"
	walldao "go-gateway/app/app-svr/app-wall/interface/dao/wall"
)

type Service struct {
	wallDao *walldao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		wallDao: walldao.New(c),
	}
	return
}

// Ping is check server ping.
func (s *Service) Ping(c context.Context) (err error) {
	return s.wallDao.Ping(c)
}
