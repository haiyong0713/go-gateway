package gitlab

import (
	"context"
	"os"

	bm "go-common/library/net/http/blademaster"

	"github.com/asaskevich/EventBus"

	"github.com/robfig/cron"
	gitlab "github.com/xanzy/go-gitlab"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service struct.
type Service struct {
	c            *conf.Config
	fkDao        *fkdao.Dao
	httpClient   *bm.Client
	gitlabClient *gitlab.Client
	cron         *cron.Cron
	cronSwitchOn bool
	event        EventBus.Bus
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		fkDao:        fkdao.New(c),
		httpClient:   bm.NewClient(c.HTTPClient),
		gitlabClient: gitlab.NewClient(nil, c.Gitlab.Token),
		cron:         cron.New(),
		cronSwitchOn: os.Getenv("CRON_SWITCH") == "on",
		event:        EventBus.New(),
	}
	_ = s.gitlabClient.SetBaseURL(c.Gitlab.Host)
	if s.cronSwitchOn {
		if err := s.cron.AddFunc(conf.Conf.Gitlab.CronExpression, func() { _ = s.RefreshStatusProc() }); err != nil {
			panic(err)
		}
		if err := s.cron.AddFunc(conf.Conf.Gitlab.CronExpression, func() { _ = s.RefreshHotfixStatusProc() }); err != nil {
			panic(err)
		}
		if err := s.cron.AddFunc(conf.Conf.Gitlab.CronExpression, func() { _ = s.RefreshBizApkStatusProc() }); err != nil {
			panic(err)
		}
		s.cron.Start()
	}

	if err := s.event.Subscribe(GitMergeEvent, s.MergeAction); err != nil {
		panic(err)
	}
	if err := s.event.Subscribe(GitCommentEvent, s.CommentAction); err != nil {
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
