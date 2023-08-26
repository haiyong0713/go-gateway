package service

import (
	"context"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/jinzhu/gorm"

	"go-gateway/app/web-svr/native-page/admin/conf"
	"go-gateway/app/web-svr/native-page/admin/dao"
)

// Service biz service def.
type Service struct {
	c         *conf.Config
	dao       *dao.Dao
	DB        *gorm.DB
	accClient acccli.AccountClient
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
	}
	s.DB = s.dao.DB
	var err error
	if s.accClient, err = acccli.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	return s
}

// Ping check dao health.
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Wait wait all closed.
func (s *Service) Wait() {}

// Close close all dao.
func (s *Service) Close() {
	s.dao.Close()
}
