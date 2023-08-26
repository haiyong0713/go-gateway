package show

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/railgun"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	"go-common/library/xstr"

	populardao "go-gateway/app/app-svr/app-show/interface/dao/popular"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-show/interface/conf"
	accdao "go-gateway/app/app-svr/app-show/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-show/interface/dao/activity"
	addao "go-gateway/app/app-svr/app-show/interface/dao/ad"
	arcdao "go-gateway/app/app-svr/app-show/interface/dao/archive"
	artdao "go-gateway/app/app-svr/app-show/interface/dao/article"
	adtdao "go-gateway/app/app-svr/app-show/interface/dao/audit"
	bgmdao "go-gateway/app/app-svr/app-show/interface/dao/bangumi"
	carddao "go-gateway/app/app-svr/app-show/interface/dao/card"
	condao "go-gateway/app/app-svr/app-show/interface/dao/control"
	dbusdao "go-gateway/app/app-svr/app-show/interface/dao/databus"
	dyndao "go-gateway/app/app-svr/app-show/interface/dao/dynamic"
	"go-gateway/app/app-svr/app-show/interface/dao/favorite"
	livedao "go-gateway/app/app-svr/app-show/interface/dao/live"
	locdao "go-gateway/app/app-svr/app-show/interface/dao/location"
	rcmmdao "go-gateway/app/app-svr/app-show/interface/dao/recommend"
	regiondao "go-gateway/app/app-svr/app-show/interface/dao/region"
	reldao "go-gateway/app/app-svr/app-show/interface/dao/relation"
	resdao "go-gateway/app/app-svr/app-show/interface/dao/resource"
	showdao "go-gateway/app/app-svr/app-show/interface/dao/show"
	tagdao "go-gateway/app/app-svr/app-show/interface/dao/tag"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	recmod "go-gateway/app/app-svr/app-show/interface/model/recommend"
	"go-gateway/app/app-svr/app-show/interface/model/region"
	"go-gateway/app/app-svr/app-show/interface/model/show"
	resource "go-gateway/app/app-svr/resource/service/model"

	tagApi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	bgroupApi "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
)

type recommend struct {
	key  string
	aids []int64
}

type rcmmndCfg struct {
	Aid   int64  `json:"aid"`
	Goto  string `json:"goto"`
	Title string `json:"title"`
	Cover string `json:"cover"`
}

// Service is show service.
type Service struct {
	c              *conf.Config
	creativeClient creativeAPI.VideoUpOpenClient
	dao            *showdao.Dao
	rcmmnd         *rcmmdao.Dao
	ad             *addao.Dao // cptbanner
	bgm            *bgmdao.Dao
	lv             *livedao.Dao
	// url
	blackURL string
	// bnnr   *bnnrdao.Dao
	adt  *adtdao.Dao
	tag  *tagdao.Dao
	arc  *arcdao.Dao
	dbus *dbusdao.Dao
	dyn  *dyndao.Dao
	// controldao
	controld *condao.Dao
	res      *resdao.Dao
	// artic   *articledao.Dao
	client  *httpx.Client
	rg      *regiondao.Dao
	cdao    *carddao.Dao
	favdao  *favorite.Dao
	act     *actdao.Dao
	acc     *accdao.Dao
	popular *populardao.Dao
	// relation
	reldao *reldao.Dao
	loc    *locdao.Dao
	// art
	art *artdao.Dao
	// cache
	rcmmndCache         []*show.Item
	rcmmndOseaCache     []*show.Item
	regionCache         map[string][]*show.Item
	regionOseaCache     map[string][]*show.Item
	regionBgCache       map[string][]*show.Item
	regionBgOseaCache   map[string][]*show.Item
	regionBgEpCache     map[string][]*show.Item
	regionBgEpOseaCache map[string][]*show.Item
	bgmCache            map[int8][]*show.Item
	liveCount           int
	liveMoeCache        []*show.Item // TODO change to liveMoeCache
	liveHotCache        []*show.Item // TODO change to liveHotCache
	bannerCache         map[int8]map[int][]*resource.Banner
	cache               map[string][]*show.Show
	cacheBg             map[string][]*show.Show
	cacheBgEp           map[string][]*show.Show
	tempCache           map[string][]*show.Show
	auditCache          map[string]map[int]struct{} // audit mobi_app builds
	blackCache          map[int64]struct{}          // black aids

	logCh     chan infoc
	logFeedCh chan interface{}
	rcmmndCh  chan recommend
	logPath   string

	// loadfile
	jsonOn bool
	// cpm percentage   0~100
	cpmNum       int
	cpmMid       map[int64]struct{}
	cpmAll       bool
	cpmRcmmndNum int
	cpmRcmmndMid map[int64]struct{}
	cpmRcmmndAll bool
	adIsPost     bool
	// recommend api
	rcmmndOn    bool
	rcmmndGroup map[int64]int    // mid -> group
	rcmmndHosts map[int][]string // group -> hosts
	// region
	reRegionCache map[int]*region.Region
	// ranking
	rankCache     []*show.Item
	rankOseaCache []*show.Item
	// card
	cardCache       map[string][]*show.Show
	columnListCache map[int]*card.ColumnList
	cardSetCache    map[int64]*operate.CardSet
	eventTopicCache map[int64]*operate.EventTopic
	// hot card
	hotTenTabCardCache map[int][]*recmod.CardList
	rankAidsCache      []int64
	rankScoreCache     map[int64]int64
	rankArchivesCache  map[int64]*api.Arc
	// hotCache           []*card.PopularCard
	rcmdCache  []*card.PopularCard
	rankCache2 []*rank.Rank
	// prom
	pHit  *prom.Prom
	pMiss *prom.Prom
	// fanout
	fanout         *fanout.Fanout
	topEntrance    []*show.EntranceMem
	largeCards     map[int64]*show.LargeCard
	largeCardsMids map[int64]struct{}
	// good history cache
	// goodHistoryCache []*show.GoodHisRes
	middleTopPhoto     string
	liveCards          map[int64]*show.LiveCard
	articleCards       map[int64]*show.ArticleCard
	tagClient          tagApi.TagRPCClient
	showServiceRailGun *railgun.Railgun
	cacheRailGun       *railgun.Railgun
	cardJobRailGun     *railgun.Railgun
	infocv2            infocv2.Infoc
	feedTabInfocv2     infocv2.Infoc
	bGroupClient       bgroupApi.BGroupServiceClient
}

// New new a show service.
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	rcmmndHosts := make(map[int][]string, len(c.Recommend.Host))
	for k, v := range c.Recommend.Host {
		key, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		rcmmndHosts[key] = v
	}
	rcmmndGroup := make(map[int64]int, len(c.Recommend.Group))
	for k, v := range c.Recommend.Group {
		key, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		rcmmndGroup[int64(key)] = v
	}
	s = &Service{
		c:        c,
		dao:      showdao.New(c),
		rcmmnd:   rcmmdao.New(c),
		ad:       addao.New(c),
		bgm:      bgmdao.New(c),
		lv:       livedao.New(c),
		controld: condao.New(c),
		// url
		blackURL: c.Host.Black + _blackUrl,
		// bnnr:   bnnrdao.New(c),
		adt:  adtdao.New(c),
		tag:  tagdao.New(c),
		arc:  arcdao.New(c),
		dbus: dbusdao.New(c),
		dyn:  dyndao.New(c),
		res:  resdao.New(c),
		// artic:   articledao.New(c),
		rg:      regiondao.New(c),
		cdao:    carddao.New(c),
		favdao:  favorite.New(c),
		act:     actdao.New(c),
		acc:     accdao.New(c),
		popular: populardao.New(c),
		// relation
		reldao: reldao.New(c),
		loc:    locdao.New(c),
		art:    artdao.New(c),
		client: httpx.NewClient(c.HTTPData),
		// cache
		jsonOn: false,

		logCh:     make(chan infoc, 1024),
		logFeedCh: make(chan interface{}, 1024),
		rcmmndCh:  make(chan recommend, 1024),
		logPath:   c.ShowLog,

		rcmmndOn:    false,
		rcmmndGroup: rcmmndGroup,
		rcmmndHosts: rcmmndHosts,
		// cpm percentage   0~100
		cpmNum:       0,
		cpmMid:       map[int64]struct{}{},
		cpmAll:       true,
		cpmRcmmndNum: 0,
		cpmRcmmndMid: map[int64]struct{}{},
		cpmRcmmndAll: true,
		adIsPost:     false,
		// region
		reRegionCache: map[int]*region.Region{},
		// ranking
		rankCache:     []*show.Item{},
		rankOseaCache: []*show.Item{},
		// card
		cardCache:       map[string][]*show.Show{},
		columnListCache: map[int]*card.ColumnList{},
		cardSetCache:    map[int64]*operate.CardSet{},
		eventTopicCache: map[int64]*operate.EventTopic{},
		// hot card
		hotTenTabCardCache: make(map[int][]*recmod.CardList),
		rankAidsCache:      []int64{},
		rankScoreCache:     map[int64]int64{},
		rankArchivesCache:  map[int64]*api.Arc{},
		// hotCache:           []*card.PopularCard{},
		rcmdCache:  []*card.PopularCard{},
		rankCache2: []*rank.Rank{},
		// prom
		pHit:        prom.CacheHit,
		pMiss:       prom.CacheMiss,
		fanout:      fanout.New("cache"),
		topEntrance: []*show.EntranceMem{},
		// cards cache
		articleCards: map[int64]*show.ArticleCard{},
	}
	var err error
	if s.creativeClient, err = creativeAPI.NewClient(nil); err != nil {
		panic("creativeGRPC not found!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
	if s.tagClient, err = tagApi.NewClient(nil); err != nil {
		panic(err)
	}
	if s.infocv2, err = infocv2.New(c.Infocv2.Infoc); err != nil {
		panic(err)
	}
	if s.feedTabInfocv2, err = infocv2.New(c.FeedTabInfocv2.Infoc); err != nil {
		panic(err)
	}
	if s.bGroupClient, err = bgroupApi.NewClient(c.BGroupGRPC); err != nil {
		panic(err)
	}

	s.loadPopEntrances()
	s.loadMiddTopPhoto()
	s.loadCache(time.Now())
	s.loadLargeCards()
	s.loadLiveCards()
	s.initShowServiceRailGun()
	s.initCacheRailGun()
	s.initCardRailGun()
	go s.infocproc()
	go s.rcmmndproc()
	go s.infocfeedproc()
	return s
}

func (s *Service) initShowServiceRailGun() {
	r := railgun.NewRailGun("showServiceRailGun", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "*/3 * * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadPopEntrances()
		s.loadMiddTopPhoto()
		return railgun.MsgPolicyNormal
	}))
	s.showServiceRailGun = r
	r.Start()
}

func (s *Service) initCacheRailGun() {
	r := railgun.NewRailGun("ShowCacheRailGun", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "0 */3 * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadCache(time.Now())
		return railgun.MsgPolicyNormal
	}))
	s.cacheRailGun = r
	r.Start()
}

func (s *Service) initCardRailGun() {
	r := railgun.NewRailGun("cardRailGun", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "*/3 * * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadLargeCards()
		s.loadLiveCards()
		return railgun.MsgPolicyNormal
	}))
	s.cardJobRailGun = r
	r.Start()
}

func (s *Service) loadPopEntrances() {
	res, err := s.dao.Entrances(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Popular Entrance err %v", err)
		return
	}
	s.topEntrance = res
}

func (s *Service) loadMiddTopPhoto() {
	res, err := s.dao.MidTopPhoto(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Middle TopPhoto err %v", err)
		return
	}
	s.middleTopPhoto = res
}

func (s *Service) loadLargeCards() {
	res, err := s.dao.LargeCards(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Popular LargeCards err %v len(%d)", err, len(res))
		return
	}
	var largeCardMaps = make(map[int64]*show.LargeCard)
	var mids = make(map[int64]struct{})
	for _, item := range res {
		largeCardMaps[item.ID] = item
		if item.WhiteList == "" {
			continue
		}
		ids, _ := xstr.SplitInts(item.WhiteList)
		for _, id := range ids {
			if _, ok := mids[id]; !ok {
				mids[id] = struct{}{}
			}
		}
	}
	s.largeCards = largeCardMaps
	s.largeCardsMids = mids
}

func (s *Service) loadLiveCards() {
	res, err := s.dao.LiveCards(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Popular LiveCards err %v len(%d)", err, len(res))
		return
	}
	var liveCardMaps = make(map[int64]*show.LiveCard)
	for _, item := range res {
		liveCardMaps[item.ID] = item
	}
	s.liveCards = liveCardMaps
}

func (s *Service) loadCache(now time.Time) {
	s.loadRcmmndCache(now)
	s.loadRegionCache(now)
	s.loadBgmCache(now)
	if !s.c.Custom.LiveRcmdClose {
		s.loadLiveCache(now)
	}
	s.loadBannerCache()
	s.loadShowCache()
	s.loadShowTempCache()
	s.loadBlackCache()
	s.loadAuditCache()
	s.loadRegionListCache()
	s.loadRankAllCache()
	s.loadColumnListCache()
	s.loadCardCache(now)
	s.loadCardSetCache()
	// s.loadPopularCard(now)
	s.loadEventTopicCache()
	s.loadArticleCardsCache()
}

// SetRcmmndOn
func (s *Service) SetRcmmndOn(on bool) {
	s.rcmmndOn = on
}

// GetRcmmndOn
func (s *Service) RcmmndOn() bool {
	return s.rcmmndOn
}

// Close dao
func (s *Service) Close() error {
	s.showServiceRailGun.Close()
	s.cacheRailGun.Close()
	s.cardJobRailGun.Close()
	return s.dao.Close()
}

// SetRcmmndGroup set rcmmnd group data.
func (s *Service) SetRcmmndGroup(m int64, g int) {
	tmp := map[int64]int{}
	tmp[m] = g
	for k, v := range s.rcmmndGroup {
		if k != m {
			tmp[k] = v
		}
	}
	s.rcmmndGroup = tmp
}

// GetRcmmndGroup get rcmmnd group data.
func (s *Service) GetRcmmndGroup() map[string]int {
	tmp := map[string]int{}
	for k, v := range s.rcmmndGroup {
		tmp[strconv.FormatInt(k, 10)] = v
	}
	return tmp
}

// SetRcmmndHost set rcmmnd host data.
func (s *Service) SetRcmmndHost(g int, hosts []string) {
	tmp := map[int][]string{}
	tmp[g] = hosts
	for k, v := range s.rcmmndHosts {
		if k != g {
			tmp[k] = v
		}
	}
	s.rcmmndHosts = tmp
}

// GetRcmmndHost get rcmmnd host data.
func (s *Service) GetRcmmndHost() map[string][]string {
	tmp := map[string][]string{}
	for k, v := range s.rcmmndHosts {
		tmp[strconv.Itoa(k)] = v
	}
	return tmp
}

// SetCpm percentage  0~100
// nolint:gomnd
func (s *Service) SetCpmNum(num int) {
	s.cpmNum = num
	if s.cpmNum < 0 {
		s.cpmNum = 0
	} else if s.cpmNum > 100 {
		s.cpmNum = 100
	}
}

// GetCpm percentage
func (s *Service) CpmNum() int {
	return s.cpmNum
}

// SetCpm percentage  0~100
func (s *Service) SetCpmMid(mid int64) {
	var mids = map[int64]struct{}{}
	mids[mid] = struct{}{}
	for mid := range s.cpmMid {
		if _, ok := mids[mid]; !ok {
			mids[mid] = struct{}{}
		}
	}
	s.cpmMid = mids
}

// GetCpm percentage
func (s *Service) CpmMid() []int {
	var mids []int
	for mid := range s.cpmMid {
		mids = append(mids, int(mid))
	}
	return mids
}

// SetCpm All
func (s *Service) SetCpmAll(isAll bool) {
	s.cpmAll = isAll
}

// GetCpm All
func (s *Service) CpmAll() int {
	if s.cpmAll {
		return 1
	}
	return 0
}

// RcmmndNum percentage
func (s *Service) RcmmndNum() int {
	return s.cpmRcmmndNum
}

// SetRcmmndNum percentage  0~100
// nolint:gomnd
func (s *Service) SetRcmmndNum(num int) {
	s.cpmRcmmndNum = num
	if s.cpmRcmmndNum < 0 {
		s.cpmRcmmndNum = 0
	} else if s.cpmRcmmndNum > 100 {
		s.cpmRcmmndNum = 100
	}
}

// CpmRcmmndMid Mid
func (s *Service) CpmRcmmndMid() []int {
	var mids []int
	for mid := range s.cpmRcmmndMid {
		mids = append(mids, int(mid))
	}
	return mids
}

// SetCpmRcmmndMid Mid
func (s *Service) SetCpmRcmmndMid(mid int64) {
	var mids = map[int64]struct{}{}
	mids[mid] = struct{}{}
	for mid := range s.cpmRcmmndMid {
		if _, ok := mids[mid]; !ok {
			mids[mid] = struct{}{}
		}
	}
	s.cpmRcmmndMid = mids
}

// CpmRcmmnd All
func (s *Service) CpmRcmmndAll() int {
	if s.cpmRcmmndAll {
		return 1
	}
	return 0
}

// SetCpmRcmmnd All
func (s *Service) SetCpmRcmmndAll(isAll bool) {
	s.cpmRcmmndAll = isAll
}

// SetIsPost Get or Post
func (s *Service) SetAdIsPost(isPost bool) {
	s.adIsPost = isPost
}

// IsPost Get or Post
func (s *Service) AdIsPost() int {
	if s.adIsPost {
		return 1
	}
	return 0
}
