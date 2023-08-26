package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/playurl"
)

func playurlWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &playurl.Param{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	referer := c.Request.Referer()
	cookie := c.Request.Header.Get(_headerCookie)
	data, msg, err := playSvc.PlayUrlWeb(c, buvid, cookie, referer, mid, param)
	if err != nil && msg != "" {
		datam := map[string]interface{}{
			"message": msg,
		}
		c.JSONMap(datam, err)
		return
	}
	c.JSON(data, err)
}

func playurlApp(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &playurl.Param{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	referer := c.Request.Referer()
	param.Mid = mid
	param.Buvid = buvid
	data, msg, err := playSvc.PlayUrl(c, buvid, referer, mid, param)
	if err != nil && msg != "" {
		datam := map[string]interface{}{
			"message": msg,
		}
		c.JSONMap(datam, err)
		return
	}
	c.JSON(data, err)
}

// eventReport 行为上报，在杜比流客户端播放、缓存播放等情况上报
func eventReport(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &common.EventReportReq{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Mid == 0 {
		param.Mid = mid
	}
	if param.Buvid == "" {
		param.Buvid = buvid
	}
	playSvc.EventReport(c, param)
	c.JSON(nil, nil)
}
