package dao

import (
	"context"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"

	"go-common/library/conf/paladin.v2"

	"github.com/google/wire"
	"github.com/pkg/errors"

	"go-gateway/app/app-svr/fawkes/job/internal/model"
	"go-gateway/app/app-svr/fawkes/job/internal/model/mod"
	"go-gateway/app/app-svr/fawkes/job/internal/model/pack"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
)

var Provider = wire.NewSet(New, NewRedis, NewDB)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	// mod
	VersionByID(ctx context.Context, id int64) (*mod.Version, error)
	VersionByIDs(ctx context.Context, ids []int64) ([]*mod.Version, error)
	OriginalFile(ctx context.Context, versionID int64) (*mod.File, error)
	LastVersionList(ctx context.Context, moduleID, version, limit int64, env mod.Env) ([]*mod.Version, error)
	VersionList(ctx context.Context, moduleID int64, version []int64, env mod.Env) ([]*mod.Version, error)
	OriginalFileList(ctx context.Context, versionIDs []int64) ([]*mod.File, error)
	VersionSucceed(ctx context.Context, id int64) error
	PatchAdd(ctx context.Context, version *mod.Version, patchFiles []*mod.File) error
	DownloadFile(ctx context.Context, url, filePath string) error
	TryLock(c context.Context, key string, timeout int32) (bool, error)
	UnLock(c context.Context, key string) (err error)

	// pack
	QueryPackList(c context.Context, oc *model.OutCfg, tStart, tEnd int64, pkgType []int64, appKey string) ([]*cimdl.BuildPack, error)
	DeleteExpiredPack(c context.Context, oc *model.OutCfg, keys []*pack.BuildKey) (*pack.DeleteResp, error)
}

// dao dao.
type dao struct {
	fawkesDB *xsql.DB
	redis    *redis.Redis
	client   *bm.Client
	cache    *fanout.Fanout
}

// New new a dao and return.
func New(r *redis.Redis, db *DB) (d Dao, cf func(), err error) {
	return newDao(r, db)
}

func newDao(r *redis.Redis, db *DB) (d *dao, cf func(), err error) {
	var cfg struct {
		HTTPClientAsyn *bm.ClientConfig
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	d = &dao{
		fawkesDB: db.fawkesDB,
		redis:    r,
		client:   bm.NewClient(cfg.HTTPClientAsyn),
		cache:    fanout.New("cache"),
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) error {
	return nil
}
