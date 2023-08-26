package region

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-job/job/conf"
	regiondao "go-gateway/app/app-svr/app-job/job/dao/region"
	"go-gateway/app/app-svr/app-job/job/model"

	"github.com/robfig/cron"
)

type Service struct {
	c                *conf.Config
	dao              *regiondao.Dao
	cron             *cron.Cron
	regionJobRailGun *railgun.Railgun
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:    c,
		dao:  regiondao.New(c),
		cron: cron.New(),
	}
	if model.EnvRun() {
		s.pub()
		// 间隔二分钟
		if err := s.cron.AddFunc("@every 2m", s.pub); err != nil {
			panic(err)
		}
		s.cron.Start()
	}
	s.initRailGun()
	s.cron.Start()
	return
}

// nolint:errcheck
func (s *Service) initRailGun() {
	if err := s.loadRegion(); err != nil {
		panic(fmt.Sprintf("loadRegion error:%+v", err))
	}
	if err := s.loadRegionlist(); err != nil {
		panic(fmt.Sprintf("loadRegionlist error:%+v", err))
	}
	if err := s.loadRegionListCache(); err != nil {
		panic(fmt.Sprintf("loadRegionListCache error:%+v", err))
	}
	r := railgun.NewRailGun("loadRegionCache", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "0 */3 * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadRegion()
		s.loadRegionlist()
		s.loadRegionListCache()
		return railgun.MsgPolicyNormal
	}))
	s.regionJobRailGun = r
	r.Start()
}

func (s *Service) pub() {
	const (
		_regionDefiniteStateOnline  = int8(1) // 定时上线
		_regionDefiniteStateOffline = int8(2) // 定时下线
		_stateOn                    = 1       // 启用
		_stateOff                   = 0       // 不启用
	)
	c := context.Background()
	now := time.Now()
	regions, err := s.dao.RegionDefinite(c)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(regions) == 0 {
		return
	}
	tx, err := s.dao.BeginTran(c)
	if err != nil {
		log.Error("pub s.dao.BeginTran %+v", err)
		return
	}
	for _, r := range regions {
		if now.Unix() < r.DefiniteTime.Unix() {
			continue
		}
		switch r.DefiniteState {
		case _regionDefiniteStateOnline:
			if err := s.dao.UpdateRegionStateSQL(tx, _stateOn, r.ID); err != nil {
				log.Error("%+v", err)
				_ = tx.Rollback()
				return
			}
		case _regionDefiniteStateOffline:
			if err := s.dao.UpdateRegionStateSQL(tx, _stateOff, r.ID); err != nil {
				log.Error("%+v", err)
				_ = tx.Rollback()
				return
			}
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("region publish success")
}

func (s *Service) Close() {
	if model.EnvRun() {
		s.cron.Stop()
	}
}
