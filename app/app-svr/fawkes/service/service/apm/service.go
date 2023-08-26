package monitor

import (
	"context"
	"time"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/robfig/cron"
)

// Service service struct info.
type Service struct {
	c           *conf.Config
	fkDao       *fawkes.Dao
	loadSuccess bool
	cron        *cron.Cron
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		fkDao:       fawkes.New(c),
		cron:        cron.New(),
		loadSuccess: true,
	}
	s.loadApmParams()
	if err := s.cron.AddFunc(s.c.Cron.LoadApmParams, s.loadApmParams); err != nil {
		log.Error("AddFunc error(%v)", err)
	}
	s.cron.Start()
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

func (s *Service) loadApmParams() {
	log.Info("cronLog start loadApmParams")
	if time.Now().Hour() == 0 || !s.loadSuccess {
		s.loadSuccess = true
	}
}
