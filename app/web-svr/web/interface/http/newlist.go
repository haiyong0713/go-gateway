package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/model"
)

func newList(c *bm.Context) {
	var (
		pn, ps  int
		rid, tp int64
		err     error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	if ridStr != "" {
		if rid, err = strconv.ParseInt(ridStr, 10, 32); err != nil || rid < 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil || ps < 1 || ps > int(conf.Conf.Rule.MaxArcsPageSize) {
		ps = int(conf.Conf.Rule.MaxArcsPageSize)
	}
	tpStr := params.Get("type")
	if tpStr != "" {
		if tp, err = strconv.ParseInt(tpStr, 10, 8); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	rs, count, err := webSvc.NewList(c, int32(rid), int8(tp), pn, ps)
	if err != nil {
		c.JSON(nil, err)
		log.Error("webSvc.Newlist(%d,%d,%d) error(%v)", rid, pn, ps, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   pn,
		"size":  ps,
		"count": count,
	}
	data["page"] = page
	var afRs []*model.BvArc
	for _, arc := range rs {
		if arc != nil && arc.IsNormal() {
			afRs = append(afRs, arc)
		}
	}
	if afRs == nil {
		afRs = make([]*model.BvArc, 0)
	}
	data["archives"] = afRs
	c.JSON(data, nil)
}

func lpNewList(ctx *bm.Context) {
	v := &model.LpNewlistReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(webSvc.LpNewlist(ctx, mid, v))
}

func information(c *bm.Context) {
	params := c.Request.Form
	rid, _ := strconv.ParseInt(params.Get("rid"), 10, 32)
	if rid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	pn, _ := strconv.Atoi(params.Get("pn"))
	if pn < 1 {
		pn = 1
	}
	ps, _ := strconv.Atoi(params.Get("ps"))
	if ps < 1 || ps > int(conf.Conf.Rule.MaxArcsPageSize) {
		ps = int(conf.Conf.Rule.MaxArcsPageSize)
	}
	var tp int64
	tpStr := params.Get("type")
	if tpStr != "" {
		var err error
		if tp, err = strconv.ParseInt(tpStr, 10, 8); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.Information(c, int32(rid), int8(tp), pn, ps, mid))
}
