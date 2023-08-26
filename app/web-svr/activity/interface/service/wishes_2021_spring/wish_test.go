package wishes_2021_spring

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/time"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/wishes_2021_spring"
)

// go test -v --count=1 wish_test.go wish.go
func TestWishBiz(t *testing.T) {
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
	component.BackUpMQ = redis.NewRedis(redisCfg)
	component.GlobalBnjDB = sql.NewMySQL(cfg)
	component.S10GlobalMC = memcache.New(mcCfg)
	if err := component.GlobalBnjDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	activityCfg4WishTree := new(model.CommonActivityConfig)
	{
		activityCfg4WishTree.ActivityID = 1
		activityCfg4WishTree.MaxUploadTimes = 20
		activityCfg4WishTree.StartTime = 1615442400
		activityCfg4WishTree.EndTime = 1615528800
		activityCfg4WishTree.UniqID = "wish_tree"
	}
	activityCfg4Fool := new(model.CommonActivityConfig)
	{
		activityCfg4Fool.ActivityID = 2
		activityCfg4Fool.MaxUploadTimes = 2
		activityCfg4Fool.StartTime = xtime.Now().Unix() + 666
		activityCfg4Fool.EndTime = xtime.Now().Unix() + 6666
		activityCfg4Fool.UniqID = "fool"
	}
	innerActivityMap[activityCfg4WishTree.UniqID] = activityCfg4WishTree
	innerActivityMap[activityCfg4Fool.UniqID] = activityCfg4Fool

	t.Run("test CommitUserContent biz", testCommitUserContent)
	t.Run("test CommitUserContent biz with invalid time range", testCommitUserContentWithInvalidTimeRange)
	t.Run("test CommitUserManuScript biz", testCommitUserManuScript)
	t.Run("test UserCommitContent4Aggregation biz", testUserCommitContent4Aggregation)
	t.Run("test FetchUserCommitContentInLive biz", testFetchUserCommitContentInLive)
}

func testFetchUserCommitContentInLive(t *testing.T) {
	req := new(model.UserCommitListRequestInLive)
	{
		req.LastID = 0
		req.ActivityUniqID = "wish_tree"
		req.Order = "desc"
		req.Ps = 20
		req.LastID = 2
	}

	d, err := FetchUserCommitContentInLive(context.Background(), req)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func testCommitUserContentWithInvalidTimeRange(t *testing.T) {
	req := new(api.CommonActivityUserCommitReq)
	{
		req.UniqID = "fool"
		req.Content = `{"a":1}`
		req.MID = 66
	}

	if err := CommitUserContent(context.Background(), req); err != ecode.RequestErr {
		t.Error("err should as ecode.RequestErr")
	}
}

func testUserCommitContent4Aggregation(t *testing.T) {
	d, err := UserCommitContent4Aggregation(context.Background(), 66, "wish_tree")
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func testCommitUserManuScript(t *testing.T) {
	req := new(api.CommonActivityUserCommitReq)
	{
		req.UniqID = "wish_tree"
		req.Content = `{"a":1}`
		req.MID = 66
		req.BvID = "bvid6"
	}

	d, err := CommitUserManuScript(context.Background(), req)
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(d)
}

func testCommitUserContent(t *testing.T) {
	req := new(api.CommonActivityUserCommitReq)
	{
		req.UniqID = "wish_tree"
		req.Content = `{"a":2}`
		req.MID = 66
	}

	if err := CommitUserContent(context.Background(), req); err != nil {
		t.Error(err)
	}
}
