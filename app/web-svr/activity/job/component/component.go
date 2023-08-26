package component

import (
	"context"
	databusv2 "go-common/library/queue/databus.v2"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"

	"go-gateway/app/web-svr/activity/job/conf"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/errgroup.v2"
)

var (
	GlobalDB                            *sql.DB
	GlobalDBOfRead                      *sql.DB
	GlobalBnjDB                         *sql.DB
	GlobalRewardsDB                     *sql.DB
	GlobalCache                         *redis.Redis
	GlobalRedis                         *redis.Redis
	GlobalRedisStore                    *redis.Redis
	BackUpMQ                            *redis.Redis
	DatabusClient                       databusv2.Client
	BGroupMessagePub                    databusv2.Producer
	LotteryAddTimesPub                  databusv2.Producer
	UpActReserveRelationPub             databusv2.Producer
	UpActReservePub                     databusv2.Producer
	UpActReserveRelationTableMonitor    databusv2.Producer
	UpActReserveRelationChannelAudit    databusv2.Producer
	UpActReserveLotteryUserReserveState databusv2.Producer
)

func InitByCfg(cfg *conf.Config) error {
	var err error
	GlobalDB = sql.NewMySQL(cfg.MySQL.Like)
	GlobalDBOfRead = sql.NewMySQL(cfg.MySQL.Read)
	GlobalBnjDB = sql.NewMySQL(cfg.MySQL.Bnj)
	GlobalRewardsDB = sql.NewMySQL(cfg.RewardsMySQL)
	GlobalCache = redis.NewRedis(cfg.S10PointShopRedis)
	GlobalRedis = redis.NewRedis(cfg.Redis.Cache)
	GlobalRedisStore = redis.NewRedis(cfg.Redis.Store)
	BackUpMQ = redis.NewRedis(cfg.BackupMQ)
	DatabusClient, err = databusv2.NewClient(
		context.Background(),
		cfg.Databus.Target,
		databusv2.WithAppID(cfg.Databus.AppID),
		databusv2.WithToken(cfg.Databus.Token),
	)
	if err != nil {
		panic(err)
	}
	BGroupMessagePub = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.BGroupMessage)
	LotteryAddTimesPub = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.LotteryAddTimes)
	UpActReserveRelationPub = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.UpActReserveRelation)
	UpActReservePub = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.UpActReserve)
	UpActReserveRelationTableMonitor = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.UpActReserveRelationTableMonitor)
	UpActReserveRelationChannelAudit = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.UpActReserveRelationChannelAudit)
	UpActReserveLotteryUserReserveState = initialize.NewProducer(DatabusClient, cfg.Databus.Topic.UpActReserveLotteryUserReserveState)
	return Ping()
}

func Ping() (err error) {
	g := errgroup.WithContext(context.Background())
	g.Go(func(ctx context.Context) error {
		return GlobalDB.Ping(ctx)
	})
	g.Go(func(ctx context.Context) error {
		return GlobalDBOfRead.Ping(ctx)
	})
	g.Go(func(ctx context.Context) error {
		return GlobalBnjDB.Ping(ctx)
	})
	return g.Wait()
}

func Close() (err error) {
	g := errgroup.WithContext(context.Background())
	g.Go(func(ctx context.Context) error {
		return GlobalDB.Close()
	})
	g.Go(func(ctx context.Context) error {
		return GlobalDBOfRead.Close()
	})
	g.Go(func(ctx context.Context) error {
		return GlobalBnjDB.Close()
	})
	g.Go(func(ctx context.Context) error {
		return GlobalCache.Close()
	})
	g.Go(func(ctx context.Context) error {
		return GlobalRedis.Close()
	})
	g.Go(func(ctx context.Context) error {
		return GlobalRedisStore.Close()
	})
	g.Go(func(ctx context.Context) error {
		return BackUpMQ.Close()
	})
	g.Go(func(ctx context.Context) error {
		return BGroupMessagePub.Close()
	})
	g.Go(func(ctx context.Context) error {
		return LotteryAddTimesPub.Close()
	})
	g.Go(func(ctx context.Context) error {
		return UpActReserveRelationChannelAudit.Close()
	})
	g.Go(func(ctx context.Context) error {
		return UpActReserveLotteryUserReserveState.Close()
	})
	return g.Wait()
}
