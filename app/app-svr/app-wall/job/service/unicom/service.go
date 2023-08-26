package unicom

import (
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-wall/job/conf"
	accDao "go-gateway/app/app-svr/app-wall/job/dao/account"
	"go-gateway/app/app-svr/app-wall/job/dao/comic"
	comicDao "go-gateway/app/app-svr/app-wall/job/dao/comic"
	liveDao "go-gateway/app/app-svr/app-wall/job/dao/live"
	lockDao "go-gateway/app/app-svr/app-wall/job/dao/redis_lock"
	seqDao "go-gateway/app/app-svr/app-wall/job/dao/seq"
	shopDao "go-gateway/app/app-svr/app-wall/job/dao/shopping"
	unicomDao "go-gateway/app/app-svr/app-wall/job/dao/unicom"

	"github.com/robfig/cron"
)

type Service struct {
	c            *conf.Config
	dao          *unicomDao.Dao
	seqdao       *seqDao.Dao
	clickRailGun *railgun.Railgun
	comicRailGun *railgun.Railgun
	packRailGun  *railgun.Railgun
	canalRailGun *railgun.Railgun
	packPub      *databus.Databus
	// prom
	pHit  *prom.Prom
	pMiss *prom.Prom
	// redis lock
	lockdao         *lockDao.Dao
	lockExpire      int32
	monthLockExpire int32
	cron            *cron.Cron
	// pack dao
	live  *liveDao.Dao
	accd  *accDao.Dao
	shop  *shopDao.Dao
	comic *comicDao.Dao
	// cache
	cache *fanout.Fanout
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		dao:    unicomDao.New(c),
		seqdao: seqDao.New(c),
		// prom
		pHit:  prom.CacheHit,
		pMiss: prom.CacheMiss,
		// redis lock
		lockdao:         lockDao.New(c),
		lockExpire:      int32(time.Duration(c.Redis.Wall.LockExpire) / time.Second),
		monthLockExpire: int32(time.Duration(c.Redis.Wall.MonthLockExpire) / time.Second),
		packPub:         databus.New(c.PackPub),
		cron:            cron.New(),
		// pack dao
		live:  liveDao.New(c),
		accd:  accDao.New(c),
		shop:  shopDao.New(c),
		comic: comic.New(c),
		// cache
		cache: fanout.New("cache", fanout.Buffer(10240)),
	}
	s.initClickRailGun(c.ReportDatabus, c.ReportRailgun)
	s.initComicRailGun(c.ComicDatabus, c.ComicRailgun)
	s.initCanalRailGun(c.CanalSub, c.CanalRailgun)
	s.initPackRailGun(c.PackSub, c.PackRailgun)

	if s.c.Monthly {
		// nolint:biligowordcheck
		go s.upBindAll()
	}
	s.loadUnicomIPOrder()
	s.initCron()
	return
}

// Close Service
func (s *Service) Close() {
	s.clickRailGun.Close()
	s.comicRailGun.Close()
	s.packRailGun.Close()
	s.canalRailGun.Close()
	s.cron.Stop()
	log.Info("app-wall-job unicom flow closed.")
}

func (s *Service) initCron() {
	checkErr(s.cron.AddFunc("@every 3m", s.loadUnicomIPOrder)) // 间隔3分钟
	checkErr(s.cron.AddFunc("0 0 2 1 * ?", s.upBindAll))       // 每个月的第一天午夜两点运行一次给用户增加福利点
	s.cron.Start()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
