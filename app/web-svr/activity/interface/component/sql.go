package component

import (
	"context"
	"go-common/library/log"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/interface/conf"

	"go-common/library/database/sql"
)

var (
	GlobalDB *sql.DB

	GlobalTiDB  *sql.DB
	S10GlobalDB *sql.DB
)

var (
	GlobalMC       *memcache.Memcache
	S10GlobalMC    *memcache.Memcache
	S10PointCostMC *memcache.Memcache
)

var (
	GlobalRedis      *redis.Redis
	GlobalRedisStore *redis.Redis
	TimeMachineRedis *redis.Redis
	GlobalVoteRedis  *redis.Redis
	GlobalStockRedis *redis.Redis
)

var (
	S10PointShopRedis *redis.Pool
)

var (
	GlobalCache *redis.Pool

	GlobalBnjDB     *sql.DB
	GlobalRewardsDB *sql.DB
	GlobalBnjCache  *redis.Pool
	BackUpMQ        *redis.Redis
)

func InitByCfg(cfg *conf.Config) error {
	GlobalDB = sql.NewMySQL(cfg.MySQL.Like)
	GlobalTiDB = sql.NewMySQL(cfg.TiDB)
	GlobalMC = memcache.New(cfg.Memcache.Like)
	S10GlobalDB = sql.NewMySQL(cfg.S10MySQL)
	S10GlobalMC = memcache.New(cfg.S10MC)
	S10PointCostMC = memcache.New(cfg.S10PointCostMC)
	S10PointShopRedis = redis.NewPool(cfg.S10PointShopRedis)
	TimeMachineRedis = redis.NewRedis(cfg.Timemachine.Redis)
	GlobalRedis = redis.NewRedis(cfg.Redis.Cache)
	GlobalRedisStore = redis.NewRedis(cfg.Redis.Store)
	GlobalBnjDB = sql.NewMySQL(cfg.MySQL.Bnj)
	GlobalCache = redis.NewPool(cfg.Redis.Config)
	GlobalBnjCache = S10PointShopRedis
	BackUpMQ = redis.NewRedis(cfg.BackupMQ)
	GlobalRewardsDB = sql.NewMySQL(cfg.RewardsMySQL)
	GlobalVoteRedis = redis.NewRedis(cfg.VoteRedis)
	GlobalStockRedis = redis.NewRedis(cfg.StockRedis.Config)
	initFanout(cfg)
	initES(cfg)
	return Ping()
}

func Ping() error {
	g := errgroup.WithContext(context.Background())
	if GlobalDB != nil {
		g.Go(func(ctx context.Context) (err error) {
			err = GlobalDB.Ping(ctx)
			if err != nil {
				log.Errorc(ctx, "GlobalDB.Ping(ctx) err[%v]", err)
			}
			return
		})
	}
	if GlobalTiDB != nil {
		g.Go(func(ctx context.Context) (err error) {
			err = GlobalTiDB.Ping(ctx)
			if err != nil {
				log.Errorc(ctx, "GlobalTiDB.Ping(ctx) err[%v]", err)
			}
			return
		})
	}
	if S10GlobalDB != nil {
		g.Go(func(ctx context.Context) (err error) {
			err = S10GlobalDB.Ping(ctx)
			if err != nil {
				log.Errorc(ctx, "S10GlobalDB.Ping(ctx) err[%v]", err)
			}
			return
		})
	}
	if GlobalBnjDB != nil {
		g.Go(func(ctx context.Context) (err error) {
			err = GlobalBnjDB.Ping(ctx)
			if err != nil {
				log.Errorc(ctx, "GlobalBnjDB.Ping(ctx) err[%v]", err)
			}
			return
		})
	}
	if GlobalRewardsDB != nil {
		g.Go(func(ctx context.Context) (err error) {
			err = GlobalRewardsDB.Ping(ctx)
			if err != nil {
				log.Errorc(ctx, "GlobalRewardsDB.Ping(ctx) err[%v]", err)
			}
			return
		})
	}
	return g.Wait()
}

func Close() error {
	g := errgroup.WithContext(context.Background())
	if GlobalDB != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalDB.Close()
		})
	}
	if GlobalTiDB != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalTiDB.Close()
		})
	}
	if GlobalMC != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalMC.Close()
		})
	}
	if S10GlobalDB != nil {
		g.Go(func(ctx context.Context) error {
			return S10GlobalDB.Close()
		})
	}
	if S10GlobalMC != nil {
		g.Go(func(ctx context.Context) error {
			return S10GlobalMC.Close()
		})
	}
	if S10PointShopRedis != nil {
		g.Go(func(ctx context.Context) error {
			return S10PointShopRedis.Close()
		})
	}
	if TimeMachineRedis != nil {
		g.Go(func(ctx context.Context) error {
			return TimeMachineRedis.Close()
		})
	}
	if GlobalRedis != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalRedis.Close()
		})
	}
	if GlobalRedisStore != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalRedisStore.Close()
		})
	}
	if GlobalBnjDB != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalBnjDB.Close()
		})
	}
	if GlobalCache != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalCache.Close()
		})
	}
	if GlobalBnjCache != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalBnjCache.Close()
		})
	}
	if DWInfo != nil {
		g.Go(func(ctx context.Context) error {
			return DWInfo.Close()
		})
	}
	if BackUpMQ != nil {
		g.Go(func(ctx context.Context) error {
			return BackUpMQ.Close()
		})
	}
	if DatabusV2ActivityClient != nil {
		g.Go(func(ctx context.Context) error {
			DatabusV2ActivityClient.Close()
			return nil
		})
	}
	if ReserveFanout != nil {
		g.Go(func(ctx context.Context) error {
			ReserveFanout.Close()
			return nil
		})
	}
	if GlobalRewardsDB != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalRewardsDB.Close()
		})
	}

	if GlobalVoteRedis != nil {
		g.Go(func(ctx context.Context) error {
			return GlobalVoteRedis.Close()
		})
	}
	return g.Wait()
}
