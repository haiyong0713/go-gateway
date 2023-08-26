package dao

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	pb "go-gateway/app/app-svr/collection-splash/api"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	AddSplash(ctx context.Context, param *pb.AddSplashReq) (int64, error)
	UpdateSplash(ctx context.Context, param *pb.UpdateSplashReq) (int64, error)
	DeleteSplash(ctx context.Context, param *pb.SplashReq) (int64, error)
	Splash(ctx context.Context, param *pb.SplashReq) (*pb.Splash, error)
	RawSplashList(ctx context.Context) ([]*pb.Splash, error)
	SplashList(ctx context.Context) ([]*pb.Splash, error)
	CacheSplashList(ctx context.Context) ([]*pb.Splash, bool, error)
	AddCacheSplashList(ctx context.Context, list []*pb.Splash) error
}

// dao dao.
type dao struct {
	db    *sql.DB
	redis *redis.Redis
}

// New new a dao and return.
func New(db *sql.DB, r *redis.Redis) (d Dao, cf func(), err error) {
	return newDao(db, r)
}

// nolint:unparam
func newDao(db *sql.DB, r *redis.Redis) (d *dao, cf func(), err error) {
	d = &dao{
		db:    db,
		redis: r,
	}
	cf = d.Close
	return
}

func (d *dao) SplashList(ctx context.Context) ([]*pb.Splash, error) {
	res, isNil, err := d.CacheSplashList(ctx)
	if err != nil {
		return nil, err
	}
	if !isNil {
		return res, nil
	}
	res, err = d.RawSplashList(ctx)
	if err != nil {
		return nil, err
	}
	if err = d.AddCacheSplashList(ctx, res); err != nil {
		return res, nil
	}
	return res, nil
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
