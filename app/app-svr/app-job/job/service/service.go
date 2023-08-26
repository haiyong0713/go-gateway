package service

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/app-job/job/conf"
	monitordao "go-gateway/app/app-svr/app-job/job/dao/monitor"
	pushdao "go-gateway/app/app-svr/app-job/job/dao/push"
	spacedao "go-gateway/app/app-svr/app-job/job/dao/space"
	viewdao "go-gateway/app/app-svr/app-job/job/dao/view"
	"go-gateway/app/app-svr/app-job/job/model"
	"go-gateway/app/app-svr/app-job/job/model/resource"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	"github.com/robfig/cron"
)

// Service is service.
type Service struct {
	c *conf.Config
	// vdao
	vdao          *viewdao.Dao
	spdao         *spacedao.Dao
	monitorDao    *monitordao.Dao
	pushDao       *pushdao.Dao
	contributeSub *databus.Databus
	waiter        sync.WaitGroup
	// space
	contributeChan chan *model.ContributeMsg
	closed         bool
	resourceRPC    *resrpc.Service
	sideBars       map[int8]map[int][]*resmdl.SideBar
	cron           *cron.Cron
	ResourceMngSub *databus.Databus
}

// New new a service.
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		vdao:        viewdao.New(c),
		spdao:       spacedao.New(c),
		monitorDao:  monitordao.New(c),
		pushDao:     pushdao.New(c),
		resourceRPC: resrpc.New(nil),
		closed:      false,
		cron:        cron.New(),
	}
	// contribute consumer
	if model.EnvRun() {
		s.contributeChan = make(chan *model.ContributeMsg, 10240)
		s.contributeSub = databus.New(c.ContributeSub)
		s.waiter.Add(1)
		go s.contributeConsumeproc()
		for i := 0; i < runtime.NumCPU(); i++ {
			s.waiter.Add(1)
			go s.contributeproc()
		}
		s.ResourceMngSub = databus.New(c.ResourceMngSub)
		s.waiter.Add(1)
		go s.resourceProc()
	}
	// retry consumer
	for i := 0; i < 4; i++ {
		s.waiter.Add(1)
		go s.retryproc()
	}
	if err := s.cron.AddFunc("@every 1m", s.watchSideBar); err != nil {
		panic(err)
	}
	s.cron.Start()
	return
}

// Close Databus consumer close.
func (s *Service) Close() {
	s.closed = true
	s.cron.Stop()
	// 原contributeSub的关闭不合理，现更正
	if model.EnvRun() {
		s.contributeSub.Close()
		s.ResourceMngSub.Close()
	}
	s.waiter.Wait()
}

func (s *Service) Ping(c context.Context) (err error) {
	return
}

// resourceProc resource mng msg process
func (s *Service) resourceProc() {
	defer s.waiter.Done()
	for {
		var (
			msg *databus.Message
			ok  bool
			err error
		)
		if s.closed {
			log.Error("s.resourceMngSub.messages closed")
			return
		}
		if msg, ok = <-s.ResourceMngSub.Messages(); !ok {
			log.Error("s.ResourceMngSub.messages closed")
			return
		}
		_ = msg.Commit()
		em := &resource.EntryMsg{}
		if err = json.Unmarshal(msg.Value, em); err != nil {
			log.Error("broadcast json.Unmarshal(%v) error(%v)", msg.Value, err)
			continue
		}
		log.Info("broadcast got resource mng message key(%s) value(%s) ", msg.Key, msg.Value)
		s.pushDao.PushEntry(context.Background(), em)
	}
}
