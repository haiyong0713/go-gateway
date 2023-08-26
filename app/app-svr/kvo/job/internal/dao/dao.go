package dao

import (
	"context"
	"encoding/json"
	"time"

	"go-gateway/app/app-svr/kvo/job/internal/model"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)

	UserConf(ctx context.Context, mid int64, moduleKey int) (userConf *model.UserConf, err error)
	Document(ctx context.Context, checkSum int64) (rm json.RawMessage, err error)

	TxUpDocuement(ctx context.Context, tx *sql.Tx, checkSum int64, data string, now time.Time) (err error)
	TxUpUserConf(ctx context.Context, tx *sql.Tx, mid int64, moduleKey int, checkSum int64, now time.Time) (err error)

	AsyncSetUserConf(ctx context.Context, uc *model.UserConf)
	AsyncSetDocument(ctx context.Context, checkSum int64, bm json.RawMessage)

	SetUserConf(ctx context.Context, uc *model.UserConf)
	SetDocument(ctx context.Context, checkSum int64, bm json.RawMessage)

	UserConfRds(ctx context.Context, mid int64, moduleKey int) (uc *model.UserConf, err error)
	DocumentRds(ctx context.Context, checkSum int64) (data json.RawMessage, err error)

	BeginTx(c context.Context) (*sql.Tx, error)

	SetUserDocRds(ctx context.Context, mid int64, buvid string, moduleKey int, bs []byte) (err error)
	UserDocRds(ctx context.Context, mid int64, buvid string, moduleKey int) (bs []byte, err error)
}

// dao dao.
type dao struct {
	db             *sql.DB
	redis          *redis.Redis
	cache          *fanout.Fanout
	rdsExpire      int32
	rdsUcDocExpire int32
}

// New new a dao and return.
func New(redis *redis.Redis, db *sql.DB) (d Dao, err error) {
	var cfg struct {
		RdsExpire      xtime.Duration
		RdsUcDocExpire xtime.Duration
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:             db,
		redis:          redis,
		cache:          fanout.New("cache"),
		rdsExpire:      int32(time.Duration(cfg.RdsExpire) / time.Second),
		rdsUcDocExpire: int32(time.Duration(cfg.RdsUcDocExpire) / time.Second),
	}
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.db.Close()
	d.redis.Close()
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

// BeginTx begin trans
func (d *dao) BeginTx(c context.Context) (*sql.Tx, error) {
	return d.db.Begin(c)
}
