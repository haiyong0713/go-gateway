package ci

import (
	"context"
	"os"
	"sync"

	bm "go-common/library/net/http/blademaster"

	"go-common/library/log/infoc.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	cdSvr "go-gateway/app/app-svr/fawkes/service/service/cd"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	"go-gateway/app/app-svr/fawkes/service/service/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/robfig/cron"
)

// Service struct.
type Service struct {
	c             *conf.Config
	fkDao         *fkdao.Dao
	httpClient    *bm.Client
	ciChan        chan func()
	crontabCIProc sync.Map
	gitSvr        *gitSvr.Service
	cdSvr         *cdSvr.Service
	tribeSvr      *tribe.Service
	cron          *cron.Cron
	cronSwitch    string
	infoc         infoc.Infoc
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		fkDao:      fkdao.New(c),
		httpClient: bm.NewClient(c.HTTPClient),
		ciChan:     make(chan func(), 512),
		gitSvr:     gitSvr.New(c),
		tribeSvr:   tribe.New(c),
		cronSwitch: os.Getenv("CRON_SWITCH"),
		infoc:      c.Infoc,
	}
	log.Info("s.cronSwitch %v", s.cronSwitch)
	if s.cronSwitch == "on" {
		s.cron = cron.New()
		s.crontab()
		_ = s.cron.AddFunc("@every 30s", s.crontab)
		s.cron.Start()
	}
	err := s.gitSvr.SubscribeAsync(gitSvr.GitJobStatusChangeEvent, s.jobStatusChangeAction)
	if err != nil {
		panic(err)
	}
	// nolint:biligowordcheck
	go s.ciproc()
	return
}

// AddCiProc add ci proc
func (s *Service) AddCiProc(f func()) {
	select {
	case s.ciChan <- f:
	default:
		log.Warn("addCi chan full")
	}
}

func (s *Service) ciproc() {
	for {
		f, ok := <-s.ciChan
		if !ok {
			log.Warn("ci proc exit")
			return
		}
		f()
	}
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.fkDao.Close()
}
