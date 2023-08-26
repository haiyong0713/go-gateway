package fm

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/conf"
	"go-gateway/app/app-svr/app-car/job/model/fm"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

type SeasonHandler func(ctx context.Context, season *fm.CommonSeason) (railgun.MsgPolicy, error)

type Dao struct {
	c *conf.Config
	// grpc
	accCli  accountgrpc.AccountClient
	arcGrpc arcgrpc.ArchiveClient
	// db
	db *sql.DB
	// cache
	redisCli *redis.Redis
	// 多机房cache
	redisCliJd *redis.Redis
}

func New(c *conf.Config) *Dao {
	var err error
	d := &Dao{
		c:          c,
		db:         sql.NewMySQL(c.MySQL.Car),
		redisCli:   redis.NewRedis(c.Redis.Entrance),
		redisCliJd: redis.NewRedis(c.Redis.EntranceJd),
	}
	if d.accCli, err = accountgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(fmt.Sprintf("accountgrpc NewClient error (%+v)", err))
	}
	if d.arcGrpc, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("arcgrpc NewClient error (%+v)", err))
	}
	return d
}

// Profile3 get profile
func (d *Dao) Profile3(c context.Context, mid int64) (*accountgrpc.Profile, error) {
	arg := &accountgrpc.MidReq{Mid: mid}
	card, err := d.accCli.ProfileWithStat3(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return card.GetProfile(), nil
}

// Arc arc info
func (d *Dao) Arc(c context.Context, aid int64) (*arcgrpc.Arc, error) {
	arg := &arcgrpc.ArcRequest{Aid: aid}
	reply, err := d.arcGrpc.Arc(c, arg)
	if err != nil {
		log.Error("d.rpcClient.Arc(%v) error(%+v)", arg, err)
		return nil, err
	}
	if reply.Arc == nil {
		return nil, ecode.NothingFound
	}
	return reply.GetArc(), nil
}
