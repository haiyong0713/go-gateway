package service

import (
	"go-gateway/app/app-svr/resource/job/conf"
	"go-gateway/app/app-svr/resource/job/dao"

	"github.com/robfig/cron"
)

type Service struct {
	c    *conf.Config
	dao  *dao.Dao
	cron *cron.Cron
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:    c,
		dao:  dao.New(c),
		cron: cron.New(),
	}
	// 本地缓存统一初始化(建议少用本地缓存)
	s.initCache()
	// 定时任务
	s.initCron()
	s.cron.Start()
	return
}

// 初始化本地缓存(建议少用)
func (s *Service) initCache() {}

func (s *Service) initCron() {
	if err := s.cron.AddFunc(s.c.Cron.LoadTabExt, s.loadTabExt); err != nil {
		panic(err)
	}
	// custom config
	if err := s.cron.AddFunc(s.c.Cron.LoadCustomConfig, s.loadCustomConfig); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc(s.c.Cron.LoadSkinExt, s.loadSkinExt); err != nil {
		panic(err)
	}
	// 通用黑白名单->数据源来自manager-gw
	if err := s.cron.AddFunc(s.c.Cron.LoadBWList, s.loadBWList); err != nil {
		panic(err)
	}

}

// Close Databus consumer close.
func (s *Service) Close() {
	s.cron.Stop()
}
