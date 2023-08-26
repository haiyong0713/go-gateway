package service

import (
	"context"
	"go-common/library/log"
	"sort"

	"go-gateway/app/app-svr/feature/service/conf"
	featuredao "go-gateway/app/app-svr/feature/service/dao/feature"
	tvdao "go-gateway/app/app-svr/feature/service/dao/tv"
	abtestMdl "go-gateway/app/app-svr/feature/service/model/abtest"
	buildLimitmdl "go-gateway/app/app-svr/feature/service/model/buildLimit"
	businessConfigMdl "go-gateway/app/app-svr/feature/service/model/businessConfig"
	"go-gateway/app/app-svr/feature/service/model/degrade"

	"github.com/robfig/cron"
)

type Service struct {
	c                   *conf.Config
	featureDao          *featuredao.Dao
	tvDao               *tvdao.Dao
	cron                *cron.Cron
	coverLimit          []*degrade.DisplayLimitRes
	videoShotLimit      []*degrade.DisplayLimitRes
	displayLimit        map[string]map[string]map[string]*degrade.Range
	displayQnLimit      map[string]map[string]map[string]*degrade.Range
	chanFeature         map[string]degrade.ChannelFeatrue
	tvSwitchCache       map[string][]*degrade.TvSwitch
	buildLimitCache     map[int64]map[string]*buildLimitmdl.BuildLimit
	businessConfigCache map[int64]map[string]*businessConfigMdl.BusinessConfig
	abtestCache         map[int64]map[string]*abtestMdl.ABTest
	tvSwitchKeysCache   map[string]*degrade.TvSwitch
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                   c,
		featureDao:          featuredao.New(c),
		tvDao:               tvdao.New(c),
		cron:                cron.New(),
		displayLimit:        make(map[string]map[string]map[string]*degrade.Range),
		displayQnLimit:      make(map[string]map[string]map[string]*degrade.Range),
		chanFeature:         make(map[string]degrade.ChannelFeatrue),
		tvSwitchCache:       make(map[string][]*degrade.TvSwitch),
		buildLimitCache:     make(map[int64]map[string]*buildLimitmdl.BuildLimit),
		businessConfigCache: make(map[int64]map[string]*businessConfigMdl.BusinessConfig),
		abtestCache:         make(map[int64]map[string]*abtestMdl.ABTest),
		tvSwitchKeysCache:   make(map[string]*degrade.TvSwitch),
	}
	// 定时任务
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.loadDisplayLimit()
	if err := s.cron.AddFunc(s.c.Degrade.Cfg.Cron, s.loadDisplayLimit); err != nil {
		panic(err)
	}
	s.loadChannelFeature()
	if err := s.cron.AddFunc(s.c.Degrade.Cfg.Cron, s.loadChannelFeature); err != nil {
		panic(err)
	}
	s.loadTvSwitch()
	if err := s.cron.AddFunc(s.c.Degrade.Cfg.Cron, s.loadTvSwitch); err != nil {
		panic(err)
	}
	s.loadBuildLimit()
	if err := s.cron.AddFunc("*/10 * * * * *", s.loadBuildLimit); err != nil {
		panic(err)
	}
	s.loadBusinessConfig()
	if err := s.cron.AddFunc("*/10 * * * * *", s.loadBusinessConfig); err != nil {
		panic(err)
	}
	s.loadAbtest()
	if err := s.cron.AddFunc("*/10 * * * * *", s.loadAbtest); err != nil {
		panic(err)
	}
}

func (s *Service) loadDisplayLimit() {
	res, err := s.tvDao.DisplayLimit(context.Background())
	if err != nil {
		log.Error("degradeError loadDisplayLimit err(%+v)", err)
		return
	}

	s.feature3QnLimits(res)

	// map[限制类型] map[渠道/品牌/机型] map[具体的渠道/品牌/机型名] *model.Range
	displayLimit := make(map[string]map[string]map[string]*degrade.Range)
	for k, limit := range res {
		if k == _featureCover { // 海报折损根据rank从低到高匹配
			sort.Slice(limit, func(i, j int) bool {
				return limit[i].Rank < limit[j].Rank
			})
			s.coverLimit = limit
			continue
		}
		if k == _videoShot { // 进度缩略图
			s.videoShotLimit = limit
			continue
		}
		displayLimit[k] = degrade.ToMap(limit)
	}
	s.displayLimit = displayLimit
}

// 加载渠道相关特性，如编码方式，开机自启动等
func (s *Service) loadChannelFeature() {
	chs, err := s.tvDao.ChannelFeature(context.Background())
	if err != nil {
		log.Error("degradeError loadDecodeType err(%+v)", err)
		return
	}
	s.chanFeature = chs
}

func (s *Service) loadTvSwitch() {
	tmpRes, tmpKeyRes, err := s.featureDao.TvSwitch(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	s.tvSwitchCache = tmpRes
	s.tvSwitchKeysCache = tmpKeyRes
}

func (s *Service) loadBuildLimit() {
	// get buildlimit trees
	trees, err := s.featureDao.BuildLimitTrees(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	var cacheTmp = make(map[int64]map[string]*buildLimitmdl.BuildLimit)
	for _, tree := range trees {
		tmp, errTmp := s.featureDao.BuildLimit(context.Background(), tree)
		if errTmp != nil {
			log.Error("%v", errTmp)
			return
		}
		cacheTmp[tree] = tmp
	}
	s.buildLimitCache = cacheTmp
}

func (s *Service) loadBusinessConfig() {
	// get business config trees
	trees, err := s.featureDao.BusinessConfigTrees(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	var cacheTmp = make(map[int64]map[string]*businessConfigMdl.BusinessConfig)
	for _, tree := range trees {
		tmp, errTmp := s.featureDao.BusinessConfig(context.Background(), tree)
		if errTmp != nil {
			log.Error("%v", errTmp)
			return
		}
		cacheTmp[tree] = tmp
	}
	s.businessConfigCache = cacheTmp
}

func (s *Service) loadAbtest() {
	// get abtest trees
	trees, err := s.featureDao.ABTestTrees(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	var cacheTmp = make(map[int64]map[string]*abtestMdl.ABTest)
	for _, tree := range trees {
		tmp, errTmp := s.featureDao.ABTest(context.Background(), tree)
		if errTmp != nil {
			log.Error("%v", errTmp)
			return
		}
		cacheTmp[tree] = tmp
	}
	s.abtestCache = cacheTmp
}
