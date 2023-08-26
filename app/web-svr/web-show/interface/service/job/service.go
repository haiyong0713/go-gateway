package job

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/web-show/interface/conf"
	"go-gateway/app/web-svr/web-show/interface/dao/job"
	jobmdl "go-gateway/app/web-svr/web-show/interface/model/job"

	"github.com/robfig/cron"
)

var (
	_emptyJobs = make([]*jobmdl.Job, 0)
)

// Service struct
type Service struct {
	dao        *job.Dao
	cache      []*jobmdl.Job
	cron       *cron.Cron
	jobRunning bool
}

// New init
func New(c *conf.Config) (s *Service) {
	s = &Service{}
	s.dao = job.New(c)
	s.cache = _emptyJobs
	s.cron = cron.New()
	if err := s.loadCron(c); err != nil {
		panic(err)
	}
	return
}

func (s *Service) loadCron(c *conf.Config) error {
	s.loadproc()
	err := s.cron.AddFunc(c.Cron.ReloadJob, s.loadproc)
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

// jobproc load job infos to cache
func (s *Service) loadproc() {
	s.reload()
	log.Info("job loadproc")
}

// reload
func (s *Service) reload() {
	if s.jobRunning {
		return
	}
	s.jobRunning = true
	defer func() {
		s.jobRunning = false
	}()
	js, err := s.dao.Jobs(context.Background())
	if err != nil {
		log.Error("s.job.Jobs error(%v)", err)
		return
	} else if len(js) == 0 {
		s.cache = _emptyJobs
	}
	cates, err := s.dao.Categories(context.Background())
	if err != nil {
		log.Error("job.Categories error(%v)", err)
		return
	}
	cs := make(map[int]string, len(cates))
	for _, cate := range cates {
		cs[cate.ID] = cate.Name
	}
	for _, j := range js {
		j.JobsCla = cs[j.CateID]
		j.Location = cs[j.AddrID]
	}
	s.cache = js
}

// Ping Service
func (s *Service) Ping(c context.Context) (err error) {
	err = s.dao.Ping(c)
	return
}

// Close Service
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}
