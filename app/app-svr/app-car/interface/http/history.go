package http

import (
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/history"
)

func historyList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &history.HisParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	plat := model.Plat(param.MobiApp, param.Device)
	data, page, err := showSvc.Cursor(c, plat, mid, buvid, param)
	// 兼容老版本依赖hasmore问题
	hasMore := false
	if len(data) > 0 {
		hasMore = true
	}
	c.JSON(struct {
		Item    []card.Handler `json:"items"`
		Page    *card.Page     `json:"page"`
		HasMore bool           `json:"has_more"`
	}{Item: data, Page: page, HasMore: hasMore}, err)
}

func historyWebList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &history.HisParam{}
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
	data, page, err := showSvc.CursorWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}

func hisReport(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &history.ReportParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	param.Timestamp = time.Now().Unix()
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	err := viewSvc.HisReport(c, mid, buvid, param)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		TimeStamp int64 `json:"timestamp"`
	}{TimeStamp: param.Timestamp}, nil)
}

func hisReportWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &history.ReportParam{}

	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	param.Timestamp = time.Now().Unix()
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}

	err := viewSvc.HisReport(c, mid, buvid, param)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		TimeStamp int64 `json:"timestamp"`
	}{TimeStamp: param.Timestamp}, nil)
}
