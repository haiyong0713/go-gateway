package bnj

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

// go test -v -count=1 live_lottery_2021_test.go live_lottery_2021.go
func TestLiveLottery2021Biz(t *testing.T) {
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
	component.GlobalBnjDB = sql.NewMySQL(cfg)
	if err := component.GlobalBnjDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("InsertUnReceivedUserInfo testing", InsertUnReceivedUserInfoTesting)
	t.Run("SetUserBnjLotteryInfo testing", SetUserBnjLotteryInfoTesting)
	t.Run("FetchBnjUnReceivedUserList testing", FetchBnjUnReceivedUserListTesting)
	t.Run("FetchBnjLotteryRuleFor2021 testing", FetchBnjLotteryRuleFor2021Testing)
	t.Run("UpdateLastIDByDurationAndSuffix testing", UpdateLastIDByDurationAndSuffixTesting)
	t.Run("FetchLastReceiveIDByDurationAndSuffix testing", FetchLastReceiveIDByDurationAndSuffixTesting)
	t.Run("FetchBnjLiveUserRecordList testing", FetchBnjLiveUserRecordListTesting)
}

func FetchBnjLiveUserRecordListTesting(t *testing.T) {
	d, err := FetchBnjLiveUserRecordList(context.Background(), 666)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func FetchLastReceiveIDByDurationAndSuffixTesting(t *testing.T) {
	d, err := FetchLastReceiveIDByDurationAndSuffix(context.Background(), 600, "00")
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(d)
}

func UpdateLastIDByDurationAndSuffixTesting(t *testing.T) {
	err := UpdateLastIDByDurationAndSuffix(context.Background(), 600, 888, "00")
	if err != nil {
		t.Error(err)
	}
}

func FetchBnjLotteryRuleFor2021Testing(t *testing.T) {
	list, err := FetchBnjLotteryRuleFor2021(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(list)
	t.Log(string(bs))
}

func FetchBnjUnReceivedUserListTesting(t *testing.T) {
	list, err := FetchBnjUnReceivedUserList(context.Background(), "88", 100, 600, 100)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(list)
	t.Log(string(bs))
}

func InsertUnReceivedUserInfoTesting(t *testing.T) {
	info := new(bnj.UserInLiveRoomFor2021)
	{
		info.Duration = 600
		info.MID = 888
		info.UniqueID = "3"
		info.AID = 8
		info.AType = "bnj"
	}
	err := InsertUnReceivedUserInfo(context.Background(), info)
	if err != nil {
		t.Error(err)
	}
}

func SetUserBnjLotteryInfoTesting(t *testing.T) {
	rows, err := UpdateUserRewardInLive(context.Background(), 888, 600, "reward888")
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(rows)
}
