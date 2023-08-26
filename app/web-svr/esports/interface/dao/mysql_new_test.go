package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-common/library/database/sql"
	xtime "go-common/library/time"
)

var (
	dao = new(Dao)
)

// go test -v auto_subscribe.go  bvid.go cache.go dao.bts.go  dao.go ld_redis.go  live_redis.go mysql.go mysql_new_test.go pointdata.go  redis.go search.go tunnel_push.go
func TestDBOperation(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = xtime.Duration(10 * time.Second)
		cfg.ExecTimeout = xtime.Duration(10 * time.Second)
		cfg.TranTimeout = xtime.Duration(10 * time.Second)
	}

	dao.db = sql.NewMySQL(cfg)
	if err := dao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("effectiveTeamList", effectiveTeamList)
	t.Run("effectiveSeasonList", effectiveSeasonList)
	t.Run("contestList", contestList)
}

func contestList(t *testing.T) {
	d, err := dao.RawEpContests(context.Background(), []int64{7609, 7610, 7611})
	if err != nil {
		t.Error(err)
	} else {
		bs, _ := json.Marshal(d)
		fmt.Println(string(bs))
	}
}

func effectiveTeamList(t *testing.T) {
	d, err := dao.FetchEffectiveTeamList(context.Background())
	if err != nil {
		t.Error(err)
	} else {
		bs, _ := json.Marshal(d)
		fmt.Println(string(bs))
	}
}

func effectiveSeasonList(t *testing.T) {
	d, err := dao.FetchEffectiveSeasonList(context.Background())
	if err != nil {
		t.Error(err)
	} else {
		bs, _ := json.Marshal(d)
		fmt.Println(string(bs))
	}
}
