package newyear2021

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

var testService *Service

func init() {
	flag.Set("conf", "../../cmd/activity-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if err := component.InitByCfg(conf.Conf.MySQL.Like, conf.Conf.Redis.Config); err != nil {
		panic(err)
	}
	testService = New(conf.Conf)
}

// go test -v notify.go rewards.go rewards_test.go service.go service_test.go task.go
func TestUpdateConfig(t *testing.T) {
	ctx := context.Background()

	//update config and wait refresh
	config := testService.GetConf()
	config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount = 15000
	assert.Equal(t, nil, testService.dao.UpdateConf(ctx, config))
	time.Sleep(16 * time.Second) //wait config update loop.
	config = testService.GetConf()
	assert.Equal(t, int64(15000), config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount)

	///update config and wait refresh
	config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount = 34000
	assert.Equal(t, nil, testService.dao.UpdateConf(ctx, config))
	time.Sleep(16 * time.Second) //wait config update loop.
	config = testService.GetConf()
	assert.Equal(t, int64(34000), config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount)

}
