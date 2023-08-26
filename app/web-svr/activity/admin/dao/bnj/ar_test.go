package bnj

import (
	"context"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/activity/admin/component"
	model "go-gateway/app/web-svr/activity/admin/model/bnj"

	"go-common/library/database/sql"
	"go-common/library/time"
)

// go test -v --count=1 ar_test.go ar.go
func TestARBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	component.GlobalDB = sql.NewMySQL(cfg)

	t.Run("AddARSetting testing", AddARSettingTesting)
	t.Run("UpsertARScore2Coupon testing", UpsertARScore2CouponTesting)
	t.Run("DeleteARScore2Coupon testing", DeleteARScore2CouponTesting)
}

func DeleteARScore2CouponTesting(t *testing.T) {
	tmp := new(model.Score2CouponRule)
	{
		tmp.ID = 6
		tmp.Score = 150
		tmp.Coupon = 3
	}
	err := DeleteARScore2Coupon(context.Background(), tmp)
	if err != nil {
		t.Error(err)
	}
}

func UpsertARScore2CouponTesting(t *testing.T) {
	tmp := new(model.Score2CouponRule)
	{
		tmp.Score = 100
		tmp.Coupon = 2
	}

	err := UpsertARScore2Coupon(context.Background(), tmp)
	if err != nil {
		t.Error(err)

		return
	}

	{
		tmp.ID = 1
		tmp.Score = 150
		tmp.Coupon = 3
	}
	err = UpsertARScore2Coupon(context.Background(), tmp)
	if err != nil {
		t.Error(err)
	}
}

func AddARSettingTesting(t *testing.T) {
	err := AddARSetting(context.Background(), `{"day_times":3}`)
	if err != nil {
		t.Error(err)
	}
}
