package datamanage

import (
	"context"
	"github.com/jinzhu/gorm"
	"go-common/library/database/orm"
	"go-gateway/app/web-svr/activity/admin/conf"
)

// Service biz service def.
type Service struct {
	c    *conf.Config
	DB   *gorm.DB
	TIDB *gorm.DB
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:    c,
		DB:   orm.NewMySQL(c.ORM),
		TIDB: orm.NewMySQL(c.TIDBORM),
	}
	return s
}

// Ping check dao health.
func (s *Service) Ping(c context.Context) (err error) {
	return nil
}

// Wait wait all closed.
func (s *Service) Wait() {}

// Close close all dao.
func (s *Service) Close() {
	s.DB.Close()
	s.TIDB.Close()
}
