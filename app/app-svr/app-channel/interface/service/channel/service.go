package channel

import (
	"context"
	"time"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-channel/interface/conf"
	accdao "go-gateway/app/app-svr/app-channel/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-channel/interface/dao/activity"
	arcdao "go-gateway/app/app-svr/app-channel/interface/dao/archive"
	artdao "go-gateway/app/app-svr/app-channel/interface/dao/article"
	audiodao "go-gateway/app/app-svr/app-channel/interface/dao/audio"
	adtdao "go-gateway/app/app-svr/app-channel/interface/dao/audit"
	bgmdao "go-gateway/app/app-svr/app-channel/interface/dao/bangumi"
	carddao "go-gateway/app/app-svr/app-channel/interface/dao/card"
	chdao "go-gateway/app/app-svr/app-channel/interface/dao/channel"
	convergedao "go-gateway/app/app-svr/app-channel/interface/dao/converge"
	gamedao "go-gateway/app/app-svr/app-channel/interface/dao/game"
	livdao "go-gateway/app/app-svr/app-channel/interface/dao/live"
	locdao "go-gateway/app/app-svr/app-channel/interface/dao/location"
	pediadao "go-gateway/app/app-svr/app-channel/interface/dao/pedia"
	rgdao "go-gateway/app/app-svr/app-channel/interface/dao/region"
	reldao "go-gateway/app/app-svr/app-channel/interface/dao/relation"
	resdao "go-gateway/app/app-svr/app-channel/interface/dao/resource"
	shopdao "go-gateway/app/app-svr/app-channel/interface/dao/shopping"
	specialdao "go-gateway/app/app-svr/app-channel/interface/dao/special"
	tabdao "go-gateway/app/app-svr/app-channel/interface/dao/tab"
	tagdao "go-gateway/app/app-svr/app-channel/interface/dao/tag"
	thumbupdao "go-gateway/app/app-svr/app-channel/interface/dao/thumbup"
	"go-gateway/app/app-svr/app-channel/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model/channel"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"

	"github.com/robfig/cron"
)

// Service channel
type Service struct {
	c *conf.Config
	// dao
	acc        *accdao.Dao
	arc        *arcdao.Dao
	act        *actdao.Dao
	art        *artdao.Dao
	adt        *adtdao.Dao
	bgm        *bgmdao.Dao
	audio      *audiodao.Dao
	rel        *reldao.Dao
	sp         *shopdao.Dao
	tg         *tagdao.Dao
	cd         *carddao.Dao
	ce         *convergedao.Dao
	g          *gamedao.Dao
	sl         *specialdao.Dao
	rg         *rgdao.Dao
	lv         *livdao.Dao
	loc        *locdao.Dao
	tab        *tabdao.Dao
	thumbupDao *thumbupdao.Dao
	chDao      *chdao.Dao
	resDao     *resdao.Dao
	pediaDao   *pediadao.Dao
	// tick
	tick time.Duration
	// cache
	cardCache         map[int64][]*card.Card
	cardPlatCache     map[string][]*card.CardPlat
	upCardCache       map[int64]*operate.Follow
	convergeCardCache map[int64]*operate.Converge
	gameDownloadCache map[int64]*operate.Download
	specialCardCache  map[int64]*operate.Special
	liveCardCache     map[int64][]*live.Card
	cardSetCache      map[int64]*operate.CardSet
	menuCache         map[int64][]*tab.Menu
	// new region list cache
	cachelist   map[string][]*channel.Region
	limitCache  map[int64][]*channel.RegionLimit
	configCache map[int64][]*channel.RegionConfig
	// audit cache
	auditCache map[string]map[int]struct{} // audit mobi_app builds
	// infoc
	logCh chan interface{}
	// cron
	cron *cron.Cron
}

// New channel
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		arc:        arcdao.New(c),
		acc:        accdao.New(c),
		adt:        adtdao.New(c),
		art:        artdao.New(c),
		act:        actdao.New(c),
		bgm:        bgmdao.New(c),
		sp:         shopdao.New(c),
		tg:         tagdao.New(c),
		cd:         carddao.New(c),
		ce:         convergedao.New(c),
		g:          gamedao.New(c),
		sl:         specialdao.New(c),
		rg:         rgdao.New(c),
		audio:      audiodao.New(c),
		lv:         livdao.New(c),
		rel:        reldao.New(c),
		loc:        locdao.New(c),
		tab:        tabdao.New(c),
		thumbupDao: thumbupdao.New(c),
		chDao:      chdao.New(c),
		resDao:     resdao.New(c),
		pediaDao:   pediadao.New(c),
		// tick
		tick: time.Duration(c.Tick),
		// cache
		cardCache:         map[int64][]*card.Card{},
		cardPlatCache:     map[string][]*card.CardPlat{},
		upCardCache:       map[int64]*operate.Follow{},
		convergeCardCache: map[int64]*operate.Converge{},
		gameDownloadCache: map[int64]*operate.Download{},
		specialCardCache:  map[int64]*operate.Special{},
		cachelist:         map[string][]*channel.Region{},
		limitCache:        map[int64][]*channel.RegionLimit{},
		configCache:       map[int64][]*channel.RegionConfig{},
		liveCardCache:     map[int64][]*live.Card{},
		cardSetCache:      map[int64]*operate.CardSet{},
		menuCache:         map[int64][]*tab.Menu{},
		// audit cache
		auditCache: map[string]map[int]struct{}{},
		// infoc
		logCh: make(chan interface{}, 1024),
		// cron
		cron: cron.New(),
	}
	s.loadCache()
	s.initCron()
	s.cron.Start()
	// nolint:biligowordcheck
	go s.infocproc()
	return
}

func (s *Service) loadCache() {
	s.loadAuditCache()
	s.loadRegionlist()
	s.loadCardCache()
	s.loadConvergeCache()
	s.loadSpecialCache()
	s.loadLiveCardCache()
	s.loadGameDownloadCache()
	s.loadCardSetCache()
	s.loadMenusCache()
}

func (s *Service) initCron() {
	var err error
	if err = s.cron.AddFunc(s.c.Cron.LoadAuditCache, s.loadAuditCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadRegionlist, s.loadRegionlist); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadCardCache, s.loadCardCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadConvergeCache, s.loadConvergeCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadSpecialCache, s.loadSpecialCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadLiveCardCache, s.loadLiveCardCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadGameDownloadCache, s.loadGameDownloadCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadCardSetCache, s.loadCardSetCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadMenusCache, s.loadMenusCache); err != nil {
		panic(err)
	}
}

// Ping is check server ping.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.cd.PingDB(c); err != nil {
		return
	}
	return
}

// Archives 存在aids中某个稿件不需要秒开的可能性 因此aids聚合做在外层
func (s *Service) Archives(c context.Context, aidsp []*arcgrpc.PlayAv, isPlayurl bool) (map[int64]*arcgrpc.ArcPlayer, error) {
	if isPlayurl {
		res, err := s.arc.ArcsPlayer(c, aidsp, false)
		if err != nil {
			log.Error("%v", err)
			return nil, err
		}
		return res, nil
	} else {
		var aids []int64
		for _, aidp := range aidsp {
			if aidp != nil && aidp.Aid != 0 {
				aids = append(aids, aidp.Aid)
			}
		}
		tmps, err := s.arc.Arcs(c, aids)
		if err != nil {
			log.Error("%v", err)
			return nil, err
		}
		var res = make(map[int64]*arcgrpc.ArcPlayer)
		for aid, tmp := range tmps {
			if tmp != nil {
				var re = new(arcgrpc.Arc)
				*re = *tmp
				res[aid] = &arcgrpc.ArcPlayer{Arc: re}
			}
		}
		return res, nil
	}
}
