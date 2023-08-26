package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func redDot(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.RedDot(c, v.Mid))
}

func clearRedDot(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Mid int64 `form:"mid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
		if loginMid != 0 {
			v.Mid = loginMid
		}
	}
	if v.Mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, service.LikeSvc.ClearRetDot(c, v.Mid))
}

func articleGiant(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.ArticleGiant(c, mid))
}

func sendPoint(c *bm.Context) {
	v := new(struct {
		Mid       int64  `form:"mid"  validate:"required"`
		TimeStamp int64  `form:"timestamp"  validate:"required"`
		Source    int64  `form:"source"  validate:"required"`
		Business  string `form:"business"  validate:"required"`
		Activity  string `form:"activity"  validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.ActSvc.SupplymentActSend(c, v.Mid, v.Source, v.Activity, v.Business, v.TimeStamp))
}
