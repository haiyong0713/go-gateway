package feedback

import (
	"context"
	"os"

	"github.com/asaskevich/EventBus"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/robfig/cron"
)

// Service struct.
type Service struct {
	c          *conf.Config
	fkDao      *fkdao.Dao
	cron       *cron.Cron
	cronSwitch string
	event      EventBus.Bus
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		fkDao:      fkdao.New(c),
		cronSwitch: os.Getenv("CRON_SWITCH"),
		event:      EventBus.New(),
	}
	log.Info("s.cronSwitch %v", s.cronSwitch)
	if s.cronSwitch == "on" {
		s.cron = cron.New()
		_ = s.cron.AddFunc("@every 60m", s.AlertFeedback)
		s.cron.Start()
	}
	if err := s.event.SubscribeAsync(UpdateEvent, s.feedbackUpdateAction, false); err != nil {
		panic(err)
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
