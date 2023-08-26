package service

import (
	"context"

	"go-common/library/conf/env"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-free/job/internal/dao"
	"go-gateway/app/app-svr/app-free/job/internal/model"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New)

// Service service.
type Service struct {
	ac             *paladin.Map
	dao            dao.Dao
	freeRecords    map[model.ISP][]*model.FreeRecord
	cron           *cron.Cron
	trackerRailGun *railgun.Railgun
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:   &paladin.TOML{},
		dao:  d,
		cron: cron.New(),
	}
	cf = s.Close
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}
	checkErr(s.loadRecordsCache())
	checkErr(s.cron.AddFunc(paladin.String(s.ac.Get("recordsConfigCron"), ""), func() {
		if err := s.loadRecordsCache(); err != nil {
			log.Error("%+v", err)
			return
		}
	}))
	s.cron.Start()
	if env.DeployEnv == env.DeployEnvProd || env.DeployEnv == env.DeployEnvPre {
		cfg := struct {
			TrackerGroup  *railgun.KafkaConfig
			TrackerSingle *railgun.SingleConfig
		}{}
		kc := paladin.Get("kafka.toml")
		if err := kc.UnmarshalTOML(&cfg); err != nil {
			panic(err)
		}
		s.initTrackerRailGun(cfg.TrackerGroup, cfg.TrackerSingle)
	}
	return
}

func (s *Service) loadRecordsCache() error {
	res, err := s.AllRecords(context.Background())
	if err != nil {
		return err
	}
	s.freeRecords = res
	return nil
}

func (s *Service) AllRecords(ctx context.Context) (rm map[model.ISP][]*model.FreeRecord, err error) {
	res, err := s.dao.AllFreeRecords(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	rm = make(map[model.ISP][]*model.FreeRecord, len(res))
	for _, r := range res {
		//0000-00-00 00:00:00
		if r.SuccessTime < 0 {
			r.SuccessTime = 0
		}
		if r.CancelTime < 0 {
			r.CancelTime = 0
		}
		if r.State != model.StateSucess {
			continue
		}
		rm[r.ISP] = append(rm[r.ISP], r)
	}
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.cron.Stop()
	if env.DeployEnv == env.DeployEnvProd || env.DeployEnv == env.DeployEnvPre {
		s.trackerRailGun.Close()
	}
}
