package dao

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/space/service/api"
	"go-gateway/app/web-svr/space/service/internal/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	// bts: -nullcache=[]*model.MemberPrivacy{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1 -singleflight=true
	PrivacySetting(ctx context.Context, req *pb.PrivacySettingReq) (reply []*model.MemberPrivacy, err error)
	UpdatePrivacySetting(ctx context.Context, req *pb.UpdatePrivacysReq) error
	CacheLivePlaybackWhitelist(ctx context.Context) (map[int64]struct{}, error)
}

// dao dao.
type dao struct {
	db                      *sql.DB
	redis                   *redis.Redis
	accClient               accgrpc.AccountClient
	cache                   *fanout.Fanout
	spaceExpire             int32
	cacheRand               int32
	settingNewUserTimePoint string
}

// New new a dao and return.
func New(r *redis.Redis, db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(r, db)
}

func newDao(r *redis.Redis, db *sql.DB) (d *dao, cf func(), err error) {
	var cfg struct {
		SpaceExpire    xtime.Duration
		CacheRand      int32
		AccGRPC        *warden.ClientConfig
		SettingNewUser struct {
			TimePoint string
		}
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:                      db,
		redis:                   r,
		cache:                   fanout.New("cache"),
		spaceExpire:             int32(time.Duration(cfg.SpaceExpire) / time.Second),
		cacheRand:               cfg.CacheRand,
		settingNewUserTimePoint: cfg.SettingNewUser.TimePoint,
	}
	if d.accClient, err = accgrpc.NewClient(cfg.AccGRPC); err != nil {
		return
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
