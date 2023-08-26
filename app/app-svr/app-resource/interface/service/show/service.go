package show

import (
	"time"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	adtdao "go-gateway/app/app-svr/app-resource/interface/dao/audit"
	"go-gateway/app/app-svr/app-resource/interface/dao/bgroup"
	bubbledao "go-gateway/app/app-svr/app-resource/interface/dao/bubble"
	"go-gateway/app/app-svr/app-resource/interface/dao/cache"
	fidao "go-gateway/app/app-svr/app-resource/interface/dao/fission"
	garbdao "go-gateway/app/app-svr/app-resource/interface/dao/garb"
	locdao "go-gateway/app/app-svr/app-resource/interface/dao/location"
	reddao "go-gateway/app/app-svr/app-resource/interface/dao/red"
	resdao "go-gateway/app/app-svr/app-resource/interface/dao/resource"
	"go-gateway/app/app-svr/app-resource/interface/dao/school"
	tabdao "go-gateway/app/app-svr/app-resource/interface/dao/tab"
	bubblemdl "go-gateway/app/app-svr/app-resource/interface/model/bubble"
	"go-gateway/app/app-svr/app-resource/interface/model/show"
	"go-gateway/app/app-svr/app-resource/interface/model/tab"
	resource "go-gateway/app/app-svr/resource/service/model"

	"go-common/library/log/infoc.v2"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	resourceApi "git.bilibili.co/bapis/bapis-go/resource/service"
	"github.com/robfig/cron"
)

// Service is showtab service.
type Service struct {
	c *conf.Config
	//dao
	rdao        *resdao.Dao
	tdao        *tabdao.Dao
	adt         *adtdao.Dao
	redDao      *reddao.Dao
	bubbleDao   *bubbledao.Dao
	garbDao     *garbdao.Dao
	loc         *locdao.Dao
	tick        time.Duration
	quickerTick time.Duration
	tabCache    map[string][]*show.Tab
	limitsCahce map[int64][]*resource.SideBarLimit
	menuCache   []*tab.Menu
	abtestCache map[string]*resource.AbTest
	showTabMids map[int64]struct{}
	auditCache  map[string]map[int]struct{} // audit mobi_app builds
	redDot      map[int8][]*show.SectionURL
	bubbleCache map[int64]*bubblemdl.Bubble
	skinCache   []*resourceApi.SkinInfo
	fissionDao  *fidao.Dao
	bgroupDao   *bgroup.Dao
	redis       *cache.Dao
	// cron
	cron          *cron.Cron
	accountClient account.AccountClient
	schoolDao     *school.Dao
	infoc         infoc.Infoc
}

// New new a showtab service.
func New(c *conf.Config, ic infoc.Infoc) (s *Service) {
	s = &Service{
		c:           c,
		rdao:        resdao.New(c),
		tdao:        tabdao.New(c),
		adt:         adtdao.New(c),
		redDao:      reddao.New(c),
		bubbleDao:   bubbledao.New(c),
		garbDao:     garbdao.New(c),
		loc:         locdao.New(c),
		fissionDao:  fidao.New(c),
		bgroupDao:   bgroup.New(c),
		schoolDao:   school.New(c),
		redis:       cache.New(c),
		tick:        time.Duration(c.Tick),
		quickerTick: time.Duration(c.QuickerTick),
		tabCache:    map[string][]*show.Tab{},
		limitsCahce: map[int64][]*resource.SideBarLimit{},
		menuCache:   []*tab.Menu{},
		abtestCache: map[string]*resource.AbTest{},
		showTabMids: map[int64]struct{}{},
		auditCache:  map[string]map[int]struct{}{},
		redDot:      map[int8][]*show.SectionURL{},
		bubbleCache: make(map[int64]*bubblemdl.Bubble),
		// cron
		cron:  cron.New(),
		infoc: ic,
	}
	var err error
	if s.accountClient, err = account.NewClient(c.AccountClient); err != nil {
		panic(err)
	}
	s.loadShowTabAids()
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	if err := s.loadCache(); err != nil {
		panic(err)
	}
	s.loadBubbleCache()
	s.loadSkinExtCache()
	var err error
	//nolint:errcheck
	if err = s.cron.AddFunc(s.c.Cron.LoadShowCache, func() { s.loadCache() }); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadBubbleCache, s.loadBubbleCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadSkinExtCache, s.loadSkinExtCache); err != nil {
		panic(err)
	}
}
