package lol

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/time"

	"go-gateway/app/web-svr/activity/interface/service"
)

func TestUnSettlementContestIDList(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	service.GlobalDB = sql.NewMySQL(cfg)
	ctx := context.Background()
	if err := service.GlobalDB.Ping(ctx); err != nil {
		t.Error(err)

		return
	}

	coinsAfterConvert, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", float64(509)/_multi), 64)
	if coinsAfterConvert != 5.1 {
		t.Error(coinsAfterConvert)
	}

	if _, err := UnSettlementContestIDList(ctx); err != nil {
		t.Error(err)
	}
}
