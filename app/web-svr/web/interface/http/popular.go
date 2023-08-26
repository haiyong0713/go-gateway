package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

func webPopular(c *bm.Context) {
	v := new(struct {
		Pn int `form:"pn" default:"1" validate:"min=1"`
		Ps int `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	buvid := reqBuvid(c)
	// get mid
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	list, noMore, err := webSvc.Popular(c, mid, buvid, c.Request.URL.Path, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"list": list, "no_more": noMore}, nil)
}

func popularSeries(c *bm.Context) {
	v := new(struct {
		Type string `form:"type" default:"weekly_selected" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, err := webSvc.PopularSeries(c, v.Type)
	if err != nil {
		log.Error("%+v", err)
		c.JSON(nil, errors.Cause(err))
		return
	}
	c.JSON(map[string]interface{}{"list": list}, nil)
}

func popularSeriesOne(c *bm.Context) {
	v := new(struct {
		Type   string `form:"type" default:"weekly_selected" validate:"required"`
		Number int64  `form:"number" default:"1" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	data, err := webSvc.PopularSeriesOne(c, v.Type, v.Number)
	if err != nil {
		log.Error("%+v", err)
		c.JSON(nil, errors.Cause(err))
		return
	}
	c.JSON(data, nil)
}

func popularPrecious(c *bm.Context) {
	c.JSON(webSvc.PopularPrecious(c))
}

func popularActivity(ctx *bm.Context) {
	midInter, _ := ctx.Get("mid")
	mid := midInter.(int64)
	ctx.JSON(webSvc.PopularActivity(ctx, mid))
}

func popularActivityArchiveList(ctx *bm.Context) {
	midInter, _ := ctx.Get("mid")
	mid := midInter.(int64)
	ctx.JSON(webSvc.PopularActivityArchiveList(ctx, mid))
}

func popularActivityAward(ctx *bm.Context) {
	v := new(struct {
		AwardName string `form:"award_name" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midInter, _ := ctx.Get("mid")
	mid := midInter.(int64)
	ctx.JSON(webSvc.PopularActivityAward(ctx, mid, v.AwardName))
}
