package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/model"
)

var (
	seasonCardType = 13
)

func dynamicRegion(c *bm.Context) {
	var (
		rid, pn, ps int64
		err         error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	platform := params.Get("platform")
	if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, err = strconv.ParseInt(pnStr, 10, 64); err != nil || pn < 1 {
		pn = 1
	}
	if ps, err = strconv.ParseInt(psStr, 10, 64); err != nil || ps < 1 {
		ps = conf.Conf.Rule.DynamicNumArcs
	} else if ps > conf.Conf.Rule.MaxArcsPageSize {
		ps = conf.Conf.Rule.MaxArcsPageSize
	}
	isFilter, _ := strconv.ParseBool(params.Get("is_filter"))
	c.JSON(webSvc.DynamicRegion(c, rid, pn, ps, "", isFilter, platform))
}

func lpDynamicRegion(ctx *bm.Context) {
	v := &model.LpDynamicRegionReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(webSvc.DynamicRegion(ctx, v.Rid, v.Pn, v.Ps, v.Business, false, ""))
}

func dynamicRegionTag(c *bm.Context) {
	var (
		tagID, rid, pn, ps int64
		err                error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	tagIDStr := params.Get("tag_id")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if tagID, err = strconv.ParseInt(tagIDStr, 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, err = strconv.ParseInt(pnStr, 10, 64); err != nil || pn < 1 {
		pn = 1
	}
	if ps, err = strconv.ParseInt(psStr, 10, 64); err != nil || ps < 1 {
		ps = conf.Conf.Rule.DynamicNumArcs
	} else if ps > conf.Conf.Rule.MaxArcsPageSize {
		ps = conf.Conf.Rule.MaxArcsPageSize
	}
	c.JSON(webSvc.DynamicRegionTag(c, tagID, rid, pn, ps))
}

func dynamicRegionTotal(c *bm.Context) {
	c.JSON(webSvc.DynamicRegionTotal(c))
}

func dynamicRegions(c *bm.Context) {
	c.JSON(webSvc.DynamicRegions(c))
}

func dynamicEntrance(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	req := new(model.DynamicEntranceParam)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(webSvc.DynamicEntrance(c, req, mid))
}

func dynamicCardType(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.DynamicCardType(c, mid))
}

func dynamicCardCanAddContent(c *bm.Context) {
	v := new(struct {
		CardType int   `form:"card_type" validate:"required"`
		Pn       int32 `form:"pn" default:"1" validate:"min=1"`
		Ps       int32 `form:"ps" default:"10" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if v.CardType != seasonCardType {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.DynamicCardCanAddContent(c, mid, v.Pn, v.Ps))
}

func dynamicCardAdd(c *bm.Context) {
	v := new(struct {
		CardType int    `form:"card_type" validate:"required"`
		Url      string `form:"url" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if v.CardType != seasonCardType {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.DynamicCardAdd(c, mid, v.Url))
}
