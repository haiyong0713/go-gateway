package bplus

import (
	"fmt"

	"go-common/library/cache/redis"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynctopic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
)

// Dao is favorite dao
type Dao struct {
	client    *httpx.Client
	favorPlus string

	groupsCount   string
	dynamicDetail string
	dynamicTopics string
	// redis
	redis *redis.Pool
	// databus
	pub *databus.Databus
	// contribute cace
	contributeExpire int32
	// topic client
	topicClient dynctopic.TopicClient
	// dynamic feed grpc
	dynGrpc dyngrpc.FeedClient
}

// New initial favorite dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:        httpx.NewClient(c.HTTPBPlus),
		favorPlus:     c.Host.APILiveCo + _favorPlus,
		groupsCount:   c.Host.VC + _groupsCount,
		dynamicDetail: c.Host.VC + _dynamicDetail,
		dynamicTopics: c.Host.VC + _dynamicTopics,
		redis:         redis.NewPool(c.Redis.Contribute.Config),
		pub:           databus.New(c.ContributePub),
		// contribute cace
		contributeExpire: 60 * 60 * 24 * 5,
	}
	var err error
	if d.topicClient, err = dynctopic.NewClient(c.DynamicTopicGRPC); err != nil {
		panic(fmt.Sprintf("dynamic topic grpc NewClientt error (%+v)", err))
	}
	if d.dynGrpc, err = dyngrpc.NewClient(c.DynGRPC); err != nil {
		panic(fmt.Sprintf("dynGrpc NewClient error(%v)", err))
	}
	return
}
