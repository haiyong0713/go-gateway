package dynamicV2

import (
	"context"

	"go-common/component/tinker"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/account"
	activitydao "go-gateway/app/app-svr/app-dynamic/interface/dao/activity"
	aidao "go-gateway/app/app-svr/app-dynamic/interface/dao/ai"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/archive"
	articledao "go-gateway/app/app-svr/app-dynamic/interface/dao/article"
	addao "go-gateway/app/app-svr/app-dynamic/interface/dao/bcg"
	channeldao "go-gateway/app/app-svr/app-dynamic/interface/dao/channel"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/cheese"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/comic"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/community"
	dramaseasondao "go-gateway/app/app-svr/app-dynamic/interface/dao/dramaseason"
	dyndao "go-gateway/app/app-svr/app-dynamic/interface/dao/dynamicV2"
	esdao "go-gateway/app/app-svr/app-dynamic/interface/dao/es"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/esport"
	favdao "go-gateway/app/app-svr/app-dynamic/interface/dao/favorite"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/game"
	garbdao "go-gateway/app/app-svr/app-dynamic/interface/dao/garb"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/live"
	locdao "go-gateway/app/app-svr/app-dynamic/interface/dao/location"
	medialistdao "go-gateway/app/app-svr/app-dynamic/interface/dao/medialist"
	musicdao "go-gateway/app/app-svr/app-dynamic/interface/dao/music"
	pangudao "go-gateway/app/app-svr/app-dynamic/interface/dao/pangu"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/pgc"
	platformdao "go-gateway/app/app-svr/app-dynamic/interface/dao/platform"
	playurldao "go-gateway/app/app-svr/app-dynamic/interface/dao/playurl"
	rcmddao "go-gateway/app/app-svr/app-dynamic/interface/dao/recommend"
	relationdao "go-gateway/app/app-svr/app-dynamic/interface/dao/relation"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/resource"
	sharedao "go-gateway/app/app-svr/app-dynamic/interface/dao/share"
	shopdao "go-gateway/app/app-svr/app-dynamic/interface/dao/shopping"
	subdao "go-gateway/app/app-svr/app-dynamic/interface/dao/subscription"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/thumbup"
	topdao "go-gateway/app/app-svr/app-dynamic/interface/dao/topic"
	ugcseasondao "go-gateway/app/app-svr/app-dynamic/interface/dao/ugcSeason"
	updao "go-gateway/app/app-svr/app-dynamic/interface/dao/up"
	"go-gateway/app/app-svr/app-dynamic/interface/dao/videoup"

	appfeaturegate "git.bilibili.co/evergarden/feature-gate/app-featuregate"
	infocV2 "go-common/library/log/infoc.v2"

	"github.com/robfig/cron"
)

type Service struct {
	// 全局功能控制
	appFeatureGate appfeaturegate.AppFeatureGate
	c              *conf.Config
	dynDao         *dyndao.Dao
	archiveDao     *archive.Dao
	pgcDao         *pgc.Dao
	accountDao     *account.Dao
	thumDao        *thumbup.Dao
	cheeseDao      *cheese.Dao
	cmtDao         *community.Dao
	esportDao      *esport.Dao
	comicDao       *comic.Dao
	gameDao        *game.Dao
	videoupDao     *videoup.Dao
	liveDao        *live.Dao
	articleDao     *articledao.Dao
	musicDao       *musicdao.Dao
	medialistDao   *medialistdao.Dao
	adDao          *addao.Dao
	subDao         *subdao.Dao
	relationDao    *relationdao.Dao
	ugcSeasonDao   *ugcseasondao.Dao
	garbDao        *garbdao.Dao
	favDao         *favdao.Dao
	activityDao    *activitydao.Dao
	aiDao          *aidao.Dao
	channelDao     *channeldao.Dao
	platformDao    *platformdao.Dao
	shareDao       *sharedao.Dao
	esDao          *esdao.Dao
	topDao         *topdao.Dao
	upDao          *updao.Dao
	rcmdDao        *rcmddao.Dao
	dramaseasonDao *dramaseasondao.Dao
	playurlDao     *playurldao.Dao
	loc            *locdao.Dao
	shopDao        *shopdao.Dao
	panguDao       *pangudao.Dao
	resourceDao    *resource.Dao
	// bottom config map
	bottomMap map[string]*conf.BottomItem
	// Hot video dict
	resRcmd map[int64]struct{}
	// 定时任务
	cron *cron.Cron
	// infoc
	infocV2 infocV2.Infoc
	// abtest
	tinker *tinker.ABTest
	// infoc
	logCh chan interface{}
}

func New(c *conf.Config, infoc infocV2.Infoc) (s *Service) {
	s = &Service{
		appFeatureGate: appfeaturegate.GetFeatureConf(),
		c:              c,
		dynDao:         dyndao.New(c),
		archiveDao:     archive.New(c),
		pgcDao:         pgc.New(c),
		accountDao:     account.New(c),
		thumDao:        thumbup.New(c),
		cheeseDao:      cheese.New(c),
		cmtDao:         community.New(c),
		esportDao:      esport.New(c),
		comicDao:       comic.New(c),
		videoupDao:     videoup.New(c),
		liveDao:        live.New(c),
		gameDao:        game.New(c),
		articleDao:     articledao.New(c),
		musicDao:       musicdao.New(c),
		medialistDao:   medialistdao.New(c),
		adDao:          addao.New(c),
		subDao:         subdao.New(c),
		relationDao:    relationdao.New(c),
		ugcSeasonDao:   ugcseasondao.New(c),
		garbDao:        garbdao.New(c),
		favDao:         favdao.New(c),
		activityDao:    activitydao.New(c),
		aiDao:          aidao.New(c),
		channelDao:     channeldao.New(c),
		platformDao:    platformdao.New(c),
		shareDao:       sharedao.New(c),
		esDao:          esdao.New(c),
		topDao:         topdao.New(c),
		upDao:          updao.New(c),
		rcmdDao:        rcmddao.New(c),
		dramaseasonDao: dramaseasondao.New(c),
		playurlDao:     playurldao.New(c),
		cron:           cron.New(),
		loc:            locdao.New(c),
		shopDao:        shopdao.New(c),
		panguDao:       pangudao.New(c),
		resourceDao:    resource.New(c),
		resRcmd:        make(map[int64]struct{}),
		// infoc
		infocV2: infoc,
		// infoc
		logCh: make(chan interface{}, 1024),
	}
	s.tinker = tinker.Init(s.infocV2, nil)
	s.bottomMap = makeBottomMap(c)
	s.initCron()
	s.cron.Start()
	// infoc上报
	// nolint:biligowordcheck
	go s.infocproc()
	return s
}

func (s *Service) initCron() {
	s.loadRecommend()
	var err error
	if err = s.cron.AddFunc(s.c.Tick.LoadHotVideo, s.loadRecommend); err != nil {
		panic(err)
	}
}

func makeBottomMap(c *conf.Config) map[string]*conf.BottomItem {
	if c.BottomConfig == nil || c.BottomConfig.TopicJumpLinks == nil {
		return nil
	}
	var res = make(map[string]*conf.BottomItem)
	for _, bottom := range c.BottomConfig.TopicJumpLinks {
		btmTmp := bottom
		if len(bottom.RelatedTopic) == 0 {
			continue
		}
		for _, topic := range bottom.RelatedTopic {
			res[topic] = &btmTmp
		}
	}
	return res
}

func (s *Service) loadRecommend() {
	var (
		rcmd map[int64]struct{}
		err  error
	)
	if rcmd, err = s.aiDao.Recommend(context.Background()); err != nil {
		log.Error("cron Recommend error %v", err)
		return
	}
	s.resRcmd = rcmd
}
