package dao

import (
	"context"
	"testing"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/time"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/model"
)

func TestAutoSubscribe(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	component.GlobalDB = sql.NewMySQL(cfg)
	if err := component.GlobalDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("test auto sub season list biz", testAutoSubSeasonList)
	t.Run("test auto sub contest detail", testAutoSubContestDetail)
	t.Run("test subscribe contest detail", testAutoSub)
}

func testAutoSub(t *testing.T) {
	req := new(model.AutoSubRequest)
	{
		req.SeasonID = 888888
		req.TeamIDList = []int64{888888}
	}
	if err := AutoSubscribeDetail(context.Background(), 88888, req); err != nil {
		t.Error(err)

		return
	}
}

func testAutoSubContestDetail(t *testing.T) {
	req := new(model.AutoSubRequest)
	{
		req.SeasonID = 888888
		req.TeamIDList = []int64{888888}
	}
	m, err := FetchAutoSubDetail(context.Background(), 8888, req)
	if err != nil {
		t.Error(err)

		return
	}

	if len(m) != 1 {
		t.Error("mid(8888) has subscribed")
	}
}

func testAutoSubSeasonList(t *testing.T) {
	list, err := FetchAutoSubSeasonList(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	if len(list) != 1 || list[0] != 666 {
		t.Error("season list should as [666]")
	}
}
