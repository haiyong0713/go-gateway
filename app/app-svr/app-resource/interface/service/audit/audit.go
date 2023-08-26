package audit

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	auditdao "go-gateway/app/app-svr/app-resource/interface/dao/audit"

	"github.com/robfig/cron"
)

// Service audit service.
type Service struct {
	c   *conf.Config
	dao *auditdao.Dao
	// tick
	tick time.Duration
	// cache
	auditCache map[string]map[int]struct{}
	// cron
	cron *cron.Cron
}

// New new a audit service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: auditdao.New(c),
		// tick
		tick: time.Duration(c.Tick),
		// cache
		auditCache: map[string]map[int]struct{}{},
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.loadAuditCache()
	if err := s.cron.AddFunc(s.c.Cron.LoadAuditCache, s.loadAuditCache); err != nil {
		panic(err)
	}
}

// Audit
func (s *Service) Audit(c context.Context, mobiApp string, build int) (err error) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return ecode.OK
		}
	}
	return ecode.NotModified
}

func (s *Service) loadAuditCache() {
	log.Info("cronLog start loadAuditCache")
	as, err := s.dao.Audits(context.TODO())
	if err != nil {
		log.Error("s.dao.Audits error(%v)", err)
		return
	}
	s.auditCache = as
}
