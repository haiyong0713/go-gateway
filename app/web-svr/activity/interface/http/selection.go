package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/service"
)

func selectionInfo(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.SelectionInfo(c, mid))
}

func sensitive(c *bm.Context) {
	v := new(struct {
		Answers string `form:"answers" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.SelectionSensitive(c, v.Answers))
}

func selectionSubmit(c *bm.Context) {
	v := new(struct {
		Contests string `form:"contests" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.SelectionSubmit(c, mid, v.Contests))
}

func seleList(c *bm.Context) {
	mid, _ := c.Get("mid")
	var midI int64
	if mid != nil {
		midI = mid.(int64)
	}
	v := new(struct {
		CategoryID int64 `form:"category_id" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.SeleList(c, midI, v.CategoryID))
}

func selectionVote(c *bm.Context) {
	arg := new(like.ParamVote)
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.SelectionVote(c, mid, arg))
}

func selectionRank(c *bm.Context) {
	arg := new(like.ParamVote)
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.SelectionRank(c, mid, arg))
}

func seleAssistance(c *bm.Context) {
	arg := new(like.ParamAssistance)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.LikeSvc.SeleAssistance(c, arg))
}

func selectionOne(ctx *bm.Context) {
	go func() {
		service.LikeSvc.ExportTableOne()
	}()
	ctx.JSON("ok", nil)
}

func selectionTwo(ctx *bm.Context) {
	go func() {
		service.LikeSvc.ExportTableTwo()
	}()
	ctx.JSON("ok", nil)
}

func prReSet(ctx *bm.Context) {
	v := new(struct {
		CategoryID int64 `form:"category_id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	service.LikeSvc.ReSetCacheProductRole(ctx, v.CategoryID)
	ctx.JSON("ok", nil)
}

func prMaxVote(ctx *bm.Context) {
	v := new(struct {
		CategoryID int64 `form:"category_id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.ProductRoleMaxVote(ctx, v.CategoryID))
}

func prNotVote(ctx *bm.Context) {
	service.LikeSvc.ReSetCachePrNotVote()
	ctx.JSON("ok", nil)
}
