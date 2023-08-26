package http

import (
	"encoding/json"
	"fmt"
	"git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	binding2 "go-common/library/net/http/blademaster/binding"
	aEcode "go-gateway/app/web-svr/activity/ecode"
	innerecode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/model/summer_camp"
	"go-gateway/app/web-svr/activity/interface/service"
	"go-gateway/ecode"
)

// isJoinSummerCamp 用户是否报名参与活动
func isJoinSummerCamp(ctx *bm.Context, activityId int64, mid int64) (bool, *likemdl.ActFollowingReply, error) {
	// 查询用户是否参与打卡
	follow, err := service.LikeSvc.ReserveFollowing(ctx, activityId, mid)
	if err != nil {
		log.Errorc(ctx, "summerCampUserInfo isJoinSummerCamp err(%v)", err)
		return false, nil, err
	}
	if follow.IsFollowing == false {
		return false, nil, nil
	}
	return true, follow, nil
}

// 获取用户参与活动情况
func summerCampUserInfo(ctx *bm.Context) {
	// 入参校验
	req := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}

	var res = summer_camp.UserInfoRes{}
	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	// 填充用户详情
	memberInfo, err := service.AccountSvr.MemberInfo(ctx, []int64{midI})
	if err != nil || memberInfo == nil {
		log.Errorc(ctx, "summerCampUserInfo service.account.MemberInfo err(%v)", err)
		ctx.JSON(nil, ecode.AccountInexistence)
		return
	}
	res.UserInfo = &summer_camp.UserInfo{
		NickName: memberInfo[midI].Name,
		Mid:      memberInfo[midI].Mid,
		Face:     memberInfo[midI].Face,
	}
	// 判断用户是否参与
	flag, _, err := isJoinSummerCamp(ctx, req.ActivityId, midI)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if flag == false {
		res.IsJoin = 0
		ctx.JSON(res, nil)
		return
	}
	res.IsJoin = 1
	// 获取用户任务情况
	resInfos, err := service.SummerCampSvr.GetUserTaskInfo(ctx, midI)
	if err != nil {
		ctx.JSON(nil, aEcode.SCTaskErr)
		return
	}
	res.SignDays = resInfos.SignDays
	res.TaskInfo = resInfos.TaskInfo
	ctx.JSON(res, nil)
}

// 获取课程列表
func getCourseList(ctx *bm.Context) {
	v := new(struct {
		Pn int `form:"pn" validate:"min=1" default:"1"`
		Ps int `form:"ps" validate:"min=1" default:"50"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SummerCampSvr.GetCourseList(ctx, v.Pn, v.Ps))
}

// 获取用户课程打卡情况等
func getUserCourseInfo(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
		Pn         int   `form:"pn" validate:"min=1" default:"1"`
		Ps         int   `form:"ps" validate:"min=1" default:"50"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	ctx.JSON(service.SummerCampSvr.UserCourseInfo(ctx, midI, v.Pn, v.Ps))
}

// getUserTodayVideos 获取用户当天应该展示的课程
func getUserTodayVideos(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
		CourseId   int64 `form:"course_id" validate:"required"`
		Pn         int   `form:"pn" validate:"min=1" default:"1"`
		Ps         int   `form:"ps" validate:"min=1" default:"50"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	ctx.JSON(service.SummerCampSvr.UserTodayVideosById(ctx, midI, v.CourseId, v.Pn, v.Ps))
}

// getOneDayVideos 获取用户某天应该展示的课程
func getOneDayVideos(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
		CourseId   int64 `form:"course_id" validate:"required"`
		Day        int   `form:"day" validate:"min=1" default:"1"`
		ShowTab    int   `form:"show_tab" validate:"max=1,min=0" default:"0"`
		Pn         int   `form:"pn" validate:"min=1" default:"1"`
		Ps         int   `form:"ps" validate:"min=1" default:"50"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	ctx.JSON(service.SummerCampSvr.UserOneDayVideosById(ctx, midI, v.CourseId, v.Day, v.Pn, v.Ps, v.ShowTab))
}

// userStartPlan 用户开启计划
func userStartPlan(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64   `form:"activity_id" json:"activity_id" validate:"required"`
		CourseId   []int64 `form:"course_id,split" json:"course_id,split" validate:"required"`
	})
	if err := ctx.BindWith(v, binding2.JSON); err != nil {
		return
	}
	if len(v.CourseId) <= 0 {
		ctx.JSON(nil, ecode.ParamIllegal)
		return
	}
	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	ctx.JSON(service.SummerCampSvr.StartPlan(ctx, midI, v.CourseId, v.ActivityId))
}

// 加入课程 type:1-加入 2-退出
func joinCourse(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64   `form:"activity_id" json:"activity_id" validate:"required"`
		CourseId   []int64 `form:"course_id,split" json:"course_id" validate:"required"`
		Type       int     `form:"type" json:"type" validate:"max=2,min=1" default:"1"`
	})
	if err := ctx.BindWith(v, binding2.JSON); err != nil {
		return
	}
	if len(v.CourseId) <= 0 {
		ctx.JSON(nil, ecode.ParamIllegal)
		return
	}
	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	// 判断用户是否参与
	flag, _, err := isJoinSummerCamp(ctx, v.ActivityId, midI)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if flag == false {
		ctx.JSON(nil, aEcode.ActivityNotJoin)
		return
	}
	ctx.JSON(service.SummerCampSvr.JoinCourse(ctx, midI, v.CourseId, v.Type))
}

// 积分兑换奖品
func exchangePrize(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64  `form:"activity_id" json:"activity_id" validate:"required"`
		AwardId    string `form:"award_id" json:"award_id" validate:"required"`
	})
	if err := ctx.BindWith(v, binding2.JSON); err != nil {
		return
	}

	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	// 判断用户是否参与
	flag, _, err := isJoinSummerCamp(ctx, v.ActivityId, midI)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if flag == false {
		ctx.JSON(nil, aEcode.ActivityNotJoin)
		return
	}
	// 风控
	params := risk(ctx, midI, summer_camp.RiskActionExchange)
	riskParams := summer_camp.SCRiskParams{
		Base:     params,
		Subscene: summer_camp.RiskSubsceneExchange,
	}
	var bs []byte
	bs, _ = json.Marshal(riskParams)
	riskReply, err := client.SilverbulletClient.RuleCheck(ctx, &api.RuleCheckReq{
		Scene:    summer_camp.Scene,
		EventCtx: string(bs),
		EventTs:  int64(params.EsTime),
	})
	if err != nil {
		log.Errorc(ctx, "client.SilverbulletClient.RuleCheck (%d) riskReply(%v) error(%v)", midI, riskReply, err)
	}
	// 风险命中判断
	if riskReply != nil && riskReply.Decisions != nil && len(riskReply.Decisions) != 0 {
		riskMap := riskReply.Decisions[0]
		if riskMap == riskmdl.Reject {
			log.Errorc(ctx, "Hit Risk User;Mid: (%d).riskReply:(%v) .error(%v)", midI, riskReply, err)
			ctx.JSON(nil, aEcode.RiskUser)
			return
		}
	}

	// 执行兑换
	ctx.JSON(struct{}{}, service.SummerCampSvr.ExchangeAward(ctx, midI, v.ActivityId, v.AwardId))
}

// userPointHistory 用户积分记录历史
func userPointHistory(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
		Pn         int   `form:"pn" validate:"min=1" default:"1"`
		Ps         int   `form:"ps" validate:"min=1" default:"50"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	// 判断用户是否参与
	flag, reserveInfo, err := isJoinSummerCamp(ctx, v.ActivityId, midI)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if flag == false {
		ctx.JSON(nil, aEcode.ActivityNotJoin)
		return
	}
	ctx.JSON(service.SummerCampSvr.UserPointHistory(ctx, midI, int64(reserveInfo.Ctime), v.Pn, v.Ps))
}

// exchangeAwardList 每日奖品列表
func exchangeAwardList(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
		Pn         int   `form:"pn" validate:"min=1" default:"1"`
		Ps         int   `form:"ps" validate:"min=1" default:"50"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	// 判断用户是否参与
	flag, _, err := isJoinSummerCamp(ctx, v.ActivityId, midI)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if flag == false {
		ctx.JSON(nil, ecode.ReqParamErr)
		return
	}

	ctx.JSON(service.SummerCampSvr.ExchangeAwardList(ctx, midI, v.Pn, v.Ps))
}

// summerLotteryGift 奖品列表
func summerLotteryGift(ctx *bm.Context) {
	ctx.JSON(service.JsonDataSvr.GetSummerGift(ctx), nil)
}

// summerLottery 积分兑奖
func summerLottery(ctx *bm.Context) {
	v := new(struct {
		TimeStamp int `form:"timestamp" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	orderNo := fmt.Sprintf("%d_%d", mid, v.TimeStamp)
	params := risk(ctx, midI, riskmdl.ActionLottery)
	gift, err := service.LotterySvc.DoLottery(ctx, service.SummerCampSvr.GetLotterySid(), midI, params, 1, false, orderNo)
	if err != nil && !xecode.EqualError(innerecode.ActivityNoTimes, err) {
		ctx.JSON(nil, err)
		return
	}
	if err == nil {
		ctx.JSON(gift, nil)
		return
	}
	err = service.SummerCampSvr.UseLotteryPoint(ctx, midI, orderNo, service.SummerCampSvr.GetLotterySid(), service.SummerCampSvr.GetLotteryActivityID())
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	err = service.LotterySvc.AddLotteryTimes(ctx, service.SummerCampSvr.GetLotterySid(), midI, service.SummerCampSvr.GetLotteryCid(), service.SummerCampSvr.GetLotteryActionType(), 1, orderNo, false)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(service.LotterySvc.DoLottery(ctx, service.SummerCampSvr.GetLotterySid(), midI, params, 1, false, orderNo))
}
