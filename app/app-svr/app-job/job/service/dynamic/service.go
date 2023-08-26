package dynamic

import (
	"fmt"

	"go-gateway/app/app-svr/app-job/job/conf"
	dyndao "go-gateway/app/app-svr/app-job/job/dao/dynamic"
	rcmddao "go-gateway/app/app-svr/app-job/job/dao/recommend"

	"github.com/robfig/cron"
)

type Service struct {
	c       *conf.Config
	cron    *cron.Cron
	rcmdDao *rcmddao.Dao
	dynDao  *dyndao.Dao
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:       c,
		cron:    cron.New(),
		rcmdDao: rcmddao.New(c),
		dynDao:  dyndao.New(c),
	}
	s.load()
	checkErr(s.cron.AddFunc("@every 3m", s.load)) // 间隔3分钟
	return s
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("cron add func loadCache error(%+v)", err))
	}
}

func (s *Service) Close() {
	s.cron.Stop()
}
