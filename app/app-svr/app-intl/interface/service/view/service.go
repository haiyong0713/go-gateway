package view

import (
	"context"
	"time"

	"go-gateway/app/app-svr/app-intl/interface/conf"
	accdao "go-gateway/app/app-svr/app-intl/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-intl/interface/dao/act"
	arcdao "go-gateway/app/app-svr/app-intl/interface/dao/archive"
	assdao "go-gateway/app/app-svr/app-intl/interface/dao/assist"
	audiodao "go-gateway/app/app-svr/app-intl/interface/dao/audio"
	bandao "go-gateway/app/app-svr/app-intl/interface/dao/bangumi"
	channeldao "go-gateway/app/app-svr/app-intl/interface/dao/channel"
	coindao "go-gateway/app/app-svr/app-intl/interface/dao/coin"
	cfcdao "go-gateway/app/app-svr/app-intl/interface/dao/content"
	dmdao "go-gateway/app/app-svr/app-intl/interface/dao/dm"
	favdao "go-gateway/app/app-svr/app-intl/interface/dao/favorite"
	fkdao "go-gateway/app/app-svr/app-intl/interface/dao/fawkes"
	locdao "go-gateway/app/app-svr/app-intl/interface/dao/location"
	rcmddao "go-gateway/app/app-svr/app-intl/interface/dao/recommend"
	rgndao "go-gateway/app/app-svr/app-intl/interface/dao/region"
	reldao "go-gateway/app/app-svr/app-intl/interface/dao/relation"
	rscdao "go-gateway/app/app-svr/app-intl/interface/dao/resource"
	seasondao "go-gateway/app/app-svr/app-intl/interface/dao/season"
	steindao "go-gateway/app/app-svr/app-intl/interface/dao/stein"
	tagdao "go-gateway/app/app-svr/app-intl/interface/dao/tag"
	thumbupdao "go-gateway/app/app-svr/app-intl/interface/dao/thumbup"
	vudao "go-gateway/app/app-svr/app-intl/interface/dao/videoup"
	"go-gateway/app/app-svr/app-intl/interface/model/region"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"

	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	"go-common/library/stat/prom"
)

//var _groupSeasonAb = int64(41)

// Service is view service
type Service struct {
	c     *conf.Config
	pHit  *prom.Prom
	pMiss *prom.Prom
	prom  *prom.Prom
	// dao
	accDao     *accdao.Dao
	arcDao     *arcdao.Dao
	tagDao     *tagdao.Dao
	favDao     *favdao.Dao
	banDao     *bandao.Dao
	rgnDao     *rgndao.Dao
	assDao     *assdao.Dao
	audioDao   *audiodao.Dao
	thumbupDao *thumbupdao.Dao
	rscDao     *rscdao.Dao
	relDao     *reldao.Dao
	coinDao    *coindao.Dao
	dmDao      *dmdao.Dao
	locDao     *locdao.Dao
	steinDao   *steindao.Dao
	rcmdDao    *rcmddao.Dao
	actDao     *actdao.Dao
	seasonDao  *seasondao.Dao
	vuDao      *vudao.Dao
	fawkes     *fkdao.Dao
	channelDao *channeldao.Dao
	cfcDao     *cfcdao.Dao
	// tick
	tick time.Duration
	// region
	region map[int8]map[int16]*region.Region
	// chan
	inCh chan interface{}
	// season white list
	seasonMids map[int64]struct{}
	// hot aids
	hotAids map[int64]struct{}
	// fawkes version
	FawkesVersionCache map[string]map[string]*fkmdl.Version
	// infoc
	relateInfocv2 infocv2.Infoc
	viewInfocv2   infocv2.Infoc
}

// New new archive
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		pHit:  prom.CacheHit,
		pMiss: prom.CacheMiss,
		prom:  prom.BusinessInfoCount,
		// dao
		accDao:     accdao.New(c),
		arcDao:     arcdao.New(c),
		tagDao:     tagdao.New(c),
		favDao:     favdao.New(c),
		banDao:     bandao.New(c),
		rgnDao:     rgndao.New(c),
		assDao:     assdao.New(c),
		relDao:     reldao.New(c),
		coinDao:    coindao.New(c),
		audioDao:   audiodao.New(c),
		thumbupDao: thumbupdao.New(c),
		rscDao:     rscdao.New(c),
		dmDao:      dmdao.New(c),
		locDao:     locdao.New(c),
		steinDao:   steindao.New(c),
		rcmdDao:    rcmddao.New(c),
		actDao:     actdao.New(c),
		seasonDao:  seasondao.New(c),
		vuDao:      vudao.New(c),
		fawkes:     fkdao.New(c),
		channelDao: channeldao.New(c),
		cfcDao:     cfcdao.New(c),
		// tick
		tick: time.Duration(c.Tick),
		// region
		region: map[int8]map[int16]*region.Region{},
		// chan
		inCh: make(chan interface{}, 1024),
		// season white mids
		seasonMids: map[int64]struct{}{},
		// hot aids
		hotAids: map[int64]struct{}{},
		// fawkes
		FawkesVersionCache: make(map[string]map[string]*fkmdl.Version),
	}
	var err error
	if s.relateInfocv2, err = infocv2.New(s.c.RelateInfocv2.Conf); err != nil {
		panic(err)
	}
	if s.viewInfocv2, err = infocv2.New(s.c.ViewInfocv2.Conf); err != nil {
		panic(err)
	}
	// load data
	s.loadRegion()
	s.loadHotCache()
	s.loadFawkes()
	go s.infocproc()
	go s.hotCacheproc()
	go s.loadFawkesProc()
	return
}

// Ping is dao ping.
func (s *Service) Ping(c context.Context) (err error) {
	return s.arcDao.Ping(c)
}

// loadRegion is.
func (s *Service) loadRegion() {
	res, err := s.rgnDao.Seconds(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.region = res
}

func (s *Service) loadHotCache() {
	tmp, err := s.rcmdDao.RecommendHot(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.hotAids = tmp
}

func (s *Service) hotCacheproc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.HotAidsTick))
		s.loadHotCache()
	}
}

func (s *Service) loadFawkes() {
	fv, err := s.fawkes.FawkesVersion(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	if len(fv) > 0 {
		s.FawkesVersionCache = fv
	}
}

func (s *Service) loadFawkesProc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.FawkesTick))
		s.loadFawkes()
	}
}
