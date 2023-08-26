package service

import (
	"context"

	"go-common/library/conf/paladin.v2"
	pb "go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.BaasServer), new(*Service)))

// Service service.
type Service struct {
	ac     *paladin.Map
	dao    dao.Dao
	Common *CommonService
	Role   *RoleService
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:     &paladin.TOML{},
		dao:    d,
		Common: newOuterService(d),
		Role:   newRoleService(d),
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
