package service

import (
	pb "go-gateway/app/app-svr/distribution/distribution/api"
	"go-gateway/app/app-svr/distribution/distribution/internal/dao"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log/infoc.v2"
	"go-common/library/sync/pipeline/fanout"

	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.DistributionServer), new(*Service)))

type serviceConfig struct {
	PreferenceLogID string
	ExpConfigLogID  string
}

// Service service.
type Service struct {
	ac     *paladin.Map
	dao    dao.Dao
	infoc  infoc.Infoc
	cfg    serviceConfig
	fanout *fanout.Fanout
}

// New new a service and return.
func New(d dao.Dao) (*Service, func(), error) {
	s := &Service{
		ac:     &paladin.TOML{},
		dao:    d,
		fanout: fanout.New("service-fanout"),
	}
	close := s.Close
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	infocConfig := &infoc.Config{}
	if err := s.ac.Get("infoc").UnmarshalTOML(infocConfig); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	infoc_, err := infoc.New(infocConfig)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	s.infoc = infoc_

	cfg := serviceConfig{}
	if err := s.ac.Get("ServiceConfig").UnmarshalTOML(&cfg); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	s.cfg = cfg
	return s, close, nil
}

// Close close the resource.
func (s *Service) Close() {
	s.fanout.Close()
	s.infoc.Close()
	s.dao.Close()
}
