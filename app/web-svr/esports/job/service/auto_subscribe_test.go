package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"

	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/model"
	innerSql "go-gateway/app/web-svr/esports/job/sql"
)

func TestAutoSubscribeBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	if err := innerSql.InitByCfg(cfg); err != nil {
		t.Error(err)

		return
	}

	newCfg := &redis.Config{
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

	component.InitRedis(newCfg)
	t.Run("auto subscribe list", autoSubList)
	//t.Run("auto subscribe biz", autoSub)
}

func autoSubList(t *testing.T) {
	conn := component.GlobalAutoSubCache.Get(context.Background())
	defer func() {
		_ = conn.Close()
	}()

	detail := new(model.AutoSubscribeDetail)
	bs1, _ := json.Marshal(detail)
	resp, err := conn.Do("LPUSH", cacheKey4AutoSubscribeList, string(bs1))
	fmt.Println(resp, err)
	l, err := redis.Int64(conn.Do("LLEN", cacheKey4AutoSubscribeList))
	fmt.Println(float64(l), err)

	bs, err := redis.Bytes(conn.Do("RPOP", cacheKey4AutoSubscribeList))
	fmt.Println(string(bs), err)
}

func autoSub(t *testing.T) {
	detail := model.AutoSubscribeDetail{
		SeasonID:  888888,
		TeamId:    888888,
		ContestID: 888888,
	}
	if err := autoSubscribeByDetail(context.Background(), detail); err != nil {
		t.Error(err)
	}
}
