package http

import (
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func displayId(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	c.JSON(struct {
		ID string `json:"id"`
	}{ID: displaySvc.DisplayId(c, mid, buvid, time.Now())}, nil)
}

func wechatAuth(c *bm.Context) {
	params := c.Request.Form
	noncestr := params.Get("nonce")
	timestamp := params.Get("timestamp")
	currentURL := params.Get("url")
	if noncestr == "" || timestamp == "" || currentURL == "" {
		log.Error("arg has empty nonce(%s) timestamp(%s) cururl(%v)", noncestr, timestamp, currentURL)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(displaySvc.WechatAuth(c, noncestr, timestamp, currentURL))
}
