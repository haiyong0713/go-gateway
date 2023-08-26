package match_component

import (
	"context"
	"testing"
	xtime "time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
)

// go test -v -count=1 biz_test.go biz.go
func TestComponentBiz(t *testing.T) {
	dbCfg := new(sql.Config)
	{
		dbCfg.Addr = "127.0.0.1:3306"
		dbCfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		dbCfg.QueryTimeout = time.Duration(10 * xtime.Second)
		dbCfg.ExecTimeout = time.Duration(10 * xtime.Second)
		dbCfg.TranTimeout = time.Duration(10 * xtime.Second)
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
	mcCfg := &memcache.Config{
		Name:         "esport/s10",
		Proto:        "tcp",
		Addr:         "127.0.0.1:11211",
		DialTimeout:  time.Duration(10 * xtime.Second),
		ReadTimeout:  time.Duration(10 * xtime.Second),
		WriteTimeout: time.Duration(10 * xtime.Second),
	}
	mcCfg.Config = new(pool.Config)
	{
		mcCfg.Config.IdleTimeout = time.Duration(10 * xtime.Second)
		mcCfg.Config.IdleTimeout = time.Duration(10 * xtime.Second)
	}
	globalCfg := new(conf.Config)
	{
		globalCfg.AutoSubCache = redisCfg
		globalCfg.Memcached = mcCfg
		globalCfg.Memcached4UserGuess = mcCfg
		globalCfg.Mysql = dbCfg
		globalCfg.MysqlMaster = dbCfg
	}
	err := component.InitComponents()
	if err != nil {
		t.Error(err)

		return
	}

	t.Run("test FetchSeasonGuessVersionBySeasonID biz", testFetchSeasonGuessVersionBySeasonID)
	t.Run("test DeleteSeasonGuessVersionBySeasonID biz", testDeleteSeasonGuessVersionBySeasonID)
}

func testDeleteSeasonGuessVersionBySeasonID(t *testing.T) {
	err := DeleteSeasonGuessVersionBySeasonID(context.Background(), 180)
	if err != nil {
		t.Error(err)

		return
	}
}

func testFetchSeasonGuessVersionBySeasonID(t *testing.T) {
	version, err := FetchSeasonGuessVersionBySeasonID(context.Background(), 179)
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(version, err)
	if version != 1 {
		t.Errorf("season(180) guess_version should as 1, but now %v", version)
	}
}
