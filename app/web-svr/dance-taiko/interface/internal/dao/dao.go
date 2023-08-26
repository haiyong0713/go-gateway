package dao

import (
	"context"
	"io"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis, NewMC)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)

	RawGame(c context.Context, gid int64) (*model.Game, error)
	GatherExamples(c context.Context, aid int64, stats []*model.Stat) error
	AllAids(c context.Context) ([]int64, error)
	PickExamples(c context.Context, aid int64) ([]*model.Stat, error)
	CreateGame(c context.Context, avid int64, gameId int64) error
	CurrentGame(c context.Context) (*model.Game, error)
	UpdateGameStatus(c context.Context, gameId int64, status string) error
	StartGame(c context.Context, gameId int64, status string) error

	RedisJoin(c context.Context, gameId int64, position int, player string) (err error)
	RedisGetJoinedPlayers(c context.Context, gameId int64) (players map[int]string, err error)
	RedisIncrPoint(c context.Context, gameId int64, mid int64, delta int64) (err error)
	RedisGetPoints(c context.Context, gameId int64) (map[int64]int64, error)
	RedisDelPoints(c context.Context, gameId int64) (err error)
	RedisSetComment(c context.Context, gameId int64, mid int64, comment string) (err error)
	RedisGetComments(c context.Context, gameId int64) (res map[int64]string, err error)
	RedisSetUserPoints(c context.Context, aid int64, points map[int64]int64) (err error)
	RedisGetUserPoints(c context.Context, aid int64, number int64) ([]*model.PlayerHonor, error)
	RedisSetGame(c context.Context, gameId int) error
	RedisGetGame(c context.Context, gameId int) (int, error)
	RedisSetFilePath(c context.Context, filePath string) (err error)
	RedisGetFilePath(c context.Context) (filePath string, err error)
	AddRedisExperiment(c context.Context, gameId int64) error
	RedisExperiment(c context.Context, gameId int64) (bool, error)

	BfsUpload(c context.Context, fileType string, body io.Reader) (location string, err error)

	//bws
	BwsCreateRoom(c context.Context) (int, error)
	BwsMidInfo(c context.Context, mid int64, gameId int) (*model.BwsPlayInfo, error)
	BwsStartGame(c context.Context, gameId int) error
	BwsJoinRoom(c context.Context, gameId int, mid int64) error
	BwsEndGame(c context.Context, gameId int, players []*model.BwsPlayResult) error
	BwsReset(c context.Context)
}

// dao dao.
type dao struct {
	db         *sql.DB
	redis      *redis.Redis
	mc         *memcache.Memcache
	cache      *fanout.Fanout
	bfs        *BfsClient
	demoExpire int32
	conf       *model.Conf
	client     *bm.Client
	httpCfg    *model.HttpConfig
}

// New new a dao and return.
func New(r *redis.Redis, mc *memcache.Memcache, db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(r, mc, db)
}

func newDao(r *redis.Redis, mc *memcache.Memcache, db *sql.DB) (d *dao, cf func(), err error) {
	d = &dao{
		db:    db,
		redis: r,
		mc:    mc,
		cache: fanout.New("cache"),
		bfs:   &BfsClient{},
	}
	_ = d.bfs.New()
	if err = paladin.Get("application.toml").UnmarshalTOML(&d.conf); err != nil {
		panic(err)
	}
	if err := paladin.Watch("application.toml", d.conf); err != nil {
		panic(err)
	}
	if err = paladin.Get("http.toml").UnmarshalTOML(&d.httpCfg); err != nil {
		panic(err)
	}
	d.client = bm.NewClient(&d.httpCfg.DanceClient)
	d.demoExpire = int32(time.Duration(d.conf.DemoExpire) / time.Second)
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
