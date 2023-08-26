package ping

import (
	"context"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	pgdao "go-gateway/app/app-svr/app-resource/interface/dao/plugin"
)

type Service struct {
	pgDao *pgdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		pgDao: pgdao.New(c),
	}
	return
}

func (s *Service) Ping(c context.Context) (err error) {
	return s.pgDao.PingDB(c)
}
