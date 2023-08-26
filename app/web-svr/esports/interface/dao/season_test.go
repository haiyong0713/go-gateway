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
	seasonDao = new(Dao)
)

// go test -v --count=1 season_test.go season.go auto_subscribe.go  bvid.go cache.go dao.bts.go  dao.go ld_redis.go  live_redis.go mysql.go pointdata.go  redis.go search.go tunnel_push.go
func TestSeasonBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = xtime.Duration(10 * time.Second)
		cfg.ExecTimeout = xtime.Duration(10 * time.Second)
		cfg.TranTimeout = xtime.Duration(10 * time.Second)
	}

	seasonDao.db = sql.NewMySQL(cfg)
	if err := seasonDao.db.Ping(context.Background()); err != nil {
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
	seasonDao.redis = redis.NewPool(newCfg)

	t.Run("test season list cache", SeasonListByIDList)
}

func SeasonListByIDList(t *testing.T) {
	m, err := seasonDao.SeasonListByIDList([]int64{179, 180})
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(m)
	fmt.Println(string(bs))
}
