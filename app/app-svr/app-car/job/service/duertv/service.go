package duertv

import (
	"sync"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/conf"
	arcdao "go-gateway/app/app-svr/app-car/job/dao/archive"
	bgmdao "go-gateway/app/app-svr/app-car/job/dao/bangumi"
	cldao "go-gateway/app/app-svr/app-car/job/dao/channel"
	dtdao "go-gateway/app/app-svr/app-car/job/dao/duertv"
	rcmddao "go-gateway/app/app-svr/app-car/job/dao/recommend"
	rgdao "go-gateway/app/app-svr/app-car/job/dao/region"
	"go-gateway/app/app-svr/app-car/job/model/bangumi"
	"go-gateway/app/app-svr/app-car/job/model/region"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

type Service struct {
	c    *conf.Config
	bgm  *bgmdao.Dao
	arc  *arcdao.Dao
	dt   *dtdao.Dao
	rg   *rgdao.Dao
	cl   *cldao.Dao
	rcmd *rcmddao.Dao

	// push
	duertvBgmChan          chan *bangumi.Content
	duertvBgmOffshelveChan chan *bangumi.Offshelve
	// waiter
	waiter         sync.WaitGroup
	archiveRailGun *railgun.Railgun
	// cache
	oneRegions map[int32]*region.Region
	// duertvBangumiAllRailGun
	duertvBangumiAllRailGun *railgun.Railgun
	regionRailGun           *railgun.Railgun
	duertvRankAllRailGun    *railgun.Railgun
	duertvHostRailGun       *railgun.Railgun
	creativeClient          creativeAPI.VideoUpOpenClient
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:    c,
		bgm:  bgmdao.New(c),
		arc:  arcdao.New(c),
		dt:   dtdao.New(c),
		rg:   rgdao.New(c),
		cl:   cldao.New(c),
		rcmd: rcmddao.New(c),
		// push
		duertvBgmChan:          make(chan *bangumi.Content, 1024000),
		duertvBgmOffshelveChan: make(chan *bangumi.Offshelve, 1024000),
		// cache
		oneRegions: map[int32]*region.Region{},
	}
	var err error
	if s.creativeClient, err = creativeAPI.NewClient(c.CreativeClient); err != nil {
		panic("creativeGRPC not found!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.duertvPushproc()
	if s.c.Custom.PushBagumiAll {
		// 全量增量数据
		s.duertvBangumiAll()
	}
	if s.c.Custom.PushBagumiOffshelve {
		// nolint:biligowordcheck
		go s.duertvPushOffshelveproc()
		// 全量下架数据
		s.duertvBangumiOffshelve()
	}
	// 稿件媒资推送
	s.initArchiveRailGun(c.ArchiveRailGun.Databus, c.ArchiveRailGun.SingleConfig, c.ArchiveRailGun.Cfg)
	// PGC下架媒资推送
	s.initPGCRailGun(c.PGCRailGun.Databus, c.PGCRailGun.SingleConfig, c.PGCRailGun.Cfg)
	// 每小时定时跑一次
	s.initduertvBangumiAllRailGun(c.DuertvBangumiGun.CronInputer, c.DuertvBangumiGun.CronProcessor, c.DuertvBangumiGun.Cfg)
	// 每5分钟跑一次
	s.initRegionRailGun(c.RegionGun.CronInputer, c.RegionGun.CronProcessor, c.RegionGun.Cfg)
	// 每小时定时跑一次
	s.initduertvRankAllRailGun(c.DuertvUGCGun.CronInputer, c.DuertvUGCGun.CronProcessor, c.DuertvUGCGun.Cfg)
	// 每小时定时跑一次
	s.initduertvHostAllRailGun(c.DuertvUGCGun.CronInputer, c.DuertvUGCGun.CronProcessor, c.DuertvUGCGun.Cfg)
	s.duertvBangumiAllRailGun.Start()
	s.regionRailGun.Start()
	s.duertvRankAllRailGun.Start()
	s.duertvHostRailGun.Start()
	return s
}

func (s *Service) Close() {
	s.waiter.Wait()
	s.duertvBangumiAllRailGun.Close()
	s.regionRailGun.Close()
	s.duertvRankAllRailGun.Close()
	s.duertvHostRailGun.Close()
	log.Info("app-car-job closed.")
}
