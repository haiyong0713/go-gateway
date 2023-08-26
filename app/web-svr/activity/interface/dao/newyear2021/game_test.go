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

// go test -v -count=1 game_test.go game.go dao.go
func TestGameBiz(t *testing.T) {
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

	t.Run("UpsertARCoupon test", UpsertARCouponTesting)
	t.Run("DecreaseARCouponByCAS test", DecreaseARCouponByCASTesting)
	t.Run("decreaseARCouponByCAS test", decreaseARCouponByCASTesting)
	t.Run("FetchUserCoupon test", FetchUserCouponTesting)
	t.Run("FetchWebViewResource4PC test", FetchWebViewResource4PCTesting)
}

func FetchWebViewResource4PCTesting(t *testing.T) {
	d, err := FetchWebViewResource4PC(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func FetchUserCouponTesting(t *testing.T) {
	coupon, err := FetchUserCoupon(context.Background(), 88888888)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(coupon)
	t.Log(string(bs))
}

func decreaseARCouponByCASTesting(t *testing.T) {
	affectRows, err := decreaseARCouponByCAS(context.Background(), 88888888, 10, 1)
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(affectRows, err)
}

func DecreaseARCouponByCASTesting(t *testing.T) {
	err := DecreaseARCouponByCAS(context.Background(), 888888, 1)
	if err != nil {
		t.Error("decrease err should be nil", err)

		return
	}

	t.Log(err)
}

func UpsertARCouponTesting(t *testing.T) {
	arLog := new(newyear2021.ARGameLog)
	{
		arLog.MID = 888888
		arLog.Score = 666
		arLog.Coupon = 8
		arLog.Index = 1
		arLog.Date = "20201222"
	}
	err := UpsertARCoupon(context.Background(), arLog)
	if err != nil {
		t.Error(err)
	}
}
