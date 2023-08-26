package service

import (
	"context"
	"sync"

	"go-common/library/conf/paladin.v2"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	pb "go-gateway/app/app-svr/up-archive/job/api"
	"go-gateway/app/app-svr/up-archive/job/internal/dao"

	"github.com/BurntSushi/toml"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.UpArchiveJobServer), new(*Service)))

type Config struct {
	ArchiveCfg, StaffCfg, UpArcCfg, ArcFlowControlCfg *railgun.SingleConfig
	BuildArchiveList                                  struct {
		Switch  bool
		LastAid int64
		Limit   int64
	}
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	*c = tmp
	return nil
}

// Service service.
type Service struct {
	dao                                                               dao.Dao
	archiveRailGun, staffRailGun, upArcRailGun, arcFlowControlRailGun *railgun.Railgun
	ac                                                                *Config
	waiter                                                            sync.WaitGroup
	aidsChan                                                          chan []int64
	closed                                                            bool
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		dao:      d,
		ac:       &Config{},
		aidsChan: make(chan []int64, 1),
	}
	cf = s.Close
	var (
		dc                                                paladin.Map
		arcNotifySub, arcSub, upArcSub, arcFlowControlSub *databus.Config
	)
	if err = paladin.Get("databus.toml").Unmarshal(&dc); err != nil {
		return
	}
	if err = dc.Get("ArchiveNotifySub").UnmarshalTOML(&arcNotifySub); err != nil {
		return
	}
	if err = dc.Get("ArchiveSub").UnmarshalTOML(&arcSub); err != nil {
		return
	}
	if err = dc.Get("UpArchiveSub").UnmarshalTOML(&upArcSub); err != nil {
		return
	}
	if err = dc.Get("ArchiveFlowControlSub").UnmarshalTOML(&arcFlowControlSub); err != nil {
		return
	}
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		return
	}
	s.initArchiveRailGun(&railgun.DatabusV1Config{Config: arcNotifySub}, s.ac.ArchiveCfg)
	s.initStaffRailGun(&railgun.DatabusV1Config{Config: arcSub}, s.ac.StaffCfg)
	s.initUpArcRailGun(&railgun.DatabusV1Config{Config: upArcSub}, s.ac.UpArcCfg)
	s.initArchiveFlowControlRailGun(&railgun.DatabusV1Config{Config: arcFlowControlSub}, s.ac.ArcFlowControlCfg)
	for i := 0; i < 4; i++ {
		s.waiter.Add(1)
		go s.buildArc()
	}
	s.waiter.Add(1)
	go s.initArcList()
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.closed = true
	s.archiveRailGun.Close()
	s.staffRailGun.Close()
	s.upArcRailGun.Close()
	s.arcFlowControlRailGun.Close()
	s.waiter.Wait()
}
