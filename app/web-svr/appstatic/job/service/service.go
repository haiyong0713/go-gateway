package service

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-gateway/app/web-svr/appstatic/job/conf"
	"go-gateway/app/web-svr/appstatic/job/dao/caldiff"
	"go-gateway/app/web-svr/appstatic/job/dao/push"
)

var ctx = context.Background()

// Service .
type Service struct {
	c         *conf.Config
	dao       *caldiff.Dao
	pushDao   *push.Dao
	waiter    *sync.WaitGroup
	daoClosed bool
}

// New creates a Service instance.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     caldiff.New(c),
		pushDao: push.New(c),
		waiter:  new(sync.WaitGroup),
	}
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.calDiffproc()
	return
}

// Close releases resources which owned by the Service instance.
func (s *Service) Close() (err error) {
	log.Info("Close dao!")
	s.daoClosed = true
	log.Info("Wait waiter!")
	s.waiter.Wait()
	log.Info("appstatic-job has been closed.")
	return
}
