package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/model/fit"
	"go-gateway/app/web-svr/activity/interface/service"
)

// isJoin 用户是否报名参与活动
func isJoin(ctx *bm.Context, activityId int64, mid int64) (bool, error) {
	// 查询用户是否参与打卡
	follow, err := service.LikeSvc.ReserveFollowing(ctx, activityId, mid)
	if err != nil {
		return false, err
	}

	if follow.IsFollowing == false {
		return false, nil
	}
	return true, nil
}

// 获取当前用户参与活动连续打卡几天
func fitUserInfo(ctx *bm.Context) {
	// 入参校验
	req := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}

	// 获取用户mid
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	// 判断用户是否参与
	flag, err := isJoin(ctx, req.ActivityId, midI)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if flag == false {
		//ctx.JSON(nil, ecode.FitActivityUserNotJoin)
		ctx.JSON(&fit.UserSignDaysRes{
			IsJoin:   0,
			SignDays: 0,
		}, nil)
		return
	}

	ctx.JSON(service.FitSvr.TaskHistoryCountProgress(ctx, midI, req.ActivityId))

}

// 获取系列计划卡片列表
func getPlanCardList(ctx *bm.Context) {
	v := new(struct {
		Pn int `form:"pn" validate:"min=1" default:"1"`
		Ps int `form:"ps" validate:"min=1" default:"40"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.FitSvr.GetPlanCardList(ctx, v.Pn, v.Ps))
}

// 获取系列计划卡片列表详情
func getPlanCardDetail(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
		PlanId     int64 `form:"plan_id" validate:"required" `
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
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.FitSvr.GetPlanCardDetail(ctx, v.PlanId, midI, v.ActivityId))
}

// 获取热门视频tags
func getHotTagsList(ctx *bm.Context) {
	ctx.JSON(service.FitSvr.GetHotTags(ctx), nil)
}

// 获取热门视频tags
func getHotVideosByTag(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64  `form:"activity_id" validate:"required"`
		Mlid       string `form:"bodan_id" validate:"required" `
		Pn         int    `form:"pn" validate:"min=1" default:"1"`
		Ps         int    `form:"ps" validate:"min=1" default:"20"`
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
	ctx.JSON(service.FitSvr.GetHotVideosByTag(ctx, v.Mlid, midI, v.ActivityId, v.Pn, v.Ps))

}
