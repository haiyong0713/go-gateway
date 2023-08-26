package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/stat/metric"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	accdao "go-gateway/app/app-svr/app-feed/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-feed/interface/dao/activity"
	addao "go-gateway/app/app-svr/app-feed/interface/dao/ad"
	arcdao "go-gateway/app/app-svr/app-feed/interface/dao/archive"
	artdao "go-gateway/app/app-svr/app-feed/interface/dao/article"
	audiodao "go-gateway/app/app-svr/app-feed/interface/dao/audio"
	bgmdao "go-gateway/app/app-svr/app-feed/interface/dao/bangumi"
	blkdao "go-gateway/app/app-svr/app-feed/interface/dao/black"
	bplusdao "go-gateway/app/app-svr/app-feed/interface/dao/bplus"
	channeldao "go-gateway/app/app-svr/app-feed/interface/dao/channel"
	coindao "go-gateway/app/app-svr/app-feed/interface/dao/coin"
	"go-gateway/app/app-svr/app-feed/interface/dao/creative"
	favdao "go-gateway/app/app-svr/app-feed/interface/dao/favorite"
	fkdao "go-gateway/app/app-svr/app-feed/interface/dao/fawkes"
	"go-gateway/app/app-svr/app-feed/interface/dao/game"
	livdao "go-gateway/app/app-svr/app-feed/interface/dao/live"
	locdao "go-gateway/app/app-svr/app-feed/interface/dao/location"
	"go-gateway/app/app-svr/app-feed/interface/dao/ng"
	rankdao "go-gateway/app/app-svr/app-feed/interface/dao/rank"
	rcmdao "go-gateway/app/app-svr/app-feed/interface/dao/recommend"
	reldao "go-gateway/app/app-svr/app-feed/interface/dao/relation"
	rscdao "go-gateway/app/app-svr/app-feed/interface/dao/resource"
	"go-gateway/app/app-svr/app-feed/interface/dao/search"
	searchdao "go-gateway/app/app-svr/app-feed/interface/dao/search"
	showdao "go-gateway/app/app-svr/app-feed/interface/dao/show"
	tagdao "go-gateway/app/app-svr/app-feed/interface/dao/tag"
	thumbupdao "go-gateway/app/app-svr/app-feed/interface/dao/thumbup"
	"go-gateway/app/app-svr/app-feed/interface/dao/topic"
	tunneldao "go-gateway/app/app-svr/app-feed/interface/dao/tunnel"
	uparc "go-gateway/app/app-svr/app-feed/interface/dao/up-arc"
	updao "go-gateway/app/app-svr/app-feed/interface/dao/upper"
	vipdao "go-gateway/app/app-svr/app-feed/interface/dao/vip"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"
)

var (
	_emptyItem = []*feed.Item{}
)

// Service is show service.
type Service struct {
	c        *conf.Config
	pHit     *prom.Prom
	pMiss    *prom.Prom
	errProm  *prom.Prom
	infoProm *prom.Prom
	// dao
	rcmd   *rcmdao.Dao
	bgm    *bgmdao.Dao
	tg     *tagdao.Dao
	blk    *blkdao.Dao
	lv     *livdao.Dao
	ad     *addao.Dao
	rank   *rankdao.Dao
	show   *showdao.Dao
	audio  *audiodao.Dao
	fawkes *fkdao.Dao
	search *searchdao.Dao
	// rpc
	arc        *arcdao.Dao
	acc        *accdao.Dao
	actDao     *actdao.Dao
	rel        *reldao.Dao
	upper      *updao.Dao
	art        *artdao.Dao
	rsc        *rscdao.Dao
	bplus      *bplusdao.Dao
	loc        *locdao.Dao
	thumbupDao *thumbupdao.Dao
	vip        *vipdao.Dao
	channelDao *channeldao.Dao
	tunnelDao  *tunneldao.Dao
	fav        *favdao.Dao
	coin       *coindao.Dao
	ng         *ng.Dao
	uparc      *uparc.Dao
	creative   *creative.Dao
	topic      *topic.Dao
	game       *game.Dao
	// databus
	cardPub          *databus.Databus
	adFeedPub        *databus.Databus
	sessionRecordPub *databus.Databus
	cron             *cron.Cron
	// audit cache
	auditCache map[string]map[int]struct{} // audit mobi_app builds
	// ai cache
	rcmdCache []*ai.Item
	// rank
	rankCache []*rank.Rank
	// converge cache
	convergeCache map[int64]*operate.Converge
	// download cache
	downloadCache map[int64]*operate.Download
	// special cache
	specialCache map[int64]*operate.Special
	// follow cache
	followCache   map[int64]*operate.Follow
	liveCardCache map[int64][]*live.Card
	// group cache
	// groupCache map[int64]int
	// cache
	cacheCh chan func()
	// infoc
	logCh chan interface{}
	// autoplay mids cache
	autoplayMidsCache map[int64]struct{}
	// follow mode list
	followModeList map[int64]struct{}
	// converge card ai databus
	cardCh chan interface{}
	// fawkes version
	FawkesVersionCache map[string]map[string]*fkmdl.Version
	// new double mids cache
	newDoubleMidsCache map[int64]struct{}
	requestCnt         uint64
	// hot aids
	hotAids    map[int64]struct{}
	fanout     *fanout.Fanout
	infocV2Log infocV2.Infoc
}

// New new a show service.
// nolint: biligowordcheck
func New(c *conf.Config, ic infocV2.Infoc) (s *Service) {
	s = &Service{
		c:        c,
		pHit:     prom.CacheHit,
		pMiss:    prom.CacheMiss,
		errProm:  prom.BusinessErrCount,
		infoProm: prom.BusinessInfoCount,
		// dao
		rcmd:   rcmdao.New(c),
		bgm:    bgmdao.New(c),
		tg:     tagdao.New(c),
		blk:    blkdao.New(c),
		lv:     livdao.New(c),
		ad:     addao.New(c),
		rank:   rankdao.New(c),
		show:   showdao.New(c),
		audio:  audiodao.New(c),
		bplus:  bplusdao.New(c),
		fawkes: fkdao.New(c),
		search: search.New(c),
		// rpc
		arc:        arcdao.New(c),
		rel:        reldao.New(c),
		acc:        accdao.New(c),
		actDao:     actdao.New(c),
		upper:      updao.New(c),
		art:        artdao.New(c),
		rsc:        rscdao.New(c),
		loc:        locdao.New(c),
		thumbupDao: thumbupdao.New(c),
		vip:        vipdao.New(c),
		channelDao: channeldao.New(c),
		tunnelDao:  tunneldao.New(c),
		fav:        favdao.New(c),
		coin:       coindao.New(c),
		cron:       cron.New(),
		ng:         ng.New(c),
		uparc:      uparc.New(c),
		creative:   creative.New(c),
		topic:      topic.New(c),
		game:       game.New(c),
		// databus
		cardPub:          databus.New(c.CardDatabus),
		adFeedPub:        databus.New(c.CardAdFeedDatabus),
		sessionRecordPub: databus.New(c.SessionRecordDatabus),
		// cache
		cacheCh: make(chan func(), 1024),
		// infoc
		logCh: make(chan interface{}, 1024),
		// autoplay mids cache
		autoplayMidsCache: map[int64]struct{}{},
		// converge card ai databus
		cardCh: make(chan interface{}, 1024),
		// fawkes
		FawkesVersionCache: make(map[string]map[string]*fkmdl.Version),
		// new double mids cache
		newDoubleMidsCache: map[int64]struct{}{},
		// hot aids
		hotAids:    map[int64]struct{}{},
		fanout:     fanout.New("cache"),
		infocV2Log: ic,
	}
	s.loadCache()
	s.cron.Start()
	go s.cacheproc()
	go s.infocproc()
	return
}

func (s *Service) addCache(f func()) {
	select {
	case s.cacheCh <- f:
	default:
		log.Warn("cacheproc chan full")
	}
}

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

// RecordSession is
// nolint: bilirailguncheck
func (s *Service) RecordSession(si *session.IndexSession) {
	if err := s.sessionRecordPub.Send(context.Background(), si.ID, si); err != nil {
		log.Error("Failed to send index sesson to databus: %+v", err)
	}
}

func (s *Service) loadCache() {
	s.loadAuditCache()
	s.loadRcmdCache()
	s.loadRankCache()
	s.loadConvergeCache()
	s.loadDownloadCache()
	s.loadSpecialCache()
	s.loadUpCardCache()
	s.loadAutoPlayMid()
	s.loadFollowModeList()
	s.loadFawkes()
	s.loadRcmdHotCache()
	s.loadLiveCardCache()                                                 // 1s
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadAuditCache))        // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadRcmdCache))         // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadRankCache))         // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadConvergeCache))     // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadDownloadCache))     // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadSpecialCache))      // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadUpCardCache))       // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadAutoPlayMid))       // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadFollowModeList))    // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadFawkes))            // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadCache, s.loadRcmdHotCache))      // 间隔1分钟
	checkErr(s.cron.AddFunc(s.c.Cron.LoadLiveCache, s.loadLiveCardCache)) // 间隔1秒钟
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("cron add func loadCache error(%+v)", err))
	}
}

// Close
func (s *Service) Close() {
	s.cron.Stop()
	s.fanout.Close()
}

var MetricFeedIndexReqDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
	Namespace: "feed_index",
	Subsystem: "requests",
	Name:      "duration",
	Help:      "http server requests duration(ms).",
	Labels:    []string{"mobi_app", "login_event"},
	Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
})

func (s *Service) CustomizedMetric() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		now := time.Now()
		ctx.Next()

		dt := time.Since(now)
		MetricFeedIndexReqDuration.Observe(int64(dt/time.Millisecond), ctx.Request.Form.Get("mobi_app"), ctx.Request.Form.Get("login_event"))
	}
}
