package space

import (
	"context"
	"runtime"
	"time"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	actdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/activity"
	addao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/ad"
	aidao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/ai"
	asdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/answer"
	arcdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/archive"
	artdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/article"
	audiodao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/audio"
	bgmdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bangumi"
	bplusdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bplus"
	channeldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/channel"
	cheesedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/cheese"
	coindao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/coin"
	comicdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/comic"
	commdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/community"
	digitaldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/digital"
	dyndao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/dynamic"
	elecdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/elec"
	favdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/favorite"
	gallerydao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/gallery"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/game"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/garb"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/guard"
	hisdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/history"
	livedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/live"
	malldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/mall"
	memberdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/member"
	paydao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/pay"
	reldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/relation"
	resdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/resource"
	schooldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/school"
	searchdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/search"
	srchdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/search"
	seriesdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/series"
	shopdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/shop"
	spcdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/space"
	tagdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/tag"
	teendao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/teenagers"
	thumbupdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/thumbup"
	ugcSeasonbdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/ugc_season"
	uparcdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/up-archive"
	vipdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/vip"

	"github.com/robfig/cron"
)

// Service is space service
type Service struct {
	c            *conf.Config
	arcDao       *arcdao.Dao
	spcDao       *spcdao.Dao
	accDao       *accdao.Dao
	coinDao      *coindao.Dao
	commDao      *commdao.Dao
	srchDao      *srchdao.Dao
	favDao       *favdao.Dao
	bgmDao       *bgmdao.Dao
	tagDao       *tagdao.Dao
	liveDao      *livedao.Dao
	elecDao      *elecdao.Dao
	artDao       *artdao.Dao
	audioDao     *audiodao.Dao
	relDao       *reldao.Dao
	bplusDao     *bplusdao.Dao
	shopDao      *shopdao.Dao
	thumbupDao   *thumbupdao.Dao
	gameDao      *game.Dao
	payDao       *paydao.Dao
	memberDao    *memberdao.Dao
	mallDao      *malldao.Dao
	comicDao     *comicdao.Dao
	ugcSeasonDao *ugcSeasonbdao.Dao
	cheeseDao    *cheesedao.Dao
	adDao        *addao.Dao
	garbDao      *garb.Dao
	aiDao        *aidao.Dao
	guardDao     *guard.Dao
	teenDao      *teendao.Dao
	hisDao       *hisdao.Dao
	searchDao    *searchdao.Dao
	seriesDao    *seriesdao.Dao
	actDao       *actdao.Dao
	channelDao   *channeldao.Dao
	upArcDao     *uparcdao.Dao
	resDao       *resdao.Dao
	dynDao       *dyndao.Dao
	schoolDao    *schooldao.Dao
	galleryDao   *gallerydao.Dao
	vipDao       *vipdao.Dao
	answerDao    *asdao.Dao
	digitalDao   *digitaldao.Dao
	// chan
	mCh            chan func()
	tick           time.Duration
	BlackList      map[int64]struct{}
	hotAids        map[int64]struct{}
	BvTick         time.Duration
	UpRcmdBlockMap map[int64]struct{}
	// cron
	cron *cron.Cron
}

// New new space
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		arcDao:       arcdao.New(c),
		spcDao:       spcdao.New(c),
		accDao:       accdao.New(c),
		coinDao:      coindao.New(c),
		commDao:      commdao.New(c),
		srchDao:      srchdao.New(c),
		favDao:       favdao.New(c),
		bgmDao:       bgmdao.New(c),
		tagDao:       tagdao.New(c),
		liveDao:      livedao.New(c),
		elecDao:      elecdao.New(c),
		artDao:       artdao.New(c),
		audioDao:     audiodao.New(c),
		relDao:       reldao.New(c),
		bplusDao:     bplusdao.New(c),
		shopDao:      shopdao.New(c),
		thumbupDao:   thumbupdao.New(c),
		gameDao:      game.New(c),
		payDao:       paydao.New(c),
		memberDao:    memberdao.New(c),
		mallDao:      malldao.New(c),
		comicDao:     comicdao.New(c),
		ugcSeasonDao: ugcSeasonbdao.New(c),
		cheeseDao:    cheesedao.New(c),
		adDao:        addao.New(c),
		garbDao:      garb.New(c),
		aiDao:        aidao.New(c),
		guardDao:     guard.New(c),
		teenDao:      teendao.New(c),
		hisDao:       hisdao.New(c),
		searchDao:    searchdao.New(c),
		seriesDao:    seriesdao.New(c),
		actDao:       actdao.New(c),
		channelDao:   channeldao.New(c),
		upArcDao:     uparcdao.New(c),
		resDao:       resdao.New(c),
		dynDao:       dyndao.New(c),
		schoolDao:    schooldao.New(c),
		galleryDao:   gallerydao.New(c),
		vipDao:       vipdao.New(c),
		answerDao:    asdao.New(c),
		digitalDao:   digitaldao.New(c),
		// mc proc
		mCh: make(chan func(), 1024),
		//nolint:staticcheck
		tick:           time.Duration(c.Tick),
		BlackList:      make(map[int64]struct{}),
		hotAids:        make(map[int64]struct{}),
		UpRcmdBlockMap: make(map[int64]struct{}),
		// cron
		cron: cron.New(),
	}
	// video db
	for i := 0; i < runtime.NumCPU(); i++ {
		//nolint:biligowordcheck
		go s.cacheproc()
	}
	//nolint:staticcheck
	if c != nil && c.Space != nil {
		for _, mid := range c.Space.ForbidMid {
			s.BlackList[mid] = struct{}{}
		}
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.loadBlacklist()
	s.loadHotCache()
	s.loadUpRcmdBlockList()
	var err error
	if err = s.cron.AddFunc(s.c.Cron.LoadBlacklist, s.loadBlacklist); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadHotCache, s.loadHotCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadUpRcmdBlockList, s.loadUpRcmdBlockList); err != nil {
		panic(err)
	}
}

// addCache add archive to mc or redis
func (s *Service) addCache(f func()) {
	select {
	case s.mCh <- f:
	default:
		log.Warn("cacheproc chan full")
	}
}

// cacheproc write memcache and stat redis use goroutine
func (s *Service) cacheproc() {
	for {
		f := <-s.mCh
		f()
	}
}

// Ping check server ok
func (s *Service) Ping(c context.Context) (err error) {
	return
}

// loadBlacklist
func (s *Service) loadBlacklist() {
	log.Info("cronLog start loadBlacklist")
	list, err := s.spcDao.Blacklist(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.BlackList = list
}

func (s *Service) loadHotCache() {
	log.Info("cronLog start loadHotCache")
	tmp, err := s.aiDao.Recommend(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.hotAids = tmp
}

func (s *Service) loadUpRcmdBlockList() {
	blockMap, err := s.spcDao.UpRcmdBlockMap(context.Background())
	if err != nil {
		log.Error("s.spcDao.UpRcmdBlockMap err(%+v)", err)
		return
	}
	s.UpRcmdBlockMap = blockMap
	log.Info("load s.UpRcmdBlockMap size (%d) success", len(s.UpRcmdBlockMap))
}
