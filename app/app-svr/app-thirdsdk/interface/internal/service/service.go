package service

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/dao"

	dm "git.bilibili.co/bapis/bapis-go/bilibili/community/service/dm/v1"
	arcpush "git.bilibili.co/bapis/bapis-go/manager/service/archive-push"

	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New)

// Service service.
type Service struct {
	ac            *paladin.Map
	dao           dao.Dao
	dmClient      dm.DMClient
	arcPushClient arcpush.ArchivePushClient
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	cf = s.Close
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		return
	}
	var cfg struct {
		DmGRPC      *warden.ClientConfig
		ArcPushGRPC *warden.ClientConfig
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dmClient, err = dm.NewClient(cfg.DmGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.arcPushClient, err = arcpush.NewClient(cfg.ArcPushGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

// Close close the resource.
func (s *Service) Close() {
}
