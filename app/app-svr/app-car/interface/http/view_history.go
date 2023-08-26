package http

import (
	"go-common/library/ecode"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"

	bm "go-common/library/net/http/blademaster"
)

func viewHistory(c *bm.Context) {
	var (
		req = new(commonmdl.ViewContinueReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.ViewHistory(c, req, mid, buvid))
}

func viewHistoryTabWeb(ctx *bm.Context) {
	req := new(commonmdl.HistoryTabReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.IsWeb = true
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	if mid <= 0 && buvid == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(commonSvc.ViewHistoryTab(ctx, req, mid, buvid))
}

func viewHistoryTab(ctx *bm.Context) {
	req := new(commonmdl.HistoryTabReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	if mid <= 0 && buvid == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(commonSvc.ViewHistoryTab(ctx, req, mid, buvid))
}

func viewHistoryTabMore(ctx *bm.Context) {
	req := new(commonmdl.HistoryTabMoreReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	if mid <= 0 && buvid == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(commonSvc.ViewHistoryTabMore(ctx, req, mid, buvid))
}
