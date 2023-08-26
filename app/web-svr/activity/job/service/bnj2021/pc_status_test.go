package bnj2021

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

// go test -v --count=1 pc_status_test.go pc_status.go
func TestPCStatusBiz(t *testing.T) {
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

	if err := UpdateBnjRewardCfg(); err != nil {
		t.Error(err)

		return
	}

	t.Run("resetWebViewData testing", resetWebViewDataTesting)
}

func resetWebViewDataTesting(t *testing.T) {
	if err := resetWebViewData(); err != nil {
		t.Error(err)
	}
}
