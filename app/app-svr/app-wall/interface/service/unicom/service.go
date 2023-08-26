package unicom

import (
	"go-common/library/log/infoc.v2"
	"go-common/library/queue/databus"
	"go-common/library/silverbullet/gaia"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	accDao "go-gateway/app/app-svr/app-wall/interface/dao/account"
	comicDao "go-gateway/app/app-svr/app-wall/interface/dao/comic"
	liveDao "go-gateway/app/app-svr/app-wall/interface/dao/live"
	locDao "go-gateway/app/app-svr/app-wall/interface/dao/location"
	lockDao "go-gateway/app/app-svr/app-wall/interface/dao/redis_lock"
	seqDao "go-gateway/app/app-svr/app-wall/interface/dao/seq"
	shopDao "go-gateway/app/app-svr/app-wall/interface/dao/shopping"
	unicomDao "go-gateway/app/app-svr/app-wall/interface/dao/unicom"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"

	"github.com/robfig/cron"
)

type Service struct {
	c               *conf.Config
	dao             *unicomDao.Dao
	live            *liveDao.Dao
	seqdao          *seqDao.Dao
	accd            *accDao.Dao
	shop            *shopDao.Dao
	locdao          *locDao.Dao
	comicdao        *comicDao.Dao
	lockdao         *lockDao.Dao
	unicomIpCache   []*unicom.UnicomIP
	unicomPackCache []*unicom.UserPack
	// infoc
	userBindCh chan interface{}
	// databus
	userbindPub *databus.Databus
	packPub     *databus.Databus
	// prom
	pHit     *prom.Prom
	pMiss    *prom.Prom
	infoProm *prom.Prom
	// infoc
	infocV2Log infoc.Infoc
	// cache
	cron       *cron.Cron
	cache      *fanout.Fanout
	gaiaEngine *gaia.GaiaEngine
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:               c,
		dao:             unicomDao.New(c),
		live:            liveDao.New(c),
		seqdao:          seqDao.New(c),
		accd:            accDao.New(c),
		shop:            shopDao.New(c),
		locdao:          locDao.New(c),
		lockdao:         lockDao.New(c),
		comicdao:        comicDao.New(c),
		unicomIpCache:   []*unicom.UnicomIP{},
		unicomPackCache: []*unicom.UserPack{},
		// databus
		userbindPub: databus.New(c.UnicomDatabus),
		packPub:     databus.New(c.PackPub),
		// infoc
		userBindCh: make(chan interface{}, 1024),
		// prom
		pHit:     prom.CacheHit,
		pMiss:    prom.CacheMiss,
		infoProm: prom.BusinessInfoCount,
		// cache
		cache: fanout.New("cache", fanout.Worker(2), fanout.Buffer(10240)),
		cron:  cron.New(),
	}
	// infoc
	var err error
	if s.infocV2Log, err = infoc.New(nil); err != nil {
		panic(err)
	}
	if s.gaiaEngine, err = gaia.New(nil); err != nil {
		panic(err)
	}
	s.loadUnicomIP()
	s.loadUnicomPacks()
	s.initCron()
	s.cron.Start()
	// nolint:biligowordcheck
	go s.userbindConsumer()
	return
}

func (s *Service) initCron() {
	checkErr(s.cron.AddFunc("@every 2m", s.loadUnicomIP))    // 间隔2分钟
	checkErr(s.cron.AddFunc("@every 2m", s.loadUnicomPacks)) // 间隔2分钟
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *Service) Close() {
	s.cron.Stop()
	s.cache.Close()
}
