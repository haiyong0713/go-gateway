package service

import (
	"context"
	"sync"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	locdao "go-gateway/app/app-svr/app-free/admin/internal/dao/location"
	rcddao "go-gateway/app/app-svr/app-free/admin/internal/dao/record"
	"go-gateway/app/app-svr/app-free/admin/internal/model"
)

// Service service.
type Service struct {
	ac          *paladin.Map
	locDao      locdao.Dao
	rcdDao      rcddao.Dao
	freeRecords map[model.ISP][]*model.FreeRecord
	pcapResult  *sync.Map
}

// New new a service and return.
func New() (s *Service) {
	ac := new(paladin.TOML)
	if err := paladin.Watch("application.toml", ac); err != nil {
		panic(err)
	}
	s = &Service{
		ac:         ac,
		locDao:     locdao.New(),
		rcdDao:     rcddao.New(),
		pcapResult: &sync.Map{},
	}
	s.loadRecordsCache()
	// nolint:biligowordcheck
	go s.cacheproc()
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return
}

// Close close the resource.
func (s *Service) Close() {
}

func (s *Service) loadRecordsCache() {
	res, err := s.AllRecords(context.Background(), true)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.freeRecords = res
	s.pcapResult = &sync.Map{}
}

func (s *Service) cacheproc() {
	for {
		time.Sleep(time.Duration(paladin.Duration(s.ac.Get("tick"), time.Minute)))
		s.loadRecordsCache()
	}
}
