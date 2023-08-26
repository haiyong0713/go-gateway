package service

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/conf"
	arcdao "go-gateway/app/app-svr/archive/service/dao/archive"
	hisDao "go-gateway/app/app-svr/archive/service/dao/history"
	ipDao "go-gateway/app/app-svr/archive/service/dao/ipdisplay"
	locationDao "go-gateway/app/app-svr/archive/service/dao/location"
	playurlDao "go-gateway/app/app-svr/archive/service/dao/playurl"
	puDao "go-gateway/app/app-svr/archive/service/dao/user"
	vasDao "go-gateway/app/app-svr/archive/service/dao/vas"
	"go-gateway/app/app-svr/archive/service/model/archive"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"

	"github.com/robfig/cron"
)

type CdnScore struct {
	WwanScoreIps []string
	WifiScoreIps []string
	LastUpTime   int64
}

// Service is service.
type Service struct {
	c *conf.Config
	// dao
	arc          *arcdao.Dao
	playurldao   *playurlDao.Dao
	hisdao       *hisDao.Dao
	locationDao  *locationDao.Dao
	vasDao       *vasDao.Dao
	puDao        *puDao.Dao
	ipDisplayDao *ipDao.Dao
	// types
	allTypes  map[int16]*archive.ArcType
	ridToReid map[int16]int16
	// cache chan
	cacheCh chan func()
	// steins gate
	authorisedCallers map[string]struct{}
	steinsGuidePages  []*api.Page
	locGRPC           locgrpc.LocationClient
	cdnScores         map[string]CdnScore
	cdnScoresMu       sync.RWMutex
	cache             *fanout.Fanout
	cdnRedis          *redis.Pool
	// cron
	cron              *cron.Cron
	recentPremiereArc map[int64][]int64
	upMidFixedIp      map[int64]string
	infoProm          *prom.Prom
	buvidFixedIp      map[int64]string
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		// dao
		arc:          arcdao.New(c),
		playurldao:   playurlDao.New(c),
		hisdao:       hisDao.New(c),
		locationDao:  locationDao.New(c),
		vasDao:       vasDao.New(c),
		puDao:        puDao.New(c),
		ipDisplayDao: ipDao.New(c),
		// types
		allTypes:  make(map[int16]*archive.ArcType),
		ridToReid: make(map[int16]int16),
		// cache chan
		cacheCh:   make(chan func(), 1024),
		cdnScores: make(map[string]CdnScore),
		cache:     fanout.New("cdn_cache"),
		cdnRedis:  redis.NewPool(c.Redis.CdnScore),
		cron:      cron.New(),
		infoProm:  prom.BusinessInfoCount,
	}
	var err error
	if s.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(fmt.Sprintf("locgrpc.NewClient err(%+v)", err))
	}
	s.loadTypes()
	s.loadSteinsCallers()
	s.loadSteinsPage()
	// nolint:biligowordcheck
	go s.loadproc()
	for i := 0; i < runtime.NumCPU(); i++ {
		// nolint:biligowordcheck
		go s.cacheproc()
	}
	s.initCron()
	s.cron.Start()
	return
}

// AllTypes return all types
func (s *Service) AllTypes(c context.Context) (types map[int16]*archive.ArcType) {
	types = s.allTypes
	return
}

// Close resource.
func (s *Service) Close() {
	s.arc.Close()
	_ = s.cache.Close()
}

func (s *Service) loadSteinsCallers() {
	tmp := make(map[string]struct{})
	for _, v := range s.c.Custom.SteinsCallers {
		tmp[v] = struct{}{}
	}
	s.authorisedCallers = tmp
}

func (s *Service) loadTypes() {
	var (
		ridToReid = make(map[int16]int16)
		types     map[int16]*archive.ArcType
		err       error
	)
	if types, err = s.arc.RawTypes(context.TODO()); err != nil {
		log.Error("s.arc.Types error(%v)", err)
		return
	}
	for _, t := range types {
		if t.Pid != 0 {
			ridToReid[t.ID] = t.Pid
		}
	}
	s.allTypes = types
	s.ridToReid = ridToReid
}

// load SteinsGate Guide Aid's Page Info
func (s *Service) loadSteinsPage() {
	if said := s.c.Custom.SteinsGuideAid; said != 0 {
		if ps, err := s.arc.Videos3(context.Background(), said); err == nil && len(ps) > 0 {
			s.steinsGuidePages = ps[:1] // pick only the first Cid
		}
	} else {
		log.Error("SteinsGuideAid is Nil!")
	}
}

func (s *Service) loadproc() {
	for {
		time.Sleep(time.Duration(s.c.Tick))
		s.loadTypes()
		s.loadSteinsCallers()
		s.loadSteinsPage()
	}
}

func (s *Service) addCache(f func()) {
	select {
	case s.cacheCh <- f:
	default:
		log.Warn("s.cacheCh is full")
	}
}

func (s *Service) cacheproc() {
	for {
		f, ok := <-s.cacheCh
		if !ok {
			return
		}
		f()
	}
}

func (s *Service) initCron() {
	s.loadRecentPremiereArc()
	s.loadFixedLocation()
	s.loadBuvidFixedLocation()

	var err error
	if err = s.cron.AddFunc(s.c.Cron.LoadRecentPremiereArc, s.loadRecentPremiereArc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadFixedLocation, s.loadFixedLocation); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadBuvidFixedLocation, s.loadBuvidFixedLocation); err != nil {
		panic(err)
	}
}

func (s *Service) loadFixedLocation() {
	res, err := s.puDao.UserFixedLocations(context.Background())
	if err != nil {
		log.Error("loadFixedLocation() fail: %+v", err)
		return
	}
	s.upMidFixedIp = res.FixedLocations
	log.Info("loadFixedLocation s.upMidFixedIp: %+v", s.upMidFixedIp)
}

func (s *Service) loadBuvidFixedLocation() {
	res, err := s.ipDisplayDao.IpDisplay(context.Background())
	if err != nil {
		log.Error("loadBuvidFixedLocation() fail: %+v", err)
		return
	}
	s.buvidFixedIp = res.Result
	log.Info("loadBuvidFixedLocation s.buvidFixedIp: %+v", s.buvidFixedIp)
}

func (s *Service) loadRecentPremiereArc() {
	//获取最近1天的首映稿件
	startTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
	endTime := time.Now().Format("2006-01-02 15:04:05")
	premiereArcs, err := s.arc.RawArchiveExpandByPremiereTime(context.TODO(), startTime, endTime)
	if err != nil {
		log.Error("日志告警 loadRecentPremiereArc error(%v)", err)
		return
	}
	midToAids := make(map[int64][]int64)
	if len(premiereArcs) == 0 {
		s.recentPremiereArc = midToAids
		return
	}

	//获取mid对应的最先首映中的稿件id
	for _, pa := range premiereArcs {
		midToAids[pa.Mid] = append(midToAids[pa.Mid], pa.Aid)
	}

	s.recentPremiereArc = midToAids
	log.Info("loadRecentPremiereArc midToAid(%+v)", midToAids)
}
