package service

import (
	"context"

	"go-common/library/conf/paladin"
	"go-gateway/app/web-svr/web-goblin/admin/internal/dao"
)

// Service service.
type Service struct {
	ac  *paladin.Map
	dao dao.Dao
}

// New new a service and return.
func New(d dao.Dao) (s *Service, err error) {
	s = &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	err = paladin.Watch("application.toml", s.ac)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return nil
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}
