package bnj

import (
	"context"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/activity/job/component"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"
)

// go test -v -count=1 reserve_lottery_2021_test.go reserve_lottery_2021.go
func TestReserveLottery2021Biz(t *testing.T) {
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

	t.Run("FetchLastReceiveIDByCountAndSuffix testing", FetchLastReceiveIDByCountAndSuffixTesting)
	t.Run("UpsertReserveRewardLog testing", UpsertReserveRewardLogTesting)
}

func UpsertReserveRewardLogTesting(t *testing.T) {
	affectedRows, err := UpsertReserveRewardLog(context.Background(), 666, 2, 2, "888")
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(affectedRows)
}

func FetchLastReceiveIDByCountAndSuffixTesting(t *testing.T) {
	lastID, err := FetchLastReceiveIDByCountAndSuffix(context.Background(), 0, "88")
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(lastID)
}
