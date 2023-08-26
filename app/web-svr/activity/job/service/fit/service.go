package fit

import (
	"fmt"
	actPlatform "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/conf/env"
	"go-common/library/queue/databus"
	"go-common/library/rate/limit/quota"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/fit"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"sync"
)

const (
	followRWWaiterNameFmt = "%s.%s.main.web-svr.activity-job|Fit|FollowRW|total"
	databusWaiterNameFmt  = "%s.%s.main.web-svr.activity-job|Fit|DataBusTM|total"
)

// Service fit service
type Service struct {
	c                     *conf.Config
	dao                   fit.Dao
	arcClient             arcapi.ArchiveClient
	actPlatClient         actPlatform.ActPlatClient
	fitActivityHistorySub *databus.Databus
	fitTunnelPub          *databus.Databus
	// waiter
	waiter              sync.WaitGroup
	waiterFollowRWLimit quota.Waiter
	waiterDatabusLimit  quota.Waiter
	closed              bool
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                     c,
		dao:                   fit.New(c),
		fitActivityHistorySub: initialize.NewDatabusV1(c.FitDatabusCfg.FitActivityHistorySub),
		fitTunnelPub:          initialize.NewDatabusV1(c.FitTunnelPub),
		waiterFollowRWLimit:   quota.NewWaiter(&quota.WaiterConfig{ID: fmt.Sprintf(followRWWaiterNameFmt, env.DeployEnv, env.Zone)}),
		waiterDatabusLimit:    quota.NewWaiter(&quota.WaiterConfig{ID: fmt.Sprintf(databusWaiterNameFmt, env.DeployEnv, env.Zone)}),
		actPlatClient:         client.ActplatClient,
		arcClient:             client.ArcClient,
	}
	// fit
	s.waiter.Add(1)
	go s.SendGiftToUser()
	return s
}

// Close ...
func (s *Service) Close() (err error) {
	defer s.waiter.Wait()
	s.closed = true
	s.fitActivityHistorySub.Close()
	s.fitTunnelPub.Close()
	return
}
