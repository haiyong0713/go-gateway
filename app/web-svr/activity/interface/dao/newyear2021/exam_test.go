package newyear2021

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/time"
)

// go test -v --count=1 exam_test.go exam.go
func TestExamBiz(t *testing.T) {
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

	t.Run("MultiUpdateExamStats test", MultiUpdateExamStatsTesting)
	t.Run("FetchExamStats test", FetchExamStatsTesting)
	t.Run("CommitUserOption test", CommitUserOptionTesting)
	t.Run("FetchUserCommitDetail test", FetchUserCommitDetailTesting)
}

func FetchUserCommitDetailTesting(t *testing.T) {
	m, err := FetchUserCommitDetail(context.Background(), 888)
	if err != nil {
		t.Error(err)

		return
	}

	if d, ok := m["1"]; !ok || d != 1 {
		t.Error("no matched commit record")
	}
}

func CommitUserOptionTesting(t *testing.T) {
	affect, err := CommitUserOption(context.Background(), 888, 1, 1)
	if err != nil {
		t.Error(err)

		return
	}

	if !affect {
		t.Log("this record has been existed")
	}
}

func FetchExamStatsTesting(t *testing.T) {
	m, err := FetchExamStats(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(m)
	t.Log(string(bs))
	if d, ok := m["1_1"]; !ok || d != 8 {
		t.Error("no 1_1 record or mot matched")

		return
	}

	if d, ok := m["1_2"]; !ok || d != 888 {
		t.Error("no 1_2 record or mot matched")

		return
	}

	if d, ok := m["2_3"]; !ok || d != 888888 {
		t.Error("no 2_3 record or mot matched")

		return
	}
}

func MultiUpdateExamStatsTesting(t *testing.T) {
	stats := new(api.ExamStatsReq)
	{
		list := make([]*api.OneExamStats, 0)
		tmp1 := new(api.OneExamStats)
		{
			tmp1.Id = 1
			tmp1.OptID = 1
			tmp1.Total = 8
		}

		tmp2 := new(api.OneExamStats)
		{
			tmp2.Id = 1
			tmp2.OptID = 2
			tmp2.Total = 888
		}

		tmp3 := new(api.OneExamStats)
		{
			tmp3.Id = 2
			tmp3.OptID = 3
			tmp3.Total = 888888
		}

		list = append(list, tmp1, tmp2, tmp3)
		stats.Stats = list
	}

	if err := MultiUpdateExamStats(context.Background(), stats); err != nil {
		t.Error(err)
	}
}
