package web

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-common/library/sync/pipeline/fanout"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	dymdl "go-gateway/app/web-svr/dynamic/service/model"
	dyrpc "go-gateway/app/web-svr/dynamic/service/rpc/client"
	"go-gateway/app/web-svr/web-goblin/job/conf"
	mdlweb "go-gateway/app/web-svr/web-goblin/job/dao/web"

	"github.com/robfig/cron"
)

// Service struct .
type Service struct {
	c                *conf.Config
	dao              *mdlweb.Dao
	dy               *dyrpc.Service
	archiveClient    arcapi.ArchiveClient
	arcNotifyRailgun *railgun.Railgun
	outArcRailgun    *railgun.Railgun
	arcTypes         map[int32]*arcapi.Tp
	cron             *cron.Cron
	cache            *fanout.Fanout
}

// New init .
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   mdlweb.New(c),
		dy:    dyrpc.New(c.DynamicRPC),
		cron:  cron.New(),
		cache: fanout.New("cache"),
	}
	var err error
	if s.archiveClient, err = arcapi.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	s.initArcNotifyRailGun(c.ArchiveNotifySub, c.ArchiveNotifyRailgun)
	s.initOutArcRailGun(c.OutArcSub, c.OutArcRailgun)
	// nolint:biligowordcheck
	go s.broadcastDy()
	s.initCron()
	return s
}

func (s *Service) initCron() {
	s.loadArcTypes()
	if err := s.cron.AddFunc(s.c.Cron.LoadArcTypes, s.loadArcTypes); err != nil {
		panic(err)
	}
	if s.c.Cron.LoadChangeOurArc != "" {
		if err := s.cron.AddFunc(s.c.Cron.LoadChangeOurArc, s.loadChangeOurArc); err != nil {
			panic(err)
		}
	}
	s.cron.Start()
}

// Ping Service .
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close Service .
func (s *Service) Close() {
	s.arcNotifyRailgun.Close()
	s.outArcRailgun.Close()
	s.dao.Close()
}

func (s *Service) broadcastDy() {
	var (
		dynamics map[string]int
		err      error
		b        []byte
	)
	for {
		if dynamics, err = s.dy.RegionTotal(context.Background(), &dymdl.ArgRegionTotal{RealIP: ""}); err != nil {
			mdlweb.PromError("RegionTotal接口错误", "s.dy.RegionTotal error(%v)", err)
			time.Sleep(time.Second)
			continue
		}
		if b, err = json.Marshal(dynamics); err != nil {
			log.Error("broadcastDy json.Marshal error(%v)", err)
			return
		}
		if err = s.dao.PushAll(context.Background(), string(b), ""); err != nil {
			log.Error("s.dao.PushAll(%+v) error(%v)", dynamics, err)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second * 5)
	}
}
