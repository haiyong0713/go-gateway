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
	guessMainDao *Dao
)

// go test -v --count=1 cache.go dao.bts.go dao.go detail.go guess_main_test.go guess.go guess_main.go mysql.go redis.go
func TestGuessMainBiz(t *testing.T) {
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

	guessMainDao = new(Dao)
	guessMainDao.db = sql.NewMySQL(cfg)
	guessMainDao.redis = redis.NewPool(redisCfg)
	if err := guessMainDao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("fetch hot mainID list", HotMainIDList)
	t.Run("fetch mainID list by oid", MainIDListByOID)
	t.Run("delete mainID cache by oid", DeleteMainIDCacheByOID)
	t.Run("reset mainID list cache by oid", ResetMainIDListInCacheByOID)
	t.Run("hot mainRes in memory cache", HotMainResMap)
	t.Run("hot main map by mainID list", TestHotMainMapByMainIDList)
}

func TestHotMainMapByMainIDList(t *testing.T) {
	list, err := guessMainDao.HotMainIDList(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	m := GenHotMainMapByMainIDList(list)
	bs, _ := json.Marshal(m)
	t.Log(string(bs), len(list), len(m))
}

func HotMainResMap(t *testing.T) {
	list, m, err := guessMainDao.HotMainResMap(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	bs1, _ := json.Marshal(list)
	bs2, _ := json.Marshal(m)
	t.Log(string(bs1), string(bs2), len(list), len(m))
}

func ResetMainIDListInCacheByOID(t *testing.T) {
	list, err := guessMainDao.ResetMainIDListInCacheByOID(context.Background(), 7466, 1)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(list)
	t.Log(string(bs))
}

func DeleteMainIDCacheByOID(t *testing.T) {
	if err := guessMainDao.DeleteMainIDCache(context.Background(), 7470, 1); err != nil {
		t.Error(err)
	}
}

func MainIDListByOID(t *testing.T) {
	list, err := guessMainDao.MainIDListByOID(context.Background(), 7470, 1)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(list)
	t.Log(string(bs))
}

func HotMainIDList(t *testing.T) {
	list, err := guessMainDao.HotMainIDList(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(list)
	t.Log(string(bs))
}
