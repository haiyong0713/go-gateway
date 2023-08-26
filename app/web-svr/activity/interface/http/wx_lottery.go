package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
)

func wxLotteryDo(c *bm.Context) {
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	request := c.Request
	buvid := request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	origin := request.Header.Get("Origin")
	platformStr := request.Form.Get("platform")
	var platform int64
	switch platformStr {
	case "android":
		platform = 1
	case "ios":
		platform = 2
	default:
		platform = 0
	}
	params := risk(c, mid, riskmdl.ActionLottery)
	res, err := service.LikeSvc.WxDoLottery(c, mid, platform, buvid, request.UserAgent(), request.Referer(), origin, metadata.String(c, metadata.RemoteIP), params)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(res, nil)
}

func wxLotteryPlayWindow(c *bm.Context) {
	var loginMid int64
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	buvid := c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(service.LikeSvc.WxLotteryPlayWindow(c, loginMid, buvid))
}

func wxLotteryAward(c *bm.Context) {
	var loginMid int64
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	buvid := c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	// TODO validate buvid
	if loginMid <= 0 && buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(service.LikeSvc.WxLotteryAward(c, loginMid, buvid))
}

func wxLotteryGifts(c *bm.Context) {
	c.JSON(service.LikeSvc.WxLotteryGifts(c))
}
