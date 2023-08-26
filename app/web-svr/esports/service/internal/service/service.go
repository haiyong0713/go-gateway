package service

import (
	"context"
	"go-gateway/app/web-svr/esports/service/internal/model"
	cache2 "k8s.io/apimachinery/pkg/util/cache"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/esports/service/component"
	"go-gateway/app/web-svr/esports/service/conf"
	"go-gateway/app/web-svr/esports/service/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
)

const (
	_defaultLRUCacheSize = 10
)

// Service service.
type Service struct {
	ac     *paladin.Map
	dao    dao.Dao
	conf   *conf.Config
	fanout *fanout.Fanout

	// 实例级内存缓存
	seasonInfoCacheMap     map[int64]*model.SeasonModel
	activeTeamsCacheMap    *cache2.LRUExpireCache
	gamesCacheMap          map[int64]*model.GameModel
	matchesCacheMap        *cache2.LRUExpireCache
	seriesCacheMap         *cache2.LRUExpireCache
	activeContestsCacheMap *cache2.LRUExpireCache
	activeSeasonTeams      map[int64][]*model.SeasonTeamModel
}

// New new a service and return.
func New(conf *conf.Config) (s *Service) {

	s = &Service{
		ac:                     &paladin.TOML{},
		dao:                    dao.New(conf),
		conf:                   conf,
		fanout:                 fanout.New("service_fanout", fanout.Worker(10), fanout.Buffer(10240)),
		seasonInfoCacheMap:     make(map[int64]*model.SeasonModel),
		activeTeamsCacheMap:    s.lruCacheInit(conf.Rule.LRUCacheMaxTeamSize),
		gamesCacheMap:          make(map[int64]*model.GameModel),
		matchesCacheMap:        s.lruCacheInit(conf.Rule.LRUCacheMaxMatchSize),
		seriesCacheMap:         s.lruCacheInit(conf.Rule.LRUCacheMaxSeriesSize),
		activeContestsCacheMap: s.lruCacheInit(conf.Rule.LRUCacheMaxContestSize),
		activeSeasonTeams:      make(map[int64][]*model.SeasonTeamModel),
	}
	s.cacheInit(ctx4Worker)
	s.goroutineRegisterInit(ctx4Worker)
	return
}

func (s *Service) lruCacheInit(size int) *cache2.LRUExpireCache {
	if size == 0 {
		size = _defaultLRUCacheSize
	}
	return cache2.NewLRUExpireCache(size)
}

func (s *Service) cacheInit(ctx context.Context) {
	s.storeSeasonCache(ctx)
	s.storeActiveTeams(ctx)
	s.storeSeasonContestsCache(ctx)
	s.storeActiveSeries(ctx)
	_, _ = s.RefreshGameCache(ctx, nil)
	s.storeGameCache(ctx)
}

func (s *Service) goroutineRegisterInit(ctx context.Context) {
	s.goroutineRegister(ctx, s.activeSeasonTicker)
	s.goroutineRegister(ctx, s.seasonContestsTicker)
	s.goroutineRegister(ctx, s.activeTeamsTicker)
	s.goroutineRegister(ctx, s.activeSeriesTicker)
	s.goroutineRegister(ctx, s.activeGamesTicker)
}

func (s *Service) activeSeasonTicker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeSeasonCache(ctx)
		}
	}
}

func (s *Service) activeTeamsTicker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeActiveTeams(ctx)
		}
	}
}

func (s *Service) activeGamesTicker(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeGameCache(ctx)
		}
	}
}

func (s *Service) activeSeriesTicker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeActiveSeries(ctx)
		}
	}
}

func (s *Service) seasonContestsTicker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeSeasonContestsCache(ctx)
		}
	}
}

func (s *Service) goroutineRegister(ctx context.Context, f func(ctx2 context.Context)) {
	log.Infoc(ctx, "[Fanout][Register][Begin]")
	if err := s.fanout.Do(ctx, func(ctx context.Context) {
		f(ctx4Worker)
	}); err != nil {
		panic(err)
	}
	log.Infoc(ctx, "[Fanout][Register][Success]")
}

var (
	seasonContestAllComponent0Map map[int64][]*model.ContestModel
	seasonContestAllComponent1Map map[int64][]*model.ContestModel
	seasonContestAllComponent2Map map[int64][]*model.ContestModel
	seasonContestAllComponent3Map map[int64][]*model.ContestModel
	seasonContestAllComponent4Map map[int64][]*model.ContestModel
	seasonContestAllComponent5Map map[int64][]*model.ContestModel
	seasonContestAllComponent6Map map[int64][]*model.ContestModel
	seasonContestAllComponent7Map map[int64][]*model.ContestModel
	seasonContestAllComponent8Map map[int64][]*model.ContestModel
	seasonContestAllComponent9Map map[int64][]*model.ContestModel

	ctx4Worker           context.Context
	ctxCancelFunc4Worker context.CancelFunc
)

func init() {

	ctx4Worker, ctxCancelFunc4Worker = context.WithCancel(context.Background())

	seasonContestAllComponent0Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent1Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent2Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent3Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent4Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent5Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent6Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent7Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent8Map = make(map[int64][]*model.ContestModel)
	seasonContestAllComponent9Map = make(map[int64][]*model.ContestModel)

}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	ctxCancelFunc4Worker()
	s.dao.Close()
	component.Close()
}
