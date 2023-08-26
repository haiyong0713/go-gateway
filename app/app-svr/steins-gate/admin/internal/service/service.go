package service

import (
	"go-gateway/app/app-svr/steins-gate/admin/conf"
	"go-gateway/app/app-svr/steins-gate/admin/internal/dao"
)

// Service service.
type Service struct {
	c   *conf.Config
	dao *dao.Dao
}

// New new a service and return.n
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
	}
	return s
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()

}
