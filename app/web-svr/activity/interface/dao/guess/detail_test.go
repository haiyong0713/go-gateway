package guess

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"
)

var (
	detailDao *Dao
)

// go test -v --count=1 cache.go dao.bts.go dao.go detail.go detail_test.go guess.go guess_main.go mysql.go redis.go
func TestDetailBiz(t *testing.T) {
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

	detailDao = new(Dao)
	detailDao.db = sql.NewMySQL(cfg)
	detailDao.redis = redis.NewPool(redisCfg)
	if err := detailDao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("main detail map 4 hot data", AvailableMainDetailMap)
	t.Run("main detail map 4 user biz", DetailListByMainIDList)
	t.Run("delete main cache data by main id", DeleteGuessAggregationInfoCache)
}

func DeleteGuessAggregationInfoCache(t *testing.T) {
	if err := detailDao.DeleteGuessAggregationInfoCache(context.Background(), 1033, 1); err != nil {
		t.Error(err)
	}
}

func DetailListByMainIDList(t *testing.T) {
	d, err := detailDao.DetailListByMainIDList(context.Background(), []int64{1088}, 1)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func AvailableMainDetailMap(t *testing.T) {
	d, err := detailDao.AvailableHotMainDetailMap(context.Background(), []int64{1033, 1031}, 1)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}
