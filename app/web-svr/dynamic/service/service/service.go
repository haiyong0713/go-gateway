package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/sync/pipeline/fanout"
	arcclient "go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/web-svr/dynamic/service/conf"
	"go-gateway/app/web-svr/dynamic/service/dao"
	"go-gateway/app/web-svr/dynamic/service/model"

	"github.com/pkg/errors"
)

const (
	// CardsByAids2接口限制最多50个aid
	_cardsLimit = 50
	// content-flow-control.service gRPC infos limit
	_cfcAttributeLimit = 30
)

// Service service.
type Service struct {
	dao *dao.Dao
	c   *conf.Config
	// new dynamic arcs.
	regionTotal        map[int32]int
	regionArcs         map[int32][]int64
	regionTagArcs      map[string][]int64
	regionBusinessArcs map[string][]int64
	regionFilterArcs   map[int32][]int64
	// live
	live int
	// grpc
	arcClient    arcclient.ArchiveClient
	seasonClient seasongrpc.SeasonClient
	// cache
	cache    *fanout.Fanout
	waiter   *sync.WaitGroup
	closeSub bool
	// databus
	videoupSub *databus.Databus
	closeRetry bool
	// init lock
	arcInit    bool
	typesMap   map[int32]*arcclient.Tp
	greyRidMap map[int64]struct{}
	cron       *cron.Cron
	// arcFlowControl databus
	arcFlowControlRailGun *railgun.Railgun
}

// New service new.
func New(c *conf.Config) *Service {
	s := &Service{
		dao: dao.New(c),
		c:   c,
		// new dynamic arcs
		regionTotal:   make(map[int32]int),
		regionArcs:    make(map[int32][]int64),
		regionTagArcs: make(map[string][]int64),
		cache:         fanout.New("dynamic cache", fanout.Buffer(1024)),
		waiter:        new(sync.WaitGroup),
		typesMap:      make(map[int32]*arcclient.Tp),
		greyRidMap:    make(map[int64]struct{}),
		cron:          cron.New(),
	}
	var err error
	if s.arcClient, err = arcclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.seasonClient, err = seasongrpc.NewClient(c.SeasonClient); err != nil {
		panic(err)
	}
	if env.DeployEnv == env.DeployEnvProd {
		s.videoupSub = databus.New(c.ArchiveNotifySub)
	}
	s.initArchiveFlowControlRailGun(&railgun.DatabusV1Config{Config: s.c.ArchiveFlowControlSub}, s.c.ArcFlowControlCfg)
	s.regionTypeID()
	s.initCron()
	// nolint:biligowordcheck
	go s.regionTypeIDproc()
	s.waiter.Add(1)
	// archive region
	// nolint:biligowordcheck
	go s.regArchive()
	// nolint:biligowordcheck
	go s.regionproc()
	// nolint:biligowordcheck
	go s.tagproc()
	// WeChat warning
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.retryReg()
	return s
}

func (s *Service) initCron() {
	s.loadBusinessRegion()
	if err := s.cron.AddFunc(s.c.Cron.LoadBusinessRegion, s.loadBusinessRegion); err != nil {
		panic(err)
	}
	s.cron.Start()
}

func regionTagKey(rid int32, tagID int64) string {
	return fmt.Sprintf("%d_%d", rid, tagID)
}

// regionproc is a routine for pull region dynamic into cache.
// nolint:gocognit
func (s *Service) regionproc() {
	var (
		c           = context.Background()
		fRids       []int32
		cacheRegion map[int32][]int64
		err         error
		cRids       []int32
	)
	if fRids, cRids, err = s.typeIDs(c); err != nil || len(cRids) == 0 || len(fRids) == 0 {
		panic(err)
	}
	// get all origin id
	for {
		// load hot tags from tag api.
		regionTotal := make(map[int32]int)
		regionArcs := make(map[int32][]int64)
		// get region cache
		cacheRegion = s.dao.RegionCache(c)
		// init dynamic arcs from bigdata.
		for _, rid := range cRids {
			// init region dynamic arcs.
			if aids, total, err := s.dao.RegionArcs(c, rid, ""); err != nil || len(aids) < conf.Conf.Rule.MinRegionCount {
				regionTotal[rid] = s.regionTotal[rid]
				if len(s.regionArcs[rid]) >= conf.Conf.Rule.MinRegionCount {
					regionArcs[rid] = s.regionArcs[rid]
				} else if cacheRegion != nil && len(cacheRegion[rid]) >= conf.Conf.Rule.MinRegionCount {
					regionArcs[rid] = cacheRegion[rid]
				} else {
					dao.PromError("热门动态数据错误", "dynamic data error rid(%d)  bigdata(%d)  memory(%d)", rid, len(aids), len(s.regionArcs[rid]))
					if len(aids) > 0 {
						regionArcs[rid] = aids
					} else {
						regionArcs[rid] = s.regionArcs[rid]
					}
				}
			} else {
				regionTotal[rid] = total
				regionArcs[rid] = aids
			}
		}
		for _, rid := range fRids {
			if aids, total, err := s.dao.RegionArcs(c, rid, ""); err != nil || len(aids) < conf.Conf.Rule.MinRegionCount {
				regionTotal[rid] = s.regionTotal[rid]
				if len(s.regionArcs[rid]) >= conf.Conf.Rule.MinRegionCount {
					regionArcs[rid] = s.regionArcs[rid]
				} else if cacheRegion != nil && len(cacheRegion[rid]) >= conf.Conf.Rule.MinRegionCount {
					regionArcs[rid] = cacheRegion[rid]
				} else {
					dao.PromError("分区动态数据错误", "dynamic data error rid(%d)  bigdata(%d)  memory(%d)", rid, len(aids), len(s.regionArcs[rid]))
					if len(aids) > 0 {
						regionArcs[rid] = aids
					} else {
						regionArcs[rid] = s.regionArcs[rid]
					}
				}
			} else {
				regionTotal[rid] = total
				regionArcs[rid] = aids
			}
		}
		if count, err := s.dao.Live(c); err != nil {
			log.Error("s.dao.Live() error(%v)", err)
		} else {
			s.live = count
		}
		regionFilterArcs := make(map[int32][]int64)
		for _, rid := range s.c.FilterRids {
			if aids, ok := regionArcs[rid]; ok {
				filterAids, err := s.filterAids(c, aids)
				if err != nil {
					continue
				}
				if len(filterAids) > 0 {
					regionFilterArcs[rid] = filterAids
				}
			}
		}
		s.regionTotal = regionTotal
		s.regionArcs = regionArcs
		s.regionFilterArcs = regionFilterArcs
		if regionNeedCache(s.regionArcs) {
			s.cache.Do(c, func(ctx context.Context) {
				_ = s.dao.SetRegionCache(ctx, s.regionArcs)
			})
		}
		time.Sleep(time.Duration(s.c.Rule.TickRegion))
	}
}

func (s *Service) loadBusinessRegion() {
	ctx := context.Background()
	cacheRegion := s.dao.RegionBusinessCache(ctx)
	regionArcs := map[string][]int64{}
	for key, val := range s.c.LandingPage {
		if len(val.TagID) == 0 {
			continue
		}
		for _, rid := range val.Rid {
			key := fmt.Sprintf("%s_%d", key, rid)
			aids, _, err := s.dao.LpRegionArcs(ctx, int32(rid), val.TagID)
			if err != nil {
				log.Error("%+v", err)
			}
			if len(aids) < conf.Conf.Rule.MinRegionCount {
				if len(s.regionBusinessArcs[key]) >= conf.Conf.Rule.MinRegionCount {
					regionArcs[key] = s.regionBusinessArcs[key]
					continue
				}
				if len(cacheRegion[key]) >= conf.Conf.Rule.MinRegionCount {
					regionArcs[key] = cacheRegion[key]
					continue
				}
				dao.PromError("分区动态数据错误", "dynamic data error rid(%d)  bigdata(%d)  memory(%d)", rid, len(aids), len(s.regionBusinessArcs[key]))
				if len(aids) > 0 {
					regionArcs[key] = aids
					continue
				}
				regionArcs[key] = s.regionBusinessArcs[key]
				continue
			}
			regionArcs[key] = aids
		}
	}
	s.regionBusinessArcs = regionArcs
	if regionBusinessNeedCache(s.regionBusinessArcs) {
		s.cache.Do(ctx, func(ctx context.Context) {
			_ = s.dao.SetRegionBusinessCache(ctx, s.regionBusinessArcs)
		})
	}
}

// tagproc is a routine for pull tag dynamic into cache.
func (s *Service) tagproc() {
	var (
		err          error
		tagNeedCache bool
		res          map[int32][]int64
		c            = context.Background()
		cacheTag     map[string][]int64
		hotRidTids   = make(map[int32][]int64)
	)
	for {
		// get tag cache
		cacheTag = s.dao.TagCache(c)
		// load hot tags from tag api.
		regionTagArcs := make(map[string][]int64)
		// get hot tag
		if res, err = s.dao.Hot(c); err != nil {
			log.Error("dao.Hot() error(%v)", err)
			time.Sleep(time.Second)
			continue
		}
		if len(res) > 0 {
			hotRidTids = res
		}
		// init dynamic arcs from bigdata.
		for rid, tids := range hotRidTids {
			// init region tag dynamic arcs.
			for _, tid := range tids {
				k := regionTagKey(rid, tid)
				if aids, err := s.dao.RegionTagArcs(c, rid, tid, ""); err != nil || len(aids) == 0 {
					if len(s.regionTagArcs[k]) == 0 {
						if cacheTag != nil && len(cacheTag[k]) > 0 {
							regionTagArcs[k] = cacheTag[k]
						}
						tagNeedCache = false || tagNeedCache
					} else {
						regionTagArcs[k] = s.regionTagArcs[k]
						tagNeedCache = true
					}
				} else {
					regionTagArcs[k] = aids
					tagNeedCache = true
				}
			}
		}
		s.regionTagArcs = regionTagArcs
		if tagNeedCache {
			s.cache.Do(c, func(ctx context.Context) {
				_ = s.dao.SetTagCache(ctx, s.regionTagArcs)
			})
		}
		time.Sleep(time.Duration(s.c.Rule.TickTag))
	}
}

func (s *Service) typeIDs(ctx context.Context) (fRids, cRids []int32, err error) {
	var typesRes *arcclient.TypesReply
	for i := 0; i < _retry; i++ {
		if typesRes, err = s.arcClient.Types(ctx, &arcclient.NoArgRequest{}); err == nil &&
			typesRes != nil && len(typesRes.Types) != 0 {
			for k, v := range typesRes.Types {
				if v.Pid == 0 { // is father region
					fRids = append(fRids, k)
				} else { // is child region
					cRids = append(cRids, k)
				}
			}
			return
		}
		log.Error("s.arcClient.Types error(%v)", err)
	}
	return
}

// secondary partition archive .
func (s *Service) regArchive() {
	defer s.waiter.Done()
	if env.DeployEnv != env.DeployEnvProd {
		return
	}
	for {
		var (
			ok  bool
			err error
			msg *databus.Message
			ctx = context.Background()
		)
		if msg, ok = <-s.videoupSub.Messages(); !ok {
			log.Error("s.videoupSub.messages closed")
			return
		}
		if s.closeSub {
			return
		}
		if err := msg.Commit(); err != nil {
			log.Error("%+v", err)
		}
		m := &model.ArcMsg{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", msg.Value, err)
			continue
		}
		log.Info("regArchive key(%s) value(%s) start", msg.Key, msg.Value)
		switch m.Action {
		case model.Insert: // insert
			if m.New.CanPlay() {
				s.addRegion(ctx, m.New)
			}
		case model.Update:
			if m.New == nil || m.Old == nil {
				log.Warn("Invalid update message: %+v", m)
				continue
			}
			if m.New.Typeid != m.Old.Typeid { // 分区变化
				s.delArc(ctx, m.Old.Typeid, m.Old)
				log.Info("[ArchiveDatabus] aid(%d) NewTypeID(%d) OldtypeID(%d) databus archive Typeid change !", m.New.Aid, m.New.Typeid, m.Old.Typeid)
			}
			s.regionCache(ctx, m.New)
		}
	}
}

// Ping check server ok
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close dao
func (s *Service) Close() {
	s.closeSub = true
	s.closeRetry = true
	s.waiter.Wait()
	// close redis and mc
	s.dao.Close()
	// close archive databus
	if s.videoupSub != nil {
		s.videoupSub.Close()
	}
}

// nolint:gomnd
func archivesLog(name string, aids []int64) {
	if aidLen := len(aids); aidLen >= 50 {
		log.Info("s.archives3 func(%s) len(%d), arg(%v)", name, aidLen, aids)
	}
}

func (s *Service) regionTypeID() {
	var (
		err      error
		typesRes *arcclient.TypesReply
	)
	if typesRes, err = s.arcClient.Types(context.Background(), &arcclient.NoArgRequest{}); err != nil {
		log.Error("[regionTypeID] s.arcClient.Types error(%v)", err)
		return
	}
	if typesRes == nil || len(typesRes.Types) == 0 {
		log.Error("[regionTypeID] s.arcClient.Types return nil")
		return
	}
	s.typesMap = typesRes.Types
}

func (s *Service) regionTypeIDproc() {
	for {
		time.Sleep(time.Second * 5)
		s.regionTypeID()
	}
}

func (s *Service) filterAids(ctx context.Context, aids []int64) ([]int64, error) {
	var aidsLimit [][]int64
	for index := 0; index < len(aids); index += _cardsLimit {
		length := index + _cardsLimit
		if length > len(aids) {
			length = len(aids)
		}
		aidsLimit = append(aidsLimit, aids[index:length])
	}
	replyMap := make(map[int64]*seasongrpc.CardInfoProto)
	for _, aids := range aidsLimit {
		reply, err := s.seasonClient.CardsByAids2(ctx, &seasongrpc.SeasonAidReq{Aid2S: aids})
		if err != nil {
			log.Error("season.CardsByAids2 err:%+v", err)
			return nil, err
		}
		for aid, info := range reply.GetCards() {
			replyMap[aid] = info
		}
	}
	var filterAids []int64
	seasonMap := make(map[int32]struct{})
	for _, aid := range aids {
		if card, ok := replyMap[aid]; ok {
			if _, ok := seasonMap[card.SeasonId]; !ok {
				seasonMap[card.SeasonId] = struct{}{}
				filterAids = append(filterAids, aid)
			}
			continue
		}
		filterAids = append(filterAids, aid)
	}
	return filterAids, nil
}

func (s *Service) regionCache(ctx context.Context, arc *model.ArchiveSub) {
	info, err := s.dao.ContentFlowControlInfoV2(ctx, arc.Aid)
	if err != nil {
		log.Error("日志告警 ContentFlowControlInfoV2 aid:%d, error:%v", arc.Aid, err)
	}
	forbidden := model.ItemToArcForbidden(info)
	if !arc.CanPlay() || !forbidden.AllowShow() {
		s.delArc(ctx, arc.Typeid, arc)
		log.Info("[regionCache] aid(%d) databus archive delete !", arc.Aid)
		return
	}
	if !arc.IsOriginArc() {
		s.delArc(ctx, arc.Typeid, arc)
		log.Info("[regionCache] aid(%d) databus archive Origin change !", arc.Aid)
	}
	s.addRegion(ctx, arc)
}

func (s *Service) batchCfcInfos(ctx context.Context, aids []int64) (map[int64]*cfcgrpc.FlowCtlInfoV2Reply, error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	cfcInfos := make(map[int64]*cfcgrpc.FlowCtlInfoV2Reply, aidsLen)
	group := errgroup.WithContext(ctx)
	for i := 0; i < aidsLen; i += _cfcAttributeLimit {
		var partAids []int64
		l := i + _cfcAttributeLimit
		if l > aidsLen {
			l = aidsLen
		}
		partAids = aids[i:l]
		group.Go(func(ctx context.Context) error {
			var reply *cfcgrpc.FlowCtlInfosV2Reply
			if err := retry(func() (err error) {
				reply, err = s.dao.ContentFlowControlInfosV2(ctx, partAids)
				return err
			}); err != nil {
				log.Error("日志告警 ContentFlowControlInfosV2 error:%v", err)
				return errors.Wrapf(err, "ContentFlowControlInfosV2 partAids:%v", partAids)
			}
			if reply == nil {
				return nil
			}
			mutex.Lock()
			for k, v := range reply.ItemsMap {
				cfcInfos[k] = v
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("日志告警 s.batchCfcInfos error:%+v, arg:%+v", err, aids)
		return nil, err
	}
	return cfcInfos, nil
}
