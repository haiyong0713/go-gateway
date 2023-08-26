package dynamicV2

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"

	"go-common/library/cache/credis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	bm "go-common/library/net/http/blademaster"

	dynactivitygrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/activity"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	dyndrawgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/draw"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicextgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic-ext"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	ipdisplaygrpc "git.bilibili.co/bapis/bapis-go/manager/operation/ip-display"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// grpc
	dynamicGRPC         dyngrpc.FeedClient
	dynamicActivityGRPC dynactivitygrpc.ActPromoRPCClient
	dynamicTopicExtGRPC dyntopicextgrpc.TopicExtClient
	dynVoteClient       dynvotegrpc.VoteSvrClient
	dyncampusClient     dyncampusgrpc.CampusSvrClient
	dynDrawGRPC         dyndrawgrpc.DrawClient
	homePageSvrClient   dyncampusgrpc.HomePageSvrClient
	ipDisplayClient     ipdisplaygrpc.OperationItemIpDisplayV1Client
	// redis
	redisSchool    credis.Redis
	redisExclusive credis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		client:         bm.NewClient(c.HTTPClient),
		redisSchool:    credis.NewRedis(c.Redis.DynamicSchool),
		redisExclusive: credis.NewRedis(c.Redis.DynamicExclusive),
	}
	var err error
	if d.dynamicGRPC, err = dyngrpc.NewClient(c.DynamicGRPC); err != nil {
		panic(err)
	}
	if d.dynamicActivityGRPC, err = dynactivitygrpc.NewClient(c.DynamicActivityGRPC); err != nil {
		panic(err)
	}
	if d.dynVoteClient, err = dynvotegrpc.NewClient(c.DynVoteGRPC); err != nil {
		panic(err)
	}
	if d.dynamicTopicExtGRPC, err = dyntopicextgrpc.NewClient(c.DynamicTopicExtGRPC); err != nil {
		panic(err)
	}
	if d.dyncampusClient, err = dyncampusgrpc.NewClient(c.DynamicCampusGRPC); err != nil {
		panic(err)
	}
	if d.dynDrawGRPC, err = dyndrawgrpc.NewClient(c.DynDrawGRPC); err != nil {
		panic(err)
	}
	if d.homePageSvrClient, err = dyncampusgrpc.NewClientHomePageSvr(c.HomePageGRPC); err != nil {
		panic(err)
	}
	if d.ipDisplayClient, err = ipdisplaygrpc.NewClientOperationItemIpDisplayV1(c.IpDisplayGRPC); err != nil {
		panic(err)
	}
	return
}

var (
	managerIpDisplayCache = unsafe.Pointer(&map[int64]string{})
	lastUpdateTime        time.Time
	managerIpFastPath     uint32
)

const (
	_managerIpDisplayUpdateInterval = 5 * time.Minute
)

// 获取管理平台指定的动态id及其定向显示的ip信息的缓存
func (d *Dao) ManagerIpDisplay(ctx context.Context) map[int64]string {
	if atomic.CompareAndSwapUint32(&managerIpFastPath, 0, 1) {
		defer func() { atomic.StoreUint32(&managerIpFastPath, 0) }()
		if time.Since(lastUpdateTime) > _managerIpDisplayUpdateInterval {
			d.updateManagerIpDisplay(ctx)
		}
	}
	return *((*map[int64]string)(atomic.LoadPointer(&managerIpDisplayCache)))
}

func (d *Dao) updateManagerIpDisplay(ctx context.Context) {
	resp, err := d.ipDisplayClient.IpDisplayRecords(ctx, &ipdisplaygrpc.IpDisplayRecordsReq{Tp: ipdisplaygrpc.IpDisplayTp_IpDisplayTpDynamic})
	if err != nil {
		log.Errorc(ctx, "error update ipDisplayClient.IpDisplayRecords: %v", err)
		return
	}
	defer func() {
		lastUpdateTime = time.Now()
	}()

	if resp != nil && resp.Result != nil {
		atomic.StorePointer(&managerIpDisplayCache, unsafe.Pointer(&resp.Result))
	} else {
		// 否则认为是清空result
		atomic.StorePointer(&managerIpDisplayCache, unsafe.Pointer(&map[int64]string{}))
	}
}
