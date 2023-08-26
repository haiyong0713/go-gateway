package feed

import (
	"context"
	"math"
	"strconv"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/sync/errgroup"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-job/job/conf"
	feeddao "go-gateway/app/app-svr/app-job/job/dao/feed"
	"go-gateway/app/app-svr/app-job/job/model/feed"
	"go-gateway/app/app-svr/archive/service/api"

	"github.com/robfig/cron"
)

// Service is show service.
type Service struct {
	c       *conf.Config
	dao     *feeddao.Dao
	cardSub *databus.Databus
	// waiter
	waiter sync.WaitGroup
	cron   *cron.Cron
}

// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     feeddao.New(c),
		cardSub: databus.New(c.CardDatabus),
		cron:    cron.New(),
	}
	s.loadRcmdCache()
	// 间隔二分钟
	if err := s.cron.AddFunc("@every 2m", s.loadRcmdCache); err != nil {
		panic(err)
	}
	s.cron.Start()
	s.waiter.Add(1)
	go s.cardConsumer()
	return
}

func (s *Service) loadRcmdCache() {
	var (
		c    = context.Background()
		now  = time.Now()
		aids []int64
		is   []*ai.Item
		err  error
	)
	if aids, err = s.dao.Hots(c); err != nil {
		log.Error("%+v", err)
		return
	}
	is = s.fromAids(c, aids, now)
	if len(is) == 0 {
		return
	}
	if err = s.dao.AddRcmdCache(c, is); err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("loadRcmdCache success")
}

func (s *Service) fromAids(c context.Context, aids []int64, now time.Time) (is []*ai.Item) {
	if len(aids) == 0 {
		return
	}
	const (
		_count   = 50
		_aidsMax = 100
	)
	var (
		shard      int
		err        error
		start, end int
		mutex      sync.Mutex
	)
	am := map[int64]*api.Arc{}
	// 获取总共需要循环几次
	num := int(math.Ceil(float64(len(aids)) / float64(_aidsMax)))
	g, ctx := errgroup.WithContext(c)
	// aids超过100个特殊处理
	for i := 0; i < num; i++ {
		start = i * _aidsMax
		end = start + _aidsMax
		var tmpAids []int64
		if len(aids) >= end {
			tmpAids = aids[start:end]
		} else if len(aids) < end {
			tmpAids = aids[start:]
		} else if len(aids) < start {
			break
		}
		g.Go(func() (err error) {
			var tmpam map[int64]*api.Arc
			if tmpam, err = s.dao.Archives(ctx, tmpAids, ""); err != nil || len(tmpam) == 0 {
				log.Error("%+v", err)
				return
			}
			mutex.Lock()
			for k, v := range tmpam {
				am[k] = v
			}
			mutex.Unlock()
			return
		})
	}
	if len(aids) < _count {
		shard = 1
	} else {
		shard = len(aids) / _count
		if len(aids)%(shard*_count) != 0 {
			shard++
		}
	}
	aidss := make([][]int64, shard)
	for i, aid := range aids {
		aidss[i%shard] = append(aidss[i%shard], aid)
	}
	tagms := make([]map[string][]*feed.Tag, len(aidss))
	for i, aids := range aidss {
		if len(aids) == 0 {
			continue
		}
		idx := i
		tmpAid := aids
		g.Go(func() (err error) {
			if tagms[idx], err = s.dao.Tags(ctx, tmpAid, now); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if len(am) == 0 {
		return
	}
	tagm := make(map[string][]*feed.Tag, len(aids))
	for _, tm := range tagms {
		for aid, tag := range tm {
			tagm[aid] = tag
		}
	}
	is = make([]*ai.Item, 0, len(am))
	for _, aid := range aids {
		a, ok := am[aid]
		if !ok {
			continue
		}
		i := &ai.Item{ID: a.Aid, Archive: a}
		if ts, ok := tagm[strconv.FormatInt(aid, 10)]; ok {
			if len(ts) != 0 {
				i.Tid = ts[0].ID
				i.Tag = ts[0].AITag()
			}
		}
		is = append(is, i)
	}
	return
}

// Close is.
func (s *Service) Close() {
	s.cron.Stop()
	s.cardSub.Close()
	s.waiter.Wait()
	log.Info("app-job closed")
}
