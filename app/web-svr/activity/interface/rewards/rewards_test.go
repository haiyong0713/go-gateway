package rewards

import (
	"context"
	"flag"
	"fmt"
	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/api"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

func init() {
	_ = flag.Set("conf", "../cmd/activity-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	component.GlobalBnjCache = redis.NewPool(conf.Conf.Redis.Config)
	component.BackUpMQ = redis.NewRedis(conf.Conf.Redis.Config)
	Init(conf.Conf)
}

//make sure run dao unit test first to write configs to db
//func TestSendReward(t *testing.T) {
//	ctx := context.Background()
//	testMid := int64(216761)
//	fakeUniqueId := 0
//	cs, err := Client.GetAwards(ctx, 0)
//	assert.Equal(t, nil, err)
//	assert.NotEqual(t, nil, cs)
//	assert.NotEqual(t, 0, len(cs))
//	for _, c := range cs {
//		_, err := Client.SendAwardById(ctx, testMid, fmt.Sprintf("unittest-sync-%v-%v-%v", testMid, c.Id, fakeUniqueId), "TEST", c.Id)
//		if err == nil {
//			t.Logf("send %v award success\n", c.DisplayName)
//		} else {
//			t.Errorf("send %v award fail, error: %v\n", c.DisplayName, err)
//		}
//	}
//}

func TestSendRewardAsync(t *testing.T) {
	ctx := context.Background()
	testMid := int64(216761)
	fakeUniqueId := 0
	cs, err := Client.GetAwards(ctx, &api.RewardsListAwardReq{
		ActivityId: 0,
	})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, cs)
	assert.NotEqual(t, 0, len(cs))
	for _, c := range cs {
		_, err := Client.SendAwardByIdAsync(ctx, testMid, fmt.Sprintf("unittest-async-%v-%v-%v", testMid, c.Id, fakeUniqueId), "TEST", c.Id, true, true)
		if err == nil {
			t.Logf("send %v award success\n", c.Name)
		} else {
			t.Errorf("send %v award fail, error: %v\n", c.Name, err)
		}
	}
}
