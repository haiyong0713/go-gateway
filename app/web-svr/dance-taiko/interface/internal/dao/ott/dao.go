package ott

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/dance-taiko/interface/api"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	accClient "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewOTTDB, NewRedis)

type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)

	LoadRanks(c context.Context, cid int64, pn int, ps int) (res []*model.PlayerHonor, err error)
	LoadFrames(c context.Context, cid int64) (res *model.FramesCache, err error)
	LoadGame(c context.Context, gameId int64) (res *model.OttGame, err error)
	LoadPlayers(c context.Context, gameId int64) (res []*model.PlayerHonor, err error)
	// cache
	CachePlayersRank(c context.Context, cid int64, mids []int64) (map[int64]int, error)
	CachePlayerScore(c context.Context, cid, mid int64) (int, error)
	DelCacheGame(c context.Context, gameId int64) error
	CachePlayer(c context.Context, gameId int64) ([]*model.PlayerHonor, error)
	AddCacheRank(c context.Context, cid int64, players []*model.PlayerHonor) error
	CachePlayerComment(c context.Context, gameId int64) (map[int64]string, error)
	CachePlayersCombo(c context.Context, gameId int64, mids []int64) (map[int64]int, error)
	AddCacheGameGap(c context.Context, gameId, gap int64) error
	DelPlayerComment(c context.Context, gameId int64) error
	AddCachePlayerStat(c context.Context, gameId, mid int64, stats []*api.StatAcc) error
	AddCacheGamePkg(c context.Context, url string) error
	CacheGamePkg(c context.Context) (string, error)
	CacheGameQRCode(c context.Context, id int64) (string, error)
	AddCacheQRCode(c context.Context, id int64, value string) error
	DelCaches(c context.Context, gameId int64, mids []int64) error
	AddCachePLayer(c context.Context, gameId int64, players []*model.PlayerHonor) error
	// grpc
	UserCards(c context.Context, mids []int64) (map[int64]*accClient.Card, error)
	// db
	CreateGame(c context.Context, aid, cid int64) (int64, error)
	StartGame(c context.Context, id int64) error
	AddPlayer(c context.Context, gameId, mid int64) error
	FinishGame(c context.Context, id int64) error
	RawPlayers(c context.Context, gameId int64) ([]*model.PlayerHonor, error)
}

type ottDB struct {
	*sql.DB
}

// dao dao.
type dao struct {
	db        *ottDB
	redis     *redis.Pool
	cache     *fanout.Fanout
	accClient accClient.AccountClient
	conf      *model.Conf
	// Expire
	rankExpire   int64
	framesExipre int64
	gameExpire   int64
}

// New new a dao and return.
func New(r *redis.Pool, db *ottDB) (d Dao, cf func(), err error) {
	return newDao(r, db)
}

func newDao(r *redis.Pool, db *ottDB) (d *dao, cf func(), err error) {
	d = &dao{
		db:    db,
		redis: r,
		cache: fanout.New("cache"),
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&d.conf); err != nil {
		panic(err)
	}
	if err = paladin.Watch("application.toml", d.conf); err != nil {
		panic(err)
	}
	d.rankExpire = int64(time.Duration(d.conf.OttCfg.Expire.RankExpire) / time.Second)
	d.framesExipre = int64(time.Duration(d.conf.OttCfg.Expire.FramesExpire) / time.Second)
	d.gameExpire = int64(time.Duration(d.conf.OttCfg.Expire.GameExpire) / time.Second)
	if d.accClient, err = accClient.NewClient(nil); err != nil {
		panic(err)
	}
	cf = d.Close
	return
}

func NewOTTDB() (db *ottDB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("steinsgate").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = &ottDB{DB: sql.NewMySQL(&cfg)}
	cf = func() { db.Close() }
	return
}

func NewRedis() (r *redis.Pool, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewPool(&cfg)
	cf = func() { r.Close() }
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

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -custom_method=d.CacheRank|d.rawRanks|d.AddCacheRanks -struct_name=dao
	LoadRanks(c context.Context, cid int64, pn int, ps int) (res []*model.PlayerHonor, err error)
	// bts: -nullcache=&model.FramesCache{Aid:-1} -check_null_code=$!=nil&&$.Aid==-1 -custom_method=d.cacheKeyFrames|d.rawKeyFrames|d.addCacheKeyFrames -struct_name=dao
	LoadFrames(c context.Context, cid int64) (res *model.FramesCache, err error)
	// bts: -nullcache=&model.OttGame{GameId:-1} -check_null_code=$!=nil&&$.GameId==-1 -custom_method=d.cacheGame|d.rawGame|d.addCacheGame -struct_name=dao
	LoadGame(c context.Context, gameId int64) (res *model.OttGame, err error)
	// bts: -custom_method=d.cachePlayer|d.rawPlayers|d.addCachePLayer -struct_name=dao
	LoadPlayers(c context.Context, gameId int64) (res []*model.PlayerHonor, err error)
}
