package bnj2021

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/bnj"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"
)

// go test -v --count=1 live_lottery_test.go live_lottery.go
func TestLiveLotteryBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}
	redisCfg := &redis.Config{
		Name:  "local",
		Proto: "tcp",
		Addr:  "127.0.0.1:6379",
		Config: &pool.Config{
			IdleTimeout: time.Duration(10 * xtime.Second),
			Idle:        2,
			Active:      8,
		},
		WriteTimeout: time.Duration(10 * xtime.Second),
		ReadTimeout:  time.Duration(10 * xtime.Second),
		DialTimeout:  time.Duration(10 * xtime.Second),
	}

	component.GlobalCache = redis.NewRedis(redisCfg)
	component.GlobalDB = sql.NewMySQL(cfg)
	if err := component.GlobalDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	component.GlobalBnjDB = sql.NewMySQL(cfg)
	if err := component.GlobalBnjDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("LiveRewardReceiveBiz testing in reserve", LiveRewardReceiveBizTestingReserve)
	t.Run("LiveRewardReceiveBiz testing in AR draw", LiveRewardReceiveBizTestingARDraw)
	t.Run("LiveRewardReceiveBiz testing in live view", LiveRewardReceiveBizTestingLiveView)
}

func LiveRewardReceiveBizTestingReserve(t *testing.T) {
	reward := new(bnj.UserRewardInLiveRoom)
	{
		reward.SceneID = bnj.SceneID4Reserve
		reward.MID = 888
		reward.Duration = 600
	}

	bs, _ := json.Marshal(reward)
	err := LiveRewardReceiveBiz(context.Background(), bs)
	if err != nil {
		t.Error(err)

		return
	}
}

func LiveRewardReceiveBizTestingARDraw(t *testing.T) {
	reward := new(bnj.UserRewardInLiveRoom)
	{
		reward.SceneID = bnj.SceneID4ARDraw
		reward.MID = 888
		reward.ReceiveUnix = 1607765659
		reward.No = 1
	}

	// test ar draw reward log
	bs, _ := json.Marshal(reward)
	err := LiveRewardReceiveBiz(context.Background(), bs)
	if err != nil {
		t.Error(err)

		return
	}
}

func LiveRewardReceiveBizTestingLiveView(t *testing.T) {
	reward := new(bnj.UserRewardInLiveRoom)
	{
		reward.SceneID = bnj.SceneID4LiveView
		reward.MID = 888
		reward.Duration = 600
	}

	bs, _ := json.Marshal(reward)
	err := LiveRewardReceiveBiz(context.Background(), bs)
	if err != nil {
		t.Error(err)

		return
	}

	// test reserve lottery
	{
		reward.SceneID = bnj.SceneID4Reserve
	}
	bs, _ = json.Marshal(reward)
	err = LiveRewardReceiveBiz(context.Background(), bs)
	if err != nil {
		t.Error(err)

		return
	}
}
