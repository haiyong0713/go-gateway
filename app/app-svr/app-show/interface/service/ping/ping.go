package ping

import (
	"context"

	"go-gateway/app/app-svr/app-show/interface/conf"
	showdao "go-gateway/app/app-svr/app-show/interface/dao/show"
)

type Service struct {
	showDao *showdao.Dao
	// bnDao   *bndao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao: showdao.New(c),
	}
	return
}

// Ping is check server ping.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.showDao.Ping(c); err != nil {
		return
	}
	return
}
