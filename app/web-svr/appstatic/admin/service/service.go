package service

import (
	"context"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/appstatic/admin/conf"
	"go-gateway/app/web-svr/appstatic/admin/dao"

	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
)

// Service biz service def.
type Service struct {
	c              *conf.Config
	dao            *dao.Dao
	DB             *gorm.DB
	waiter         *sync.WaitGroup
	daoClosed      bool  // logic close the dao's DB
	MaxSize        int64 // max size supported for the file to upload
	cache          *fanout.Fanout
	BigfileTimeout time.Duration
	cron           *cron.Cron
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:              c,
		dao:            dao.New(c),
		daoClosed:      false,
		waiter:         new(sync.WaitGroup),
		cache:          fanout.New("cache"),
		BigfileTimeout: time.Duration(c.Cfg.BigfileTimeout),
		cron:           cron.New(),
	}
	s.DB = s.dao.DB
	s.loadPackageInfoForAppView()
	if err := s.cron.AddFunc("@every 5s", s.loadPackageInfoForAppView); err != nil {
		panic(err)
	}
	s.cron.Start()
	return s
}

// Ping check dao health.
func (s *Service) Ping(c context.Context) (err error) {
	return
}

// Wait wait all closed.
func (s *Service) Wait() {
	if s.dao != nil {
		s.daoClosed = true
		log.Info("Dao is logically closed!")
	}
	log.Info("Wait waiter!")
	s.waiter.Wait()
}

// Close close all dao.
func (s *Service) Close() {
	log.Info("Close Dao physically!")
	s.dao.Close()
	log.Info("Service Closed!")
	s.cron.Stop()
}
