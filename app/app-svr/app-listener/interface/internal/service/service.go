package service

import (
	"context"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, wire.Bind(new(v1.ListenerServer), new(*Service)), wire.Bind(new(v1.MusicServer), new(*Service)))

// Service service.
type Service struct {
	C   *conf.AppConfig
	dao dao.Dao
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		C:   &conf.AppConfig{},
		dao: d,
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.C)
	conf.C = s.C
	return
}

func (s *Service) Ping(_ context.Context, _ *empty.Empty) (e *empty.Empty, err error) {
	return &empty.Empty{}, nil
}

// Close close the resource.
func (s *Service) Close() {
}
