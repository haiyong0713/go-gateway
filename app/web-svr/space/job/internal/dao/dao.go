package dao

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/space/interface/model"
	jobmdl "go-gateway/app/web-svr/space/job/internal/model"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	TopPhotoArc(ctx context.Context, mid int64) (*model.TopPhotoArc, error)
	TopPhotoArcCancel(ctx context.Context, mid int64) (int64, error)
	DelCacheTopPhotoArc(ctx context.Context, mid int64) error
	SendLetter(ctx context.Context, arg *jobmdl.LetterParam) error
	DelCachePrivacySetting(ctx context.Context, mid int64) error
	SetLivePlaybackWhitelist(ctx context.Context) error
}

// dao dao.
type dao struct {
	db                       *sql.DB
	redis                    *redis.Redis
	cache                    *fanout.Fanout
	topPhotoArcExpire        int32
	httpClient               *bm.Client
	msgKeyURL                string
	sendMsgURL               string
	livePlaybackWhitelistURL string
}

// New new a dao and return.
func New(db *sql.DB, r *redis.Redis) (d Dao, cf func(), err error) {
	return newDao(db, r)
}

func newDao(db *sql.DB, r *redis.Redis) (d *dao, cf func(), err error) {
	var cfg struct {
		TopPhotoArcExpire xtime.Duration
		HTTPClient        *bm.ClientConfig
		Host              struct {
			Dynamic string
		}
		LivePlayback struct {
			WhitelistURL string
		}
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:                       db,
		redis:                    r,
		cache:                    fanout.New("cache"),
		topPhotoArcExpire:        int32(time.Duration(cfg.TopPhotoArcExpire) / time.Second),
		httpClient:               bm.NewClient(cfg.HTTPClient),
		msgKeyURL:                cfg.Host.Dynamic + _msgKeyURI,
		sendMsgURL:               cfg.Host.Dynamic + _sendMsgURI,
		livePlaybackWhitelistURL: cfg.LivePlayback.WhitelistURL,
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
