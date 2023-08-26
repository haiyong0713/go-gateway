package ping

import (
	"context"

	"go-gateway/app/app-svr/app-admin/admin/conf"
	pgdao "go-gateway/app/app-svr/app-admin/admin/dao/audit"
)

// Service dao
type Service struct {
	pgDao *pgdao.Dao
}

// New init
func New(c *conf.Config) (s *Service) {
	s = &Service{
		pgDao: pgdao.New(c),
	}
	return
}

// Ping ping
func (s *Service) Ping(c context.Context) (err error) {
	return s.pgDao.PingDB(c)
}
