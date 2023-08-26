package newyear2021

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
)

var (
	gameService *Service
)

// go test -v --count=1 game_test.go game.go service.go exam.go time.go
func TestGameBiz(t *testing.T) {
	gameService = new(Service)

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
	if err := component.GlobalBnjDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	maxCommitTimes = 3
	score2CouponRelations = make([]*model.Score2Coupon, 0)
	{
		tmp := new(model.Score2Coupon)
		{
			tmp.Score = 50
			tmp.Coupon = 1
		}

		score2CouponRelations = append(score2CouponRelations, tmp)
	}

	updateDeviceScoreMapByFilename(genWatchedFilename(filenameOfAndroidScore))
	updateDeviceScoreMapByFilename(genWatchedFilename(filenameOfIosScore))
	updateBlackListByFilename(genWatchedFilename(filenameOfBlacklist))
	updateQuotaActivityIDByFilename(genWatchedFilename(filenameOfActivityID))
	updateBnjActivityIDByFilename(genWatchedFilename(filenameOfBnjActivityID))
	fmt.Println(QuotaActivityIDMap, BnjReserveInfo.ActivityID)
	//bs1, _ := json.Marshal(deviceScoreMap4Android)
	//bs2, _ := json.Marshal(deviceScoreMap4Ios)
	//t.Log(string(bs1), string(bs2))
	//
	//bs3, _ := json.Marshal(blackList4Android)
	//bs4, _ := json.Marshal(blackList4Ios)
	//t.Log(string(bs3), string(bs4))
	//xtime.Sleep(2*xtime.Second)
	t.Run("backup pub testing", BackupPubTesting)
	t.Run("redirect biz", GenAppRedirect4GatewayTesting)
	//t.Run("deviceScore biz", updateDeviceScoreMapTesting)
	//t.Run("test AR pre_exchange biz", preExchange)
	//t.Run("test AR quota biz", quota)
	//t.Run("test AR exchange biz", exchange)
	//t.Run("test AR score biz", score)
	//t.Run("Score2Coupon list test", score2CouponRuleListTesting)
}

func BackupPubTesting(t *testing.T) {
	err := pubExchangeIntoBackup(context.Background(), "11111")
	if err != nil {
		t.Error(err)
	}
}

func GenAppRedirect4GatewayTesting(t *testing.T) {
	req := new(api.AppJumpReq)
	{
		req.BizType = api.AppJumpBizType_Type4Bnj2021AR
		req.UserAgent = "Mozilla/5.0 BiliDroid/6.16.0 (bbcallen@gmail.com) os/android model/MI 8 mobi_app/android build/6160000 channel/master innerVer/6160010 osVer/10 network/2"
		req.Memory = 2048
	}

	reply := GenAppRedirect4Gateway(context.Background(), req)
	if reply.JumpUrl != appRedirect.ARScheme {
		t.Error("mi 8 should in AR")
	}

	req.UserAgent = "bili-universal/10355 CFNetwork/978.0.7 Darwin/18.6.0 os/ios model/iPhone XR11 mobi_app/iphone build/10355 osVer/12.3.1 network/2 channel/AppStore"
	req.Memory = 1024
	reply = GenAppRedirect4Gateway(context.Background(), req)
	if reply.JumpUrl != appRedirect.UnSupportBuildH5 {
		t.Error("iPhone XR11 should in unSupport Build")
	}

	req.UserAgent = "bili-universal/10355 CFNetwork/978.0.7 Darwin/18.6.0 os/ios model/ipad XR111 mobi_app/iphone build/6160000 osVer/12.3.1 network/2 channel/AppStore"
	req.Memory = 1024
	reply = GenAppRedirect4Gateway(context.Background(), req)
	if reply.JumpUrl != appRedirect.UnSupportAppH5 {
		t.Error("ipad XR111 should in unSupport app")
	}

	req.UserAgent = "bili-universal/10355 CFNetwork/978.0.7 Darwin/18.6.0 os/ios model/iPhone XR123 mobi_app/iphone build/6160000 osVer/7.1 network/2 channel/AppStore"
	req.Memory = 1024
	reply = GenAppRedirect4Gateway(context.Background(), req)
	if reply.JumpUrl != appRedirect.GameH5 {
		t.Error("iPhone XR123 should in game h5")
	}
}

func updateDeviceScoreMapTesting(t *testing.T) {
	level, err := calculateAdaptLevel("Mozilla/5.0 BiliDroid/6.15.0 (bbcallen@gmail.com) os/android model/HLK-AL00 mobi_app/android build/6150000 channel/master innerVer/6150000 osVer/1 network/2", 2955)
	fmt.Println(level, err)
	if err != nil {
		t.Error(err)

		return
	}

	level, err = calculateAdaptLevel("bili-universal/10355 CFNetwork/978.0.7 Darwin/18.6.0 os/ios model/iPhone XR mobi_app/iphone build/10355 osVer/12.3.1 network/2 channel/AppStore", 2955)
	fmt.Println(level, err)
	if err != nil {
		t.Error(err)

		return
	}

	if level != adaptLevelOfHigh {
		t.Error("iPhone XR should as high")
	}

	level, err = calculateAdaptLevel("bili-universal/10355 CFNetwork/1128.0.1 Darwin/19.6.0 os/ios model/iPhone 11 mobi_app/iphone build/10355 osVer/13.6 network/2 channel/AppStore", 3072)
	fmt.Println(level, err)
	if err != nil {
		t.Error(err)

		return
	}

	if level != adaptLevelOfMiddle {
		t.Error("iPhone 11 should as middle")
	}

	level, err = calculateAdaptLevel("bili-universal/10355 CFNetwork/1128.0.1 Darwin/19.6.0 os/ios model/iPhone9,2 mobi_app/iphone build/10355 osVer/13.6 network/2 channel/AppStore", 5000)
	if err != nil {
		t.Error(err)

		return
	}

	if level != adaptLevelOfMiddle {
		t.Error("iPhone 9,2 should as middle")
	}
}

func score2CouponRuleListTesting(t *testing.T) {
	if err := UpdateScore2CouponRelations(context.Background()); err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(score2CouponRelations)
	t.Log(string(bs))

	resp := calculateScore2CouponByScore(500)
	bs2, _ := json.Marshal(resp)
	t.Log(string(bs2))
}

func score(t *testing.T) {
	resp, err := gameService.ARProfile(context.Background(), 88888888)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(resp)
	t.Log(string(bs))
}

func preExchange(t *testing.T) {
	score := new(model.GameScore)
	{
		score.Score = 888
	}

	resp, err := gameService.ARPreExchange(context.Background(), 88888888, score.Score)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(resp)
	t.Log(string(bs))
}

func quota(t *testing.T) {
	resp, err := gameService.ARQuota(context.Background(), 88888888)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(resp)
	t.Log(string(bs))
}

func exchange(t *testing.T) {
	score := new(model.GameScore)
	{
		score.Score = 888
	}

	resp, err := gameService.ARPreExchange(context.Background(), 88888888, score.Score)
	if err != nil {
		t.Error(err)

		return
	}

	report := new(model.RiskManagementReportInfoOfGame)
	score.RequestID = resp.RequestID
	resp1, err1 := gameService.ARExchange(context.Background(), 88888888, score, report)
	if err1 != nil {
		t.Error(err1)

		return
	}

	bs, _ := json.Marshal(resp1)
	t.Log(string(bs))
}
