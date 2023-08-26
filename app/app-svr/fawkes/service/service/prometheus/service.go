package prometheus

import (
	"context"
	"os"

	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/robfig/cron"
)

// Service struct.
type Service struct {
	c          *conf.Config
	fkDao      *fkdao.Dao
	httpClient *bm.Client
	cron       *cron.Cron
	cronSwitch string
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		fkDao:      fkdao.New(c),
		httpClient: bm.NewClient(c.HTTPClient),
		cronSwitch: os.Getenv("CRON_SWITCH"),
	}
	if s.cronSwitch == "on" {
		s.cron = cron.New()
		_ = s.cron.AddFunc(s.c.Cron.LoadFawkesMoni, s.ciInWaiting)
		_ = s.cron.AddFunc(s.c.Cron.LoadFawkesMoniMergeNotice, s.longProcessMerge)
		s.cron.Start()
	}
	return
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
