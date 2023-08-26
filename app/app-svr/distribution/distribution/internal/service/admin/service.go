package service

import (
	originservice "go-gateway/app/app-svr/distribution/distribution/internal/service"

	"go-common/library/conf/paladin.v2"

	safecenter "git.bilibili.co/bapis/bapis-go/passport/service/safecenter"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

type Service struct {
	ac         *paladin.Map
	origin     *originservice.Service
	safecenter safecenter.SafecenterClient
}

var Provider = wire.NewSet(New, originservice.New)

// New new a service and return.
func New(origin *originservice.Service) (*Service, func(), error) {
	s := &Service{
		ac:     &paladin.TOML{},
		origin: origin,
	}
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	safecenter, err := safecenter.NewClient(nil)
	if err != nil {
		return nil, nil, err
	}
	s.safecenter = safecenter
	return s, s.Close, nil
}

// Close close the resource.
func (s *Service) Close() {}
