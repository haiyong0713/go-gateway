package dao

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/job/internal/model"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewRedis, NewDB)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	WebTop(ctx context.Context) ([]int64, error)
	RankIndex(ctx context.Context, day int64) ([]int64, error)
	RankRecommend(ctx context.Context, rid int64) ([]int64, error)
	LpRankRecommend(ctx context.Context, business string) ([]int64, error)
	RankRegion(ctx context.Context, rid, day, original int64) ([]*model.RankAid, error)
	RankTag(ctx context.Context, rid, tagID int64) ([]*model.RankAid, error)
	RankList(ctx context.Context, typ model.RankListType, rid int64) (*model.RankList, error)
	RankListOld(ctx context.Context, rid int64) (*model.RankList, error)
	OnlineAids(ctx context.Context, num int64) ([]*model.OnlineAid, error)
	TagHots(ctx context.Context, rid int64) ([]int64, error)
	AddCacheWebTop(ctx context.Context, aids []int64) error
	AddCacheRankIndex(ctx context.Context, day int64, aids []int64) error
	AddCacheRankRecommend(ctx context.Context, rid int64, aids []int64) error
	AddCacheLpRankRecommend(ctx context.Context, business string, aids []int64) error
	AddCacheRankRegion(ctx context.Context, rid, day, original int64, list []*model.RankAid) error
	AddCacheRankTag(ctx context.Context, rid, tagID int64, list []*model.RankAid) error
	AddCacheRankList(ctx context.Context, typ model.RankListType, rid int64, data *model.RankList) error
	AddCacheOnlineAids(ctx context.Context, list []*model.OnlineAid) error
	AddCacheNewList(ctx context.Context, rid, typ int64, list []*model.BvArc, total int) error
	RegionList(ctx context.Context) ([]*model.Region, error)
	RegionConfig(ctx context.Context) (map[int64][]*model.RegionConfig, error)
	AddCacheRegionList(ctx context.Context, data map[string][]*model.Region) error
	PopularSeries(ctx context.Context) ([]*model.MgrSeriesData, error)
	AddCacheSeries(ctx context.Context, typ string, data []*model.MgrSeriesConfig) error
	AddCacheSeriesDetail(ctx context.Context, data map[int64][]*model.MgrSeriesList) error
	AddPopularRank(ctx context.Context, mid int64) (int64, error)
	AddPopularWatchTime(ctx context.Context, mid int64, stage int8, t time.Time) (int64, error)
	DelCachePopularWatchTime(ctx context.Context, mid int64, stage int64) error
	DelCachePopularRank(ctx context.Context, mid int64) error
}

// dao dao.
type dao struct {
	redis            *redis.Redis
	httpR            *bm.Client
	showDB           *sql.DB
	webTopURL        string
	rankIndexURL     string
	rankRcmdURL      string
	lpRankRcmdURL    string
	rankRegionURL    string
	rankTagURL       string
	rankListURL      string
	rankListOldURL   string
	onlineListURL    string
	tagHotsURL       string
	popularSeriesURL string
}

// New new a dao and return.
func New(r *redis.Redis, db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(r, db)
}

func newDao(r *redis.Redis, db *sql.DB) (d *dao, cf func(), err error) {
	var cfg struct {
		HTTPClient *bm.ClientConfig
		Host       struct {
			Data    string
			Api     string
			Manager string
		}
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		redis:            r,
		showDB:           db,
		httpR:            bm.NewClient(cfg.HTTPClient),
		webTopURL:        cfg.Host.Data + _webTopURI,
		rankIndexURL:     cfg.Host.Data + _rankIndexURI,
		rankRcmdURL:      cfg.Host.Data + _rankRcmdURI,
		lpRankRcmdURL:    cfg.Host.Data + _lpRankRcmdURI,
		rankRegionURL:    cfg.Host.Data + _rankRegionURI,
		rankTagURL:       cfg.Host.Data + _rankTagURI,
		rankListURL:      cfg.Host.Data + _rankListURI,
		rankListOldURL:   cfg.Host.Data + _rankListOldURI,
		onlineListURL:    cfg.Host.Api + _onlineListURI,
		tagHotsURL:       cfg.Host.Api + _tagHotsURI,
		popularSeriesURL: cfg.Host.Manager + _popularSeriesURL,
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.redis.Close()
}

// Ping ping the resource.
func (d *dao) Ping(_ context.Context) (err error) {
	return nil
}
