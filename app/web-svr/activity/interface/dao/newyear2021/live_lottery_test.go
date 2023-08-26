package newyear2021

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/newyear2021"

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

	component.GlobalBnjCache = redis.NewPool(redisCfg)
	component.GlobalBnjDB = sql.NewMySQL(cfg)
	if err := component.GlobalBnjDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("genTestData testing", genTestData)
	t.Run("FetchUserLiveLotteryList testing", FetchUserLiveLotteryListTesting)
	t.Run("PopUserRewardBySceneID with filled data", PopUserRewardBySceneIDWithData)
	t.Run("PopUserRewardBySceneID without filled data", PopUserRewardBySceneIDWithoutData)
}

func FetchUserLiveLotteryListTesting(t *testing.T) {
	d, err := FetchUserLiveLotteryList(context.Background(), 888888888)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func genTestData(t *testing.T) {
	conn := component.GlobalCache.Get(context.Background())
	defer func() {
		_ = conn.Close()
	}()

	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Error(err)

		return
	}

	reward := new(newyear2021.UserRewardInLiveRoom)
	{
		reward.MID = 88888888
		reward.SceneID = SceneID4Reserve
		reward.AwardId = 1
	}
	bs, _ := json.Marshal(reward)
	_, err = conn.Do("SETEX", cacheKey4UserLotteryOfReserve(reward.MID), 3600, string(bs))
	if err != nil {
		t.Errorf("genTestData1 err: %v", err)

		return
	}

	reward = new(newyear2021.UserRewardInLiveRoom)
	{
		reward.MID = 88888888
		reward.SceneID = SceneID4LiveView
		reward.AwardId = 1
	}
	bs, err = json.Marshal(reward)
	_, err = conn.Do("RPUSH", cacheKey4UserLotteryOfLive(reward.MID), string(bs))
	if err != nil {
		t.Errorf("genTestData2 err: %v", err)

		return
	}

	reward = new(newyear2021.UserRewardInLiveRoom)
	{
		reward.MID = 88888888
		reward.SceneID = SceneID4LiveView
		reward.AwardId = 1
	}
	bs, _ = json.Marshal(reward)
	_, err = conn.Do("RPUSH", cacheKey4UserLotteryOfLive(reward.MID), string(bs))
	if err != nil {
		t.Errorf("genTestData3 err: %v", err)

		return
	}
}

func PopUserRewardBySceneIDWithData(t *testing.T) {
	d, err := PopUserRewardBySceneID(context.Background(), 88888888, SceneID4LiveView)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))

	d, err = PopUserRewardBySceneID(context.Background(), 88888888, SceneID4Reserve)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ = json.Marshal(d)
	t.Log(string(bs))
}

func PopUserRewardBySceneIDWithoutData(t *testing.T) {
	d, err := PopUserRewardBySceneID(context.Background(), 888, SceneID4LiveView)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}
