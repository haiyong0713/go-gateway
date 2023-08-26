package http

import (
	"gopkg.in/go-playground/validator.v9"

	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"go-gateway/app/web-svr/activity/interface/rewards"
)

var v *validator.Validate

func init() {
	v = validator.New()
}

func addExternalRewardsRouter(group *bm.RouterGroup) {
	rewardsGroup := group.Group("/rewards")
	{
		rewardsGroup.GET("/awards/mylist", authSvc.User, RewardsGetMyList)
		rewardsGroup.POST("/address/add", authSvc.User, RewardsAddActivityAddress)
		rewardsGroup.GET("/address/get", authSvc.User, RewardsGetActivityAddress)
		rewardsGroup.GET("/awards/mycoupon", authSvc.User, RewardsGetMyCouponList)
	}
}

func RewardsGetMyCouponList(ctx *bm.Context) {
	v := new(struct {
		CouponId   int64 `form:"coupon_id" validate:"min=1"`
		ActivityId int64 `form:"activity_id"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	res := make(map[string]interface{}, 0)
	var list []*model.CdKeyInfo
	var err error
	if v.ActivityId != 0 {
		list, err = rewards.Client.GetAwardCdKeysByActivityId(ctx, mid, v.ActivityId)
	} else {
		list, err = rewards.Client.GetAwardCdKeysById(ctx, mid, v.CouponId)
	}

	res["list"] = list
	ctx.JSON(res, err)
}

func RewardsGetMyList(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	res := make(map[string]interface{}, 0)
	list, err := rewards.Client.GetAwardRecordByMidAndActivityIdWithCache(ctx, mid, []int64{v.ActivityId}, 50)
	res["list"] = list
	ctx.JSON(res, err)
}

func RewardsGetActivityAddress(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(rewards.Client.GetActivityAddress(ctx, mid, v.ActivityId))

}

func RewardsAddActivityAddress(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"min=1"`
		AddressId  int64 `form:"address_id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON("", rewards.Client.AddActivityAddress(ctx, mid, v.ActivityId, v.AddressId))

}
