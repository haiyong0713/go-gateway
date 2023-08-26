package service

import (
	"context"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/dao"

	"go-common/library/conf/paladin.v2"
	"go-common/library/sync/pipeline/fanout"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, new(*Service))

// Service service.
type Service struct {
	ac                    *paladin.Map
	dao                   dao.Dao
	renameDao             dao.RenameDao
	tusEditDao            dao.TusEditDao
	tusMultipleVersionDao dao.TusMultipleVersionDao
	fanout                *fanout.Fanout
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:                    &paladin.TOML{},
		dao:                   d,
		renameDao:             d.CreateRenameDao(),
		tusEditDao:            d.CreateTusEditDao(),
		tusMultipleVersionDao: d.CreateTusMultipleVersionDao(),
		fanout:                fanout.New("admin-fanout"),
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	if err != nil {
		return
	}
	s.InitConfigVersion()
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.fanout.Close()
	s.dao.Close()
}
