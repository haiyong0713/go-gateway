package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	xtime "go-common/library/time"
)

var (
	contestDao = new(Dao)
)

// go test -v --count=1 contest_test.go contest.go auto_subscribe.go  bvid.go cache.go dao.bts.go  dao.go ld_redis.go  live_redis.go mysql.go pointdata.go  redis.go search.go tunnel_push.go
func TestContestBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = xtime.Duration(10 * time.Second)
		cfg.ExecTimeout = xtime.Duration(10 * time.Second)
		cfg.TranTimeout = xtime.Duration(10 * time.Second)
	}

	contestDao.db = sql.NewMySQL(cfg)
	if err := contestDao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	newCfg := &redis.Config{
		Name:  "local",
		Proto: "tcp",
		Addr:  "127.0.0.1:6379",
		Config: &pool.Config{
			IdleTimeout: xtime.Duration(10 * time.Second),
			Idle:        2,
			Active:      8,
		},
		WriteTimeout: xtime.Duration(10 * time.Second),
		ReadTimeout:  xtime.Duration(10 * time.Second),
		DialTimeout:  xtime.Duration(10 * time.Second),
	}
	contestDao.redis = redis.NewPool(newCfg)

	t.Run("test season list cache", ContestListByIDList)
	t.Run("recent contest id list", RecentContestIDList)
}

func RecentContestIDList(t *testing.T) {
	d, err := contestDao.RecentContestIDList(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(d)
}

func ContestListByIDList(t *testing.T) {
	m, err := contestDao.ContestListByIDList(context.Background(), []int64{76109, 76110, 76111})
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(m)
	fmt.Println(string(bs))
}
