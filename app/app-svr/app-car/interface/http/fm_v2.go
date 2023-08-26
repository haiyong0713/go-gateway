package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
)

func fmShow(c *bm.Context) {
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

	param := &fm_v2.FmShowParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	c.JSON(commonSvc.FmShow(c, param))
}

func fmShowV2(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	guestId, _ := guestIdFromCtx(c)
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}

	param := &fm_v2.ShowV2Param{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	param.GuestId = guestId
	c.JSON(commonSvc.FmShowV2(c, param))
}

func fmListRefactor(c *bm.Context) {
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

	param := &fm_v2.FmListParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	c.JSON(commonSvc.FmListRefactor(c, param))
}

func fmLike(c *bm.Context) {
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

	param := &fm_v2.FmLikeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	param.Path = c.Request.URL.Path
	param.UA = c.Request.UserAgent()
	c.JSON(nil, commonSvc.FmLike(c, param))
}

func pinPage(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	guestId, _ := guestIdFromCtx(c)
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}

	param := &fm_v2.PinPageParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	param.GuestId = guestId
	c.JSON(commonSvc.PinPage(c, param))
}
