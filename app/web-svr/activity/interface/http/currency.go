package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func actCurrency(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(map[string]int64{
		"amount": service.LikeSvc.ActUserCurrency(c, mid, v.Sid),
	}, nil)
}

func allCurrency(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.CurCurrency(c, v.Sid, loginMid), nil)
}

func mikuList(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	mid := int64(0)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.MikuList(c, mid, v.Sid))
}

func specialList(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	mid := int64(0)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.SpecialList(c, mid, v.Sid))
}

func specialAward(c *bm.Context) {
	v := new(struct {
		Sid       int64 `form:"sid" validate:"min=1"`
		AwardType int   `form:"award_type"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(nil, service.LikeSvc.SpecialAward(c, v.Sid, mid, v.AwardType))
}

func certificateWall(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	mid := int64(0)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.CertificateWall(c, v.Sid, mid))
}
