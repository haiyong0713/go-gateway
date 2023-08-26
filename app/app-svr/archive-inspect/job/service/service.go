package service

import (
	"fmt"

	"go-gateway/app/app-svr/archive-inspect/job/conf"
	"go-gateway/app/app-svr/archive-inspect/job/dao/archive"
	locdao "go-gateway/app/app-svr/archive-inspect/job/dao/location"
	"go-gateway/app/app-svr/archive-inspect/job/dao/result"

	"github.com/robfig/cron"
	"go-common/library/cache/redis"
	"go-common/library/database/taishan"
)

// Service service
type Service struct {
	c          *conf.Config
	archiveDao *archive.Dao
	resultDao  *result.Dao
	locDao     *locdao.Dao
	arcRedises []*redis.Pool
	Taishan    *Taishan
	cron       *cron.Cron
}

type Taishan struct {
	client   taishan.TaishanProxyClient
	tableCfg tableConfig
}

type tableConfig struct {
	Table string
	Token string
}

// New is archive service implementation.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		archiveDao: archive.New(c),
		resultDao:  result.New(c),
		locDao:     locdao.New(c),
		cron:       cron.New(),
	}
	for _, re := range s.c.ArcRedises {
		s.arcRedises = append(s.arcRedises, redis.NewPool(re))
	}
	t, err := taishan.NewClient(c.TaiShanClient)
	if err != nil {
		panic(fmt.Sprintf("taishan.NewClient error(%+v)", err))
	}
	s.Taishan = &Taishan{
		client: t,
		tableCfg: tableConfig{
			Table: c.Taishan.Table,
			Token: c.Taishan.Token,
		},
	}
	s.initCron()
	s.cron.Start()
	return s
}

func (s *Service) initCron() {
	var err error
	if err = s.cron.AddFunc(s.c.Cron.CheckModifyAids, s.checkModifyAids); err != nil {
		panic(err)
	}
}

// Close kafaka consumer close.
func (s *Service) Close() (err error) {
	s.cron.Stop()
	s.resultDao.Close()
	s.archiveDao.Close()
	return
}
