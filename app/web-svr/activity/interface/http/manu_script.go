package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/wishes_2021_spring"
	"go-gateway/app/web-svr/activity/interface/service/wishes_2021_spring"
)

func ManuScriptAggregation(ctx *bm.Context) {
	v := new(struct {
		UniqID string `form:"actid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(wishes_2021_spring.UserCommitContent4Aggregation(ctx, mid, v.UniqID))
}

func CommonActivityUserCommitContent(ctx *bm.Context) {
	v := new(struct {
		UniqID  string `form:"actid" validate:"required"`
		Content string `form:"config" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	req := new(api.CommonActivityUserCommitReq)
	{
		req.UniqID = v.UniqID
		req.MID = mid
		req.Content = v.Content
	}

	ctx.JSON(nil, wishes_2021_spring.CommitUserContent(ctx, req))
}

func ManuScriptCommit(ctx *bm.Context) {
	v := new(struct {
		UniqID  string `form:"actid" validate:"required"`
		Bvid    string `form:"bvid" validate:"required"`
		Content string `form:"postinfo" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	req := new(api.CommonActivityUserCommitReq)
	{
		req.UniqID = v.UniqID
		req.MID = mid
		req.Content = v.Content
		req.BvID = v.Bvid
	}

	ctx.JSON(wishes_2021_spring.CommitUserManuScript(ctx, req))
}

func ManuScriptListInLive(ctx *bm.Context) {
	v := new(struct {
		UniqID string `form:"actid" validate:"required"`
		LastID int64  `form:"lastid" validate:"min=0"`
		Ps     int64  `form:"ps" validate:"min=10"`
		Order  string `form:"order" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	req := new(model.UserCommitListRequestInLive)
	{
		req.LastID = v.LastID
		req.ActivityUniqID = v.UniqID
		req.Ps = v.Ps
		req.Order = v.Order
	}

	ctx.JSON(wishes_2021_spring.FetchUserCommitContentInLive(ctx, req))
}
