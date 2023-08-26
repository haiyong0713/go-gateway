package service

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

// Service service.
type Service struct {
	ac               *paladin.Map
	dao              dao.Dao
	sendScheduleTask *railgun.Railgun
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	s.initBG()
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.sendScheduleTask.Close()
}
