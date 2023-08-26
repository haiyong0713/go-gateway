package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/web-svr/activity/admin/model/bnj"
)

func pendantCheck(c *bm.Context) {
	v := new(struct {
		Mid int64 `json:"mid" form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(bnjSrv.PendantCheck(c, v.Mid))
}

func bnjValueChange(c *bm.Context) {
	v := new(struct {
		Value int64 `form:"value"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if v.Value == 0 {
		return
	}
	c.JSON(nil, bnjSrv.ValueChange(c, v.Value))
}

func bnjValue(c *bm.Context) {
	c.JSON(bnjSrv.Value(c))
}

func bnj2021ARScore2CouponRule(ctx *bm.Context) {
	v := new(model.Score2CouponRule)
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(nil, bnjSrv.UpsertScore2CouponRule(ctx, v))
}

func bnj2021ARScore2CouponRuleDel(ctx *bm.Context) {
	v := new(model.Score2CouponRule)
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(nil, bnjSrv.DelScore2CouponRule(ctx, v))
}

func bnj2021ARSetting(ctx *bm.Context) {
	v := new(struct {
		Setting string `form:"setting" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(nil, bnjSrv.AddARSetting(ctx, v.Setting))
}
