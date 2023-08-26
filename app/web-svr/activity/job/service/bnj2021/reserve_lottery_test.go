package bnj2021

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/activity/job/component"
	bnjDao "go-gateway/app/web-svr/activity/job/dao/bnj"
	"go-gateway/app/web-svr/activity/job/model/bnj"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"
)

var (
	tmpRule  *bnj.ReserveRewardRuleFor2021
	userList []*bnj.ReservedUser
)

func init() {
	tmpRule = new(bnj.ReserveRewardRuleFor2021)
	{
		tmpRule.Count = 1000000
		tmpRule.RewardID = 8
		tmpRule.StartTime = 1607768323
		tmpRule.EndTime = 1607768323
		tmpRule.ActivityID = 888888
	}

	userList = make([]*bnj.ReservedUser, 0)
	{
		tmp1 := new(bnj.ReservedUser)
		{
			tmp1.ID = 888
			tmp1.MID = 888
		}

		tmp2 := new(bnj.ReservedUser)
		{
			tmp2.ID = 888888
			tmp2.MID = 888888
		}

		userList = append(userList, tmp1, tmp2)
	}
}

// go test -v --count=1 reserve_lottery_test.go reserve_lottery.go
func TestReserveLotteryBiz(t *testing.T) {
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

	component.GlobalCache = redis.NewRedis(redisCfg)
	component.GlobalDB = sql.NewMySQL(cfg)
	if err := component.GlobalDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	component.GlobalBnjDB = sql.NewMySQL(cfg)
	if err := component.GlobalBnjDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("reserveRewardPayBizByUserList testing", reserveRewardPayBizByUserListTesting)
	t.Run(
		"reserveRewardPayBizByUserList testing in reserve lottery",
		reserveRewardPayBizByUserListTestingInReserveLottery)
	t.Run("FetchBnjReserveRewardRuleFor2021 testing", FetchBnjReserveRewardRuleFor2021Testing)
}

func FetchBnjReserveRewardRuleFor2021Testing(t *testing.T) {
	d, err := bnjDao.FetchBnjReserveRewardRuleFor2021(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	t.Log(string(bs))
}

func reserveRewardPayBizByUserListTestingInReserveLottery(t *testing.T) {
	otherRule := tmpRule.DeepCopy()
	{
		otherRule.Count = 0
	}

	err := reserveRewardPayBizByUserList(context.Background(), userList, otherRule, "88")
	if err != nil {
		t.Error(err)
	}
}

func reserveRewardPayBizByUserListTesting(t *testing.T) {
	err := reserveRewardPayBizByUserList(context.Background(), userList, tmpRule, "88")
	if err != nil {
		t.Error(err)
	}
}
