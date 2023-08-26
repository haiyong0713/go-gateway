package service

import (
	"context"

	"go-gateway/app/app-svr/newmont/service/api"
	"go-gateway/app/app-svr/newmont/service/conf"
	"go-gateway/app/app-svr/newmont/service/internal/dao"

	"go-common/library/conf/paladin.v2"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(api.NewmontServer), new(*Service)))

// Service service.
type Service struct {
	c          *conf.Config
	dao        dao.Dao
	sectionDao dao.SectionDao
	*sidebarIconLoader
	*sidebarLoader
	*grpcDep
	cron *cron.Cron
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		c:                 &conf.Config{},
		dao:               d,
		sectionDao:        d.CreateSectionDao(),
		cron:              cron.New(),
		sidebarLoader:     &sidebarLoader{},
		sidebarIconLoader: &sidebarIconLoader{},
	}
	cf = s.Close
	if err := paladin.Get("application.toml").UnmarshalTOML(s.c); err != nil {
		panic(err)
	}
	if err := paladin.Watch("application.toml", s.c); err != nil {
		panic(err)
	}
	s.sidebarLoader.load = s.loadSideBarCache
	s.sidebarIconLoader.load = s.loadIconCache
	s.grpcDep = initDep()
	s.StartLoad(s.sidebarLoader, s.sidebarIconLoader)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.cron.Stop()
}
