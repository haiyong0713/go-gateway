package service

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	railgunV2 "go-common/library/railgun.v2"
	pb "go-gateway/app/web-svr/space/job/api"
	"go-gateway/app/web-svr/space/job/internal/dao"

	archiveapi "git.bilibili.co/bapis/bapis-go/archive/service"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.DemoServer), new(*Service)))

type Config struct {
	Msg struct {
		SenderUID uint64
		NotifyMsg string
	}
	ArchiveClient *warden.ClientConfig
	Spec          struct {
		LivePlaybackWhitelist string
	}
	RailgunV2 struct {
		SpaceBinlog   string
		ArchiveNotify string
	}
}

// Service service.
type Service struct {
	ac          *Config
	archiveGRPC archiveapi.ArchiveClient
	// arcNotifyRailGun, spaceRailGun *railgun.Railgun
	dao           dao.Dao
	cron          *cron.Cron
	spaceConsumer railgunV2.Consumer
	arcConsumer   railgunV2.Consumer
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:   &Config{},
		dao:  d,
		cron: cron.New(),
	}
	cf = s.Close
	if err = paladin.Get("application.toml").UnmarshalTOML(&s.ac); err != nil {
		return nil, nil, err
	}
	if s.archiveGRPC, err = archiveapi.NewClient(s.ac.ArchiveClient); err != nil {
		return nil, nil, err
	}
	if err := s.dao.SetLivePlaybackWhitelist(context.Background()); err != nil {
		return nil, nil, err
	}
	// var (
	// 	dc          paladin.Map
	// 	dcfg, sdcfg *railgun.DatabusV1Config
	// 	pcfg, spcfg *railgun.SingleConfig
	// )
	// if err = paladin.Get("databus.toml").Unmarshal(&dc); err != nil {
	// 	return
	// }
	// if err = dc.Get("ArchiveNotifySub").UnmarshalTOML(&dcfg); err != nil {
	// 	return
	// }
	// if err = dc.Get("ArchiveNotifyRailgun").UnmarshalTOML(&pcfg); err != nil {
	// 	return
	// }

	// if err = dc.Get("SpaceSub").UnmarshalTOML(&sdcfg); err != nil {
	// 	return
	// }
	// if err = dc.Get("SpaceRailgun").UnmarshalTOML(&spcfg); err != nil {
	// 	return
	// }
	s.initRailGunTask()
	if err = s.initCron(); err != nil {
		return
	}
	return
}

func (s *Service) initRailGunTask() {
	s.initArcNotifyRailGun()
	s.initSpaceRailGun()
}

func (s *Service) initCron() error {
	if err := s.cron.AddFunc(s.ac.Spec.LivePlaybackWhitelist, func() {
		if err := s.dao.SetLivePlaybackWhitelist(context.Background()); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.cron.Stop()
	// s.arcNotifyRailGun.Close()
	// if s.spaceRailGun != nil {
	// 	s.spaceRailGun.Close()
	// }
	if s.spaceConsumer != nil {
		s.spaceConsumer.Close()
	}
	if s.arcConsumer != nil {
		s.arcConsumer.Close()
	}
}
