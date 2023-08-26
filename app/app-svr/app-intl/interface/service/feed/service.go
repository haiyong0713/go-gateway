package feed

import (
	"time"

	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/stat/prom"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	accdao "go-gateway/app/app-svr/app-intl/interface/dao/account"
	arcdao "go-gateway/app/app-svr/app-intl/interface/dao/archive"
	bgmdao "go-gateway/app/app-svr/app-intl/interface/dao/bangumi"
	blkdao "go-gateway/app/app-svr/app-intl/interface/dao/black"
	carddao "go-gateway/app/app-svr/app-intl/interface/dao/card"
	cfcdao "go-gateway/app/app-svr/app-intl/interface/dao/content"
	fkdao "go-gateway/app/app-svr/app-intl/interface/dao/fawkes"
	locdao "go-gateway/app/app-svr/app-intl/interface/dao/location"
	rankdao "go-gateway/app/app-svr/app-intl/interface/dao/rank"
	rcmdao "go-gateway/app/app-svr/app-intl/interface/dao/recommend"
	reldao "go-gateway/app/app-svr/app-intl/interface/dao/relation"
	rscdao "go-gateway/app/app-svr/app-intl/interface/dao/resource"
	tagdao "go-gateway/app/app-svr/app-intl/interface/dao/tag"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"
)

// Service is show service.
type Service struct {
	c     *conf.Config
	pHit  *prom.Prom
	pMiss *prom.Prom
	// dao
	rcmd *rcmdao.Dao
	tg   *tagdao.Dao
	bgm  *bgmdao.Dao
	blk  *blkdao.Dao
	rank *rankdao.Dao
	card *carddao.Dao
	// rpc
	arc    *arcdao.Dao
	acc    *accdao.Dao
	rel    *reldao.Dao
	loc    *locdao.Dao
	rsc    *rscdao.Dao
	fawkes *fkdao.Dao
	cfc    *cfcdao.Dao
	// tick
	tick time.Duration
	// black cache
	blackCache map[int64]struct{} // black aids
	// ai cache
	rcmdCache []*ai.Item
	// rank cache
	rankCache []*rank.Rank
	// follow cache
	followCache map[int64]*operate.Follow
	// tab cache
	menuCache  []*operate.Menu
	tabCache   map[int64][]*operate.Active
	coverCache map[int64]string
	// converge cache
	convergeCache map[int64]*operate.Converge
	// special cache
	specialCache map[int64]*operate.Special
	// group cache
	groupCache map[int64]int
	// cache
	cacheCh chan func()
	// infoc
	logCh         chan interface{}
	infocV2Client infocV2.Infoc
	// fawkes version
	FawkesVersionCache map[string]map[string]*fkmdl.Version
	// audit cache
	auditCache map[string]map[int]struct{} // audit mobi_app builds
}

// New new a show service.
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		pHit:  prom.CacheHit,
		pMiss: prom.CacheMiss,
		// dao
		rcmd: rcmdao.New(c),
		blk:  blkdao.New(c),
		rank: rankdao.New(c),
		tg:   tagdao.New(c),
		bgm:  bgmdao.New(c),
		card: carddao.New(c),
		// rpc
		arc:    arcdao.New(c),
		rel:    reldao.New(c),
		acc:    accdao.New(c),
		loc:    locdao.New(c),
		rsc:    rscdao.New(c),
		fawkes: fkdao.New(c),
		cfc:    cfcdao.New(c),
		// tick
		tick: time.Duration(c.Tick),
		// group cache
		groupCache: map[int64]int{},
		// cache
		cacheCh: make(chan func(), 1024),
		// infoc
		logCh: make(chan interface{}, 1024),
		// fawkes
		FawkesVersionCache: make(map[string]map[string]*fkmdl.Version),
	}
	var err error
	if s.infocV2Client, err = infocV2.New(s.c.FeedInfocv2.Conf); err != nil {
		panic(err)
	}
	s.loadBlackCache()
	s.loadRcmdCache()
	s.loadRankCache()
	s.loadUpCardCache()
	s.loadGroupCache()
	s.loadFawkes()
	s.loadAuditCache()
	s.loadTabCache()
	s.loadConvergeCache()
	s.loadSpecialCache()
	go s.cacheproc()
	go s.blackproc()
	go s.rcmdproc()
	go s.rankproc()
	go s.upCardproc()
	go s.groupproc()
	go s.infocproc()
	go s.loadFawkesProc()
	go s.auditproc()
	go s.tabproc()
	go s.convergeproc()
	go s.specialproc()
	return
}

// addCache is.
func (s *Service) addCache(f func()) {
	select {
	case s.cacheCh <- f:
	default:
		log.Warn("cacheproc chan full")
	}
}

// cacheproc is.
func (s *Service) cacheproc() {
	for {
		f, ok := <-s.cacheCh
		if !ok {
			log.Warn("cache proc exit")
			return
		}
		f()
	}
}
