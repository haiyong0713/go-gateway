package dao

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/elastic"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/model"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewRedis, NewElastic)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	CacheArcPassed(ctx context.Context, mid, start, end int64, isAsc bool, without api.Without) ([]int64, error)
	CacheArcPassedStory(ctx context.Context, mid, start, end int64, isAsc bool) ([]int64, error)
	CacheArcPassedTotal(ctx context.Context, mid int64, without api.Without) (int64, error)
	CacheArcsPassed(ctx context.Context, mids []int64, start, end int64, isAsc bool, without api.Without) (map[int64][]int64, error)
	CacheArcsPassedTotal(ctx context.Context, mids []int64, without api.Without) (map[int64]int64, error)
	ExpireEmptyArcPassed(ctx context.Context, mid int64, without api.Without) error
	CacheArcPassedStoryTotal(ctx context.Context, mid int64) (int64, error)
	ExpireEmptyArcPassedStory(ctx context.Context, mid int64) error
	CacheArcPassedExists(ctx context.Context, mid int64, without api.Without) (bool, error)
	CacheArcPassedStoryExists(ctx context.Context, mid int64) (bool, error)
	SendBuildCacheMsg(ctx context.Context, mid, nowTs int64) error
	CacheArcPassedCursor(ctx context.Context, mid, score, ps int64, isAsc, containScore bool, without api.Without) ([]*api.ArcPassed, error)
	CacheArcPassedStoryAidRank(ctx context.Context, mid, aid int64, isAsc bool) (int64, error)
	CacheArcPassedScoreRank(ctx context.Context, mid, aid int64, isAsc bool, without api.Without) (score, rank int64, err error)
	CacheUpsPassed(ctx context.Context, mids []int64, start, end int64, isAsc bool, without api.Without) (map[int64][]*api.AidPubTime, error)
	CacheArcPassedExist(ctx context.Context, mid, aid int64, without api.Without) (bool, error)
	ArcSearch(ctx context.Context, mid int64, tid int64, keyword string, kwFields []string, highlight bool, pn int, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcPassedSearchReply, error)
	ArcSearchTag(ctx context.Context, mid int64, keyword string, kwFields []string, without []api.Without) (map[int64]int64, error)
	ArcSearchCursor(ctx context.Context, mid, score int64, containScore bool, ps int, without []api.Without, sort api.Sort) (*model.ArcPassedSearchReply, error)
	ArcSearchCursorAid(ctx context.Context, mid, score int64, equalScore bool, aid int64, tid int64, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcCursorAidSearchReply, error)
	ArcSearchScore(ctx context.Context, mid, aid, tid int64, order api.SearchOrder, without []api.Without) (*model.ArcScoreResult, error)
	ArcsSearchSort(ctx context.Context, mids []int64, tid int64, ps int, order api.SearchOrder, sort api.Sort) (map[int64][]int64, error)
}

// dao dao.
type dao struct {
	ac                      *paladin.Map
	redis, dRedis           *redis.Redis
	elastic                 *elastic.Elastic
	cache                   *fanout.Fanout
	upArcPub                *databus.Databus
	emptyCacheExpire        int32
	emptyCacheRand          int32
	degradeCacheExpire      int32
	degradeEmptyCacheExpire int32
	degradeCacheRand        int32
	notAttrs                []int64
	notAttrV2s              []int64
	livePlaybackUpFrom      []int64
	noSpace                 int64
	upNoSpace               int64
}

// New new a dao and return.
func New(r *Redis, ela *elastic.Elastic) (d Dao, cf func(), err error) {
	return newDao(r, ela)
}

func newDao(r *Redis, ela *elastic.Elastic) (d *dao, cf func(), err error) {
	var (
		dc       paladin.Map
		upArcPub *databus.Config
		cfg      struct {
			EmptyExpire        xtime.Duration
			EmptyRand          int32
			DegradeExpire      xtime.Duration
			DegradeEmptyExpire xtime.Duration
			DegradeRand        int32
			Search             struct {
				NotAttrs           []int64
				NotAttrV2s         []int64
				LivePlaybackUpFrom []int64
				NoMealIDs          *struct {
					NoSpace   int64
					UpNoSpace int64
				}
			}
		}
	)
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		ac:                      &paladin.TOML{},
		redis:                   r.r,
		dRedis:                  r.dr,
		elastic:                 ela,
		cache:                   fanout.New("cache"),
		emptyCacheExpire:        int32(time.Duration(cfg.EmptyExpire) / time.Second),
		emptyCacheRand:          cfg.EmptyRand,
		degradeCacheExpire:      int32(time.Duration(cfg.DegradeExpire) / time.Second),
		degradeEmptyCacheExpire: int32(time.Duration(cfg.DegradeEmptyExpire) / time.Second),
		degradeCacheRand:        cfg.DegradeRand,
		livePlaybackUpFrom:      cfg.Search.LivePlaybackUpFrom,
	}
	cf = d.Close
	if err = paladin.Watch("application.toml", d.ac); err != nil {
		return
	}
	if err = paladin.Get("databus.toml").Unmarshal(&dc); err != nil {
		return
	}
	if err = dc.Get("UpArchivePub").UnmarshalTOML(&upArcPub); err != nil {
		return
	}
	for _, val := range cfg.Search.NotAttrs {
		d.notAttrs = append(d.notAttrs, val+1) // 比特位+1
	}
	for _, val := range cfg.Search.NotAttrV2s {
		d.notAttrV2s = append(d.notAttrV2s, val+1) // 比特位+1
	}
	if cfg.Search.NoMealIDs != nil {
		d.noSpace = cfg.Search.NoMealIDs.NoSpace
		d.upNoSpace = cfg.Search.NoMealIDs.UpNoSpace
	}
	d.upArcPub = databus.New(upArcPub)
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
