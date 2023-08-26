package service

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/space/service/api"
	"go-gateway/app/web-svr/space/service/internal/dao"

	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.SpaceServer), new(*Service)))

type Config struct {
	Spec struct {
		LivePlaybackWhitelist string
	}
}

// Service service.
type Service struct {
	ac                    *Config
	dao                   dao.Dao
	cron                  *cron.Cron
	livePlaybackWhitelist map[int64]struct{}
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:   &Config{},
		dao:  d,
		cron: cron.New(),
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&s.ac); err != nil {
		return nil, nil, err
	}
	if s.livePlaybackWhitelist, err = s.dao.CacheLivePlaybackWhitelist(context.Background()); err != nil {
		return
	}
	s.initCron()
	cf = s.Close
	return
}

// Close close the resource.
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}

func (s *Service) initCron() {
	s.cron.AddFunc(s.ac.Spec.LivePlaybackWhitelist, func() {
		res, err := s.dao.CacheLivePlaybackWhitelist(context.Background())
		if err != nil {
			log.Error("%+v", err)
			return
		}
		s.livePlaybackWhitelist = res
	})
	s.cron.Start()
}
