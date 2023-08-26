package fawkes

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	alarmdao "go-gateway/app/app-svr/app-resource/interface/dao/alarm"
	fkdao "go-gateway/app/app-svr/app-resource/interface/dao/fawkes"
	locdao "go-gateway/app/app-svr/app-resource/interface/dao/location"
	"go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model"
	fkappmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	fkcdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"

	"github.com/robfig/cron"
)

// Service module service.
type Service struct {
	c        *conf.Config
	fkDao    *fkdao.Dao
	locDao   *locdao.Dao
	alarmDao *alarmdao.Dao
	// cache
	versionCache       map[string]map[int64]*fkmdl.Version
	upgrdConfigCache   map[string]map[int64]*fkcdmdl.UpgradConfig
	packCache          map[string]map[int64][]*fkcdmdl.Pack
	patchCache         map[string]map[string]*fkcdmdl.Patch
	filterConfigCache  map[string]map[int64]*fkcdmdl.FilterConfig
	channelCache       map[string]map[int64]*fkappmdl.Channel
	flowCache          map[string]map[int64]*fkcdmdl.FlowConfig
	hfUpgradeCache     map[string]map[int64][]*fkappmdl.HfUpgrade
	apkCache           map[int64]map[string]map[string][]*bizapkmdl.Apk
	testFlightCache    map[string]map[string]*fawkes.TestFlight
	tribeCache         map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk
	tribeRelationCache map[int64]int64
	// cron
	cron *cron.Cron
}

// New new a module service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:        c,
		fkDao:    fkdao.New(c),
		locDao:   locdao.New(c),
		alarmDao: alarmdao.New(c),
		// cache
		versionCache:      make(map[string]map[int64]*fkmdl.Version),
		upgrdConfigCache:  make(map[string]map[int64]*fkcdmdl.UpgradConfig),
		packCache:         make(map[string]map[int64][]*fkcdmdl.Pack),
		patchCache:        make(map[string]map[string]*fkcdmdl.Patch),
		filterConfigCache: make(map[string]map[int64]*fkcdmdl.FilterConfig),
		channelCache:      make(map[string]map[int64]*fkappmdl.Channel),
		flowCache:         make(map[string]map[int64]*fkcdmdl.FlowConfig),
		hfUpgradeCache:    make(map[string]map[int64][]*fkappmdl.HfUpgrade),
		testFlightCache:   make(map[string]map[string]*fawkes.TestFlight),
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.loadFawkes()
	s.loadApk()
	s.loadTribe()
	if err := s.cron.AddFunc(s.c.Cron.LoadFawkes, s.loadFawkes); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc(s.c.Cron.LoadFawkes, s.loadApk); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc(s.c.Cron.LoadFawkes, s.loadTribe); err != nil {
		panic(err)
	}
	s.loadTestFlight()
	if err := s.cron.AddFunc(s.c.Cron.LoadFawkes, s.loadTestFlight); err != nil {
		panic(err)
	}

}

func (s *Service) loadFawkes() {
	log.Info("cronLog start loadFawkes")
	var (
		err             error
		tmpVersion      map[string]map[int64]*fkmdl.Version
		tmpUpgrdConfig  map[string]map[int64]*fkcdmdl.UpgradConfig
		tmpPack         map[string]map[int64][]*fkcdmdl.Pack
		tmpPatch        map[string]map[string]*fkcdmdl.Patch
		tmpFilterConfig map[string]map[int64]*fkcdmdl.FilterConfig
		tmpChannel      map[string]map[int64]*fkappmdl.Channel
		tmpFlow         map[string]map[int64]*fkcdmdl.FlowConfig
		tmpHfUpgrade    map[string]map[int64][]*fkappmdl.HfUpgrade
	)
	if tmpVersion, err = s.fkDao.Versions(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpVersion %+v", tmpVersion)
	if tmpUpgrdConfig, err = s.fkDao.UpgradConfig(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpUpgrdConfig %+v", tmpUpgrdConfig)
	if tmpPack, err = s.fkDao.Packs(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpPack %+v", tmpPack)
	if tmpPatch, err = s.fkDao.Patch(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpPatch %+v", tmpPatch)
	if tmpFilterConfig, err = s.fkDao.FilterConfig(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpFilterConfig %+v", tmpFilterConfig)
	if tmpChannel, err = s.fkDao.AppChannel(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpChannel %+v", tmpChannel)
	if tmpFlow, err = s.fkDao.FlowConfig(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpFlow %+v", tmpFlow)
	if tmpHfUpgrade, err = s.fkDao.HfUpgrade(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("tmpHfUpgrade %+v", tmpHfUpgrade)
	s.versionCache = tmpVersion
	s.upgrdConfigCache = tmpUpgrdConfig
	s.packCache = tmpPack
	s.patchCache = tmpPatch
	s.filterConfigCache = tmpFilterConfig
	s.channelCache = tmpChannel
	s.flowCache = tmpFlow
	s.hfUpgradeCache = tmpHfUpgrade
	//nolint:gosimple
	return
}

func (s *Service) loadApk() {
	log.Info("load apk start")
	res, err := s.fkDao.ApkList(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	s.apkCache = res
}

func (s *Service) loadTribe() {
	log.Info("load tribe start")
	res, err := s.fkDao.TribeList(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	tribeRelation, err := s.fkDao.TribeRelation(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	s.tribeCache = res
	s.tribeRelationCache = tribeRelation
}
