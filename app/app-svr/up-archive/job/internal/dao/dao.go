package dao

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
	"go-common/library/net/rpc/warden"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/up-archive/job/internal/model"
	"go-gateway/app/app-svr/up-archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	videoUpOpen "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	RawArcPassed(ctx context.Context, mid int64) ([]*model.UpArc, error)
	RawArcs(ctx context.Context, aids []int64) ([]*model.UpArc, error)
	RawArc(ctx context.Context, aid int64) (*model.UpArc, error)
	RawStaffAids(ctx context.Context, mid int64) ([]int64, error)
	RawStaffMids(ctx context.Context, aid int64) ([]int64, error)
	RawUpper(ctx context.Context, mid, limit int64) ([]int64, error)
	AddCacheArcPassed(ctx context.Context, mid int64, arcs []*model.UpArc, without api.Without) error
	AddCacheArcStoryPassed(ctx context.Context, mid int64, arcs []*model.UpArc) error
	AppendCacheArcPassed(ctx context.Context, mid int64, arcs []*model.UpArc, without api.Without) error
	AppendCacheArcStoryPassed(ctx context.Context, mid int64, arcs []*model.UpArc) error
	DelCacheArcPassed(ctx context.Context, mid int64, arcs []*model.UpArc, without api.Without) error
	DelCacheArcStoryPassed(ctx context.Context, mid int64, arcs []*model.UpArc) error
	BuildArcPassedLock(ctx context.Context, mid int64) (bool, error)
	DelCacheAllArcPassed(ctx context.Context, mid int64) error
	DelCacheArcNoSpace(ctx context.Context, mid int64) error
	ContentFlowControlInfo(ctx context.Context, aid int64) ([]*cfcgrpc.ForbiddenItem, error)
	ContentFlowControlInfos(ctx context.Context, aids []int64) (map[int64][]*cfcgrpc.ForbiddenItem, error)
	CacheArcPassedExists(ctx context.Context, mid int64, without api.Without) (bool, error)
}

// dao dao.
type dao struct {
	resultDB          *sql.DB
	tempDB            *sql.DB
	redis             *redis.Redis
	cache             *fanout.Fanout
	lockExpire        int32
	emptyCacheExpire  int32
	emptyCacheRand    int32
	cfcGRPC           cfcgrpc.FlowControlClient
	secret            string
	videoUpOpenClient videoUpOpen.VideoUpOpenClient
	infoProm          *prom.Prom
}

// New new a dao and return.
func New(r *redis.Redis, db *DB) (d Dao, cf func(), err error) {
	return newDao(r, db)
}

func newDao(r *redis.Redis, db *DB) (d *dao, cf func(), err error) {
	var cfg struct {
		LockExpire         xtime.Duration
		EmptyExpire        xtime.Duration
		EmptyRand          int32
		CfcGRPC            *warden.ClientConfig
		VideoUpOpenClient  *warden.ClientConfig
		ContentFlowControl *struct {
			Secret string
		}
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		resultDB:         db.resultDB,
		tempDB:           db.tempDB,
		redis:            r,
		cache:            fanout.New("cache"),
		lockExpire:       int32(time.Duration(cfg.LockExpire) / time.Second),
		emptyCacheExpire: int32(time.Duration(cfg.EmptyExpire) / time.Second),
		emptyCacheRand:   cfg.EmptyRand,
		secret:           cfg.ContentFlowControl.Secret,
		infoProm:         prom.BusinessInfoCount,
	}
	if d.cfcGRPC, err = cfcgrpc.NewClient(cfg.CfcGRPC); err != nil {
		return
	}
	if d.videoUpOpenClient, err = videoUpOpen.NewClient(cfg.VideoUpOpenClient); err != nil {
		panic(err)
	}
	cf = d.Close
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
