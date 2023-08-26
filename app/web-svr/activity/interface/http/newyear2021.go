package http

import (
	"encoding/json"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"
	"time"

	"go-common/library/net/metadata"

	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service/newyear2021"

	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"

	bm "go-common/library/net/http/blademaster"
)

func Bnj2021LiveLotteryDetail(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}

	ctx.JSON(service.NewYear2021Svc.LiveLotteryDetail(ctx, mid))
}

func Bnj2021LotteryUserHistory(ctx *bm.Context) {
	v := new(struct {
		SceneID int64 `form:"scene_id" validate:"min=1,max=3,required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data := make(map[string]interface{}, 0)
	list, quota, err := service.NewYear2021Svc.GetAwardRecordByMid(ctx, mid, v.SceneID)
	data["reward"] = list
	data["quota"] = quota
	ctx.JSON(data, err)
}

func Bnj2021PublicizeRewardPay(ctx *bm.Context) {
	v := new(struct {
		OpType int64 `form:"op_type" validate:"min=1,max=2,required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	var lotteryCount int64
	switch v.OpType {
	case 1: //单抽
		lotteryCount = 1
	case 2: //十连抽
		lotteryCount = 10
	default:
	}
	newyear2021.PubUserDrawLog(ctx, mid, v.OpType)
	data := make(map[string]interface{}, 0)
	params := risk(ctx, mid, riskmdl.ActionLottery)
	list, err := service.NewYear2021Svc.DoLottery(ctx, mid, 1, lotteryCount, service.LotterySvc, params, int64(80) /*风控活动ID*/, true, false, true, true)
	data["reward"] = list
	if err != nil {
		ctx.JSON(data, err)
		return
	}
	userCoupon, err := service.NewYear2021Svc.FetchUserCoupon(ctx, mid, 1)
	if err != nil {
		ctx.JSON(data, err)
		return
	}
	data["quota"] = userCoupon
	ctx.JSON(data, err)
}

// 直播间抽奖
func Bnj2021RewardPay(ctx *bm.Context) {
	v := new(struct {
		SceneID int64 `form:"scene_id" validate:"min=2,max=3,required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	data := make(map[string]interface{}, 0)
	list, err := service.NewYear2021Svc.LiveLotteryExchange(ctx, mid, v.SceneID)

	data["reward"] = list
	if err != nil {
		ctx.JSON(data, err)
		return
	}
	userCoupon, err := service.NewYear2021Svc.FetchUserCoupon(ctx, mid, v.SceneID)
	if err != nil {
		ctx.JSON(data, err)
		return
	}
	data["quota"] = userCoupon
	ctx.JSON(data, err)
}

func LiveDrawReIssue(ctx *bm.Context) {
	v := new(struct {
		SceneID int64 `form:"scene_id" validate:"min=2,max=3,required"`
		MID     int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(service.NewYear2021Svc.LiveDrawReIssue(ctx, v.MID, v.SceneID))
}

func ARConfiguration(ctx *bm.Context) {
	v := new(struct {
		OpType int64 `form:"op_type" validate:"min=1,max=5"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(service.NewYear2021Svc.ARConfiguration(v.OpType), nil)
}

func Bnj2021Profile(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	ctx.JSON(service.NewYear2021Svc.ARProfile(ctx, mid))
}

func ARAdaptLevel(ctx *bm.Context) {
	v := new(struct {
		Memory int64 `form:"memory" validate:"min=0,max=32768"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ua := ctx.Request.Header.Get("User-Agent")
	ctx.JSON(service.NewYear2021Svc.ARAdaptLevel(ctx, ua, v.Memory))
}

func BnjReserveStatus(ctx *bm.Context) {
	var mid int64
	midStr, _ := ctx.Get("mid")
	if midStr != nil {
		mid = midStr.(int64)
	}

	ctx.JSON(service.NewYear2021Svc.ReserveStatus(ctx, mid))
}

func BnjLiveExamCommit(ctx *bm.Context) {
	v := new(struct {
		QID    int64  `form:"q_id" validate:"min=1,max=10"`
		OptID  int64  `form:"o_id" validate:"min=1,max=4"`
		AppKey string `form:"appkey"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	var mid int64
	midStr, _ := ctx.Get("mid")
	if midStr != nil {
		mid = midStr.(int64)
	}

	request := ctx.Request
	report := new(model.RiskManagementReportInfoOfExam)
	{
		report.MID = mid
		report.OrderID = v.QID
		report.UserAnswer = v.OptID
		report.TopicTime = time.Now().Unix()
		report.UserAgent = request.UserAgent()
		report.Referer = request.Referer()
		report.IP = metadata.String(ctx, metadata.RemoteIP)
		buvid := request.Header.Get(_headerBuvid)
		if buvid == "" {
			cookie, _ := request.Cookie(_headerBuvid)
			if cookie != nil {
				buvid = cookie.Value
			} else {
				cookie, _ = request.Cookie(_buvid)
				if cookie != nil {
					buvid = cookie.Value
				}
			}
		}
		report.Buvid = buvid
		report.Origin = request.Header.Get("Origin")
		if ua := request.Header.Get("User-Agent"); ua != "" {
			if d, err := model.ParseUserAgent2UserAppInfo(ua); err == nil {
				report.Build = strconv.FormatInt(d.Build, 10)
				report.Platform = d.Os
			}
		}
	}

	ctx.JSON(nil, service.NewYear2021Svc.CommitUserAnswer(ctx, mid, report))
}

func BnjLiveExamDetail(ctx *bm.Context) {
	var mid int64
	midStr, _ := ctx.Get("mid")
	if midStr != nil {
		mid = midStr.(int64)
	}

	ctx.JSON(service.NewYear2021Svc.ExamDetail(ctx, mid))
}

func BnjLiveStatus(ctx *bm.Context) {
	ctx.JSON(service.NewYear2021Svc.LiveStatus(ctx))
}

func Bnj2021PublicizeBiz(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}

	ctx.JSON(service.NewYear2021Svc.PublicizeAggregation(ctx, mid))
}

func BnjARConfig(ctx *bm.Context) {
	ctx.JSON(service.NewYear2021Svc.ARSetting(ctx))
}

func Bnj2021ARQuota(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	ctx.JSON(service.NewYear2021Svc.ARQuota(ctx, mid))
}

func Bnj2021ARConfirm(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	ctx.JSON(service.NewYear2021Svc.ARConfirm(ctx, mid))
}

func Bnj2021PreExchange(ctx *bm.Context) {
	v := new(struct {
		Score int64 `form:"score" validate:"min=0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	ctx.JSON(service.NewYear2021Svc.ARPreExchange(ctx, mid, v.Score))
}

func Bnj2021Exchange(ctx *bm.Context) {
	v := new(model.GameScore)
	if err := ctx.Bind(v); err != nil {
		return
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	request := ctx.Request
	report := new(model.RiskManagementReportInfoOfGame)
	{
		report.MID = mid
		report.UserAgent = request.UserAgent()
		report.Referer = request.Referer()
		report.IP = metadata.String(ctx, metadata.RemoteIP)
		buvid := request.Header.Get(_headerBuvid)
		if buvid == "" {
			cookie, _ := request.Cookie(_headerBuvid)
			if cookie != nil {
				buvid = cookie.Value
			} else {
				cookie, _ = request.Cookie(_buvid)
				if cookie != nil {
					buvid = cookie.Value
				}
			}
		}
		report.Buvid = buvid
		report.Origin = request.Header.Get("Origin")
		report.GameType = v.GameType
		if ua := request.Header.Get("User-Agent"); ua != "" {
			if d, err := model.ParseUserAgent2UserAppInfo(ua); err == nil {
				report.Build = strconv.FormatInt(d.Build, 10)
				report.Platform = d.Os
			}
		}
	}

	ctx.JSON(service.NewYear2021Svc.ARExchange(ctx, mid, v, report))
}

func Bnj2021TaskStatus(ctx *bm.Context) {
	v := new(struct {
		Type int64 `form:"type" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	switch v.Type {
	case int64(1):
		ctx.JSON(service.NewYear2021Svc.GetDailyTaskStatus(ctx, mid))
	case int64(3):
		ctx.JSON(service.NewYear2021Svc.GetLevelTaskStatus(ctx, mid))
	default:
		ctx.JSON("", ecode.ActivityTaskNotExist)
	}
}

func Bnj2021ReceiveReward(ctx *bm.Context) {
	v := new(struct {
		Type   int64 `form:"type" validate:"min=1"`
		TaskId int64 `form:"task_id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	switch v.Type {
	case int64(1):
		ctx.JSON(service.NewYear2021Svc.ReceivePersonalDailyReward(ctx, mid, v.TaskId, false))
	case int64(3):
		ctx.JSON(service.NewYear2021Svc.ReceivePersonalLevelReward(ctx, mid, v.TaskId, false))
	default:
		ctx.JSON("", ecode.ActivityTaskNotExist)
	}
}

func Bnj2021GetConfig(ctx *bm.Context) {
	version, config, err := service.NewYear2021Svc.GetConfFromDB(ctx)
	if err != nil {
		ctx.JSON("", err)
		return
	}
	res := map[string]interface{}{}
	res["config"] = config
	res["version"] = version
	ctx.JSONMap(res, nil)
}

func Bnj2021UpdateConfig(ctx *bm.Context) {
	v := new(struct {
		Config string `form:"config" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	config := &model.Config{}
	if err := json.Unmarshal([]byte(v.Config), &config); err != nil {
		ctx.JSON("", err)
		return
	}
	ctx.JSON("", service.NewYear2021Svc.UpdateConfInDB(ctx, config))
}

func Bnj2021DeleteConfig(ctx *bm.Context) {
	v := new(struct {
		Version int64 `form:"version" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON("", service.NewYear2021Svc.DeleteConfInDB(ctx, v.Version))
}

func Bnj2021Support(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.NewYear2021Svc.Support(ctx, v.Mid))
}

func Bnj2021PubMallVisit(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON("", service.NewYear2021Svc.PubMallVisit(ctx, mid))
}
