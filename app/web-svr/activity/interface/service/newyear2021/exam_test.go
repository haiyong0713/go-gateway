package newyear2021

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/time"

	"go-gateway/app/web-svr/activity/interface/component"
	dao "go-gateway/app/web-svr/activity/interface/dao/newyear2021"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
)

var (
	examService *Service
)

func init111() {
	list := make([]*model.BnjExamItem, 0)
	tmp1 := new(model.BnjExamItem)
	{
		tmpOpts := make([]*model.BnjExamOption, 0)
		opt1 := new(model.BnjExamOption)
		{
			opt1.ID = 1
			opt1.Title = "李"
		}
		opt2 := new(model.BnjExamOption)
		{
			opt2.ID = 2
			opt1.Title = "王"
		}
		opt3 := new(model.BnjExamOption)
		{
			opt3.ID = 3
			opt3.Title = "赵"
		}
		opt4 := new(model.BnjExamOption)
		{
			opt4.ID = 4
			opt4.Title = "刘"
		}
		tmpOpts = append(tmpOpts, opt1, opt2, opt3, opt4)

		tmp1.ID = 1
		tmp1.StartTime = xtime.Now().Unix() - 100
		tmp1.EndTime = xtime.Now().Unix() - 50
		tmp1.Title = "李白姓什么"
		tmp1.Answer = 1
		tmp1.UserOpt = -1
		tmp1.Options = tmpOpts
	}

	tmp2 := new(model.BnjExamItem)
	{
		tmpOpts := make([]*model.BnjExamOption, 0)
		opt1 := new(model.BnjExamOption)
		{
			opt1.ID = 1
			opt1.Title = "zhongguo"
		}
		opt2 := new(model.BnjExamOption)
		{
			opt2.ID = 2
			opt1.Title = "China"
		}
		opt3 := new(model.BnjExamOption)
		{
			opt3.ID = 3
			opt3.Title = "Bhina"
		}
		opt4 := new(model.BnjExamOption)
		{
			opt4.ID = 4
			opt4.Title = "Asia"
		}
		tmpOpts = append(tmpOpts, opt1, opt2, opt3, opt4)

		tmp2.ID = 2
		tmp2.StartTime = xtime.Now().Unix() - 10
		tmp2.EndTime = xtime.Now().Unix() + 10
		tmp2.Title = "中国的英文是什么"
		tmp2.Answer = 2
		tmp2.UserOpt = -1
		tmp2.Options = tmpOpts
	}

	tmp3 := new(model.BnjExamItem)
	{
		tmpOpts := make([]*model.BnjExamOption, 0)
		opt1 := new(model.BnjExamOption)
		{
			opt1.ID = 1
			opt1.Title = "2018"
		}
		opt2 := new(model.BnjExamOption)
		{
			opt2.ID = 2
			opt1.Title = "2019"
		}
		opt3 := new(model.BnjExamOption)
		{
			opt3.ID = 3
			opt3.Title = "2020"
		}
		opt4 := new(model.BnjExamOption)
		{
			opt4.ID = 4
			opt4.Title = "2021"
		}
		tmpOpts = append(tmpOpts, opt1, opt2, opt3, opt4)

		tmp3.ID = 3
		tmp3.StartTime = xtime.Now().Unix() + 50
		tmp3.EndTime = xtime.Now().Unix() + 100
		tmp3.Title = "今天是？年"
		tmp3.Answer = 3
		tmp3.UserOpt = -1
		tmp3.Options = tmpOpts
	}

	tmp4 := new(model.BnjExamItem)
	{
		tmpOpts := make([]*model.BnjExamOption, 0)
		opt1 := new(model.BnjExamOption)
		{
			opt1.ID = 1
			opt1.Title = "2018"
		}
		opt2 := new(model.BnjExamOption)
		{
			opt2.ID = 2
			opt1.Title = "2019"
		}
		opt3 := new(model.BnjExamOption)
		{
			opt3.ID = 3
			opt3.Title = "2020"
		}
		opt4 := new(model.BnjExamOption)
		{
			opt4.ID = 4
			opt4.Title = "2021"
		}
		tmpOpts = append(tmpOpts, opt1, opt2, opt3, opt4)

		tmp4.ID = 4
		tmp4.StartTime = xtime.Now().Unix() - 10000
		tmp4.EndTime = xtime.Now().Unix() - 5000
		tmp4.Title = "今天是？年"
		tmp4.Answer = 3
		tmp4.UserOpt = -1
		tmp4.Options = tmpOpts
	}

	examBank = append(list, tmp1, tmp2, tmp3)
}

// go test -v --count=1 exam_test.go game.go exam.go service.go
func TestExamBiz(t *testing.T) {
	examService = new(Service)

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

	updateExamBankByFilename(genWatchedFilename(filenameOfExamBank))
	updateDeviceScoreMapByFilename(genWatchedFilename(filenameOfAndroidScore))
	updateDeviceScoreMapByFilename(genWatchedFilename(filenameOfIosScore))
	updateBlackListByFilename(genWatchedFilename(filenameOfBlacklist))
	updateARConfigurationByFilename(genWatchedFilename(filenameOfAR))
	batchRegisterFileWatch()
	//bs, _ := json.Marshal(examBank)
	//bs1, _ := json.Marshal(deviceScoreMap4Android)
	//bs2, _ := json.Marshal(deviceScoreMap4Ios)
	bs3, _ := json.Marshal(BnjARConfig)
	//bs4, _ := json.Marshal(blackList4Ios)
	//bs5, _ := json.Marshal(blackList4Android)
	//t.Log(string(bs))
	//t.Log(string(bs1))
	//t.Log(string(bs2))
	t.Log(string(bs3))
	//t.Log(string(bs4))
	//t.Log(string(bs5))

	xtime.Sleep(xtime.Hour)
	return

	t.Run("ExamDetail test", ExamDetailTesting)
	t.Run("CommitUserAnswer test", CommitUserAnswerTesting)
}

func CommitUserAnswerTesting(t *testing.T) {
	ctx := context.Background()
	err := examService.CommitUserAnswer(ctx, 888, 1, 1)
	if err == nil {
		t.Error("item 1 is end, can not commit")
	}

	err = examService.CommitUserAnswer(ctx, 8888, 2, 1)
	if err != nil {
		t.Errorf("item 2 can commit, err: %v", err)
	}

	err = examService.CommitUserAnswer(ctx, 8888, 3, 3)
	if err == nil {
		t.Error("item 3 can not commit")
	}
}

func ExamDetailTesting(t *testing.T) {
	ctx := context.Background()
	_, err := dao.CommitUserOption(ctx, 888, 1, 1)
	_, err = dao.CommitUserOption(ctx, 888, 2, 2)
	_, err = dao.CommitUserOption(ctx, 888, 3, 3)
	if err != nil {
		t.Error("init user option failed")

		return
	}

	resp := new(model.BnjExamResponse)
	resp, err = examService.ExamDetail(ctx, 888)
	if err != nil {
		t.Errorf("fetch exam detail failed, err: %v", err)

		return
	}

	for _, v := range resp.Bank {
		bs, _ := json.Marshal(v)

		switch v.ID {
		case 1:
			if v.Status != model.ExamStatusOfEnd || v.UserOpt != 1 {
				t.Errorf("item 1 is not expected, info: %v", string(bs))
			}
		case 2:
			if v.Status != model.ExamStatusOfDoing || v.UserOpt != 2 {
				t.Errorf("item 2 is not expected, info: %v", string(bs))
			}
		case 3:
			if v.Status != model.ExamStatusOfNotBegin || v.UserOpt != -1 {
				t.Errorf("item 3 is not expected, info: %v", string(bs))
			}
		case 4:
			if v.Status != model.ExamStatusOfEnd300Seconds || v.UserOpt != -1 {
				t.Errorf("item 4 is not expected, info: %v", string(bs))
			}
		}
	}
}
