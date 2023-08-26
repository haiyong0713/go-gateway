package bnj

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	arcclient "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bnj"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/dao/task"
	bnjmdl "go-gateway/app/web-svr/activity/interface/model/bnj"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"

	"github.com/robfig/cron"
)

// Service .
type Service struct {
	c            *conf.Config
	dao          *bnj.Dao
	likeDao      *like.Dao
	taskDao      *task.Dao
	cache        *fanout.Fanout
	previewArcs  map[int64]*arcclient.Arc
	previewTasks map[int64]*taskmdl.Task
	rareMaterial map[int64]*conf.Bnj20Material
	likeCount    int64
	timeReset    int64
	resetMid     int64
	timeFinish   int64
	resetCD      int32
	bnj20Mem     *bnjmdl.MemBnj20
	bnjPub       *databus.Databus
	bnjAwardPub  *databus.Databus
	cron         *cron.Cron
}

// New init bnj service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		dao:         bnj.New(c),
		likeDao:     like.New(c),
		taskDao:     task.New(c),
		cache:       fanout.New("cache", fanout.Worker(5), fanout.Buffer(1024)),
		bnj20Mem:    new(bnjmdl.MemBnj20),
		bnjPub:      databus.New(c.DataBus.BnjPub),
		bnjAwardPub: databus.New(c.DataBus.BnjAwardPub),
		cron:        cron.New(),
	}
	rand.Seed(time.Now().UnixNano())
	s.initBnj20Cfg()
	s.timeFinish = 1
	s.bnj20Mem.GameFinish = 1
	s.bnj20Mem.HotpotValue = s.c.Bnj2020.MaxValue
	s.loadBnj20Task()
	s.initBnj()
	s.createCron()
	go s.ASyncResetPlayerReserveCount(context.Background())

	return s
}

func (s *Service) createCron() {
	var err error
	if err = s.cron.AddFunc("@every 11m", s.loadBnjArc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 8m", s.loadBnjReserveTotal); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 2m", s.loadBnj20); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 1h", s.loadBnj20Task); err != nil {
		panic(err)
	}
	s.cron.Start()
}

// Close .
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}

func (s *Service) loadBnjArc() {
	now := time.Now().Unix()
	var aids []int64
	for _, v := range s.c.Bnj2019.Info {
		if v.Publish.Unix() < now {
			if v.Aid > 0 {
				aids = append(aids, v.Aid)
			}
		}
	}
	if len(aids) > 0 {
		if arcsReply, err := client.ArchiveClient.Arcs(context.Background(), &arcclient.ArcsRequest{Aids: aids}); err != nil {
			log.Error("bnjArcproc s.arcClient.Arcs(%v) error(%v)", aids, err)
		} else if len(arcsReply.GetArcs()) > 0 {
			tmp := make(map[int64]*arcclient.Arc, len(aids))
			for _, aid := range aids {
				if arc, ok := arcsReply.Arcs[aid]; ok && arc != nil {
					tmp[aid] = arc
				} else {
					log.Error("bnjArcproc aid(%d) data(%v)", aid, arc)
					continue
				}
			}
			s.previewArcs = tmp
		}
	} else {
		log.Error("bnjArcproc aids(%v) conf error", aids)
		return
	}
	log.Info("loadBnjArc() success")
}

func (s *Service) ASyncResetPlayerReserveCount(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, _ = s.likeDao.ResetPlayerReserveTotal(ctx, 15200)
		}
	}
}

func (s *Service) loadBnjReserveTotal() {
	totals, err := s.likeDao.ReservesTotal(context.Background(), []int64{s.c.Bnj2020.Sid})
	if err != nil {
		log.Error("bnjReserveTotalproc s.likeDao.ReservesTotal sid(%d) error(%v)", s.c.Bnj2020.Sid, err)
		return
	}
	total := totals[s.c.Bnj2020.Sid]
	if total > s.bnj20Mem.AppointCnt {
		atomic.StoreInt64(&s.bnj20Mem.AppointCnt, total)
		var level int64 = 1
		for _, v := range s.c.Bnj2020.Award {
			if total < v.Count {
				break
			}
			if v.Type == _awardTypeHotpot {
				level++
			}
		}
		if level > atomic.LoadInt64(&s.bnj20Mem.HotpotLevel) {
			atomic.StoreInt64(&s.bnj20Mem.HotpotLevel, level)
		}
	}
	log.Info("loadBnjReserveTotal() success")
}

func (s *Service) initBnj20Cfg() {
	// 初始level == 1
	s.bnj20Mem.HotpotLevel = 1
	tmp := make(map[int64]*conf.Bnj20Material, len(s.c.Bnj2020.RareList))
	for _, v := range s.c.Bnj2020.RareList {
		tmp[v.ID] = v
	}
	s.rareMaterial = tmp
}

func (s *Service) initBnj() {
	value, err := func() (amount int64, err error) {
		for i := 0; i < 3; i++ {
			if amount, err = s.dao.RawCurrencyAmount(context.Background(), s.c.Bnj2020.Sid); err == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		return
	}()
	if err != nil {
		panic(err)
	}
	if value >= s.c.Bnj2020.MaxValue {
		log.Warn("initBnj game finish value(%d)", value)
		atomic.StoreInt64(&s.bnj20Mem.GameFinish, 1)
	}
}
