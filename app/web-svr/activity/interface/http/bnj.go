package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func previewInfo(c *bm.Context) {
	var loginMid int64
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.BnjSvc.PreviewInfo(c, loginMid), nil)
}

func timeline(c *bm.Context) {
	c.JSON(service.BnjSvc.Timeline(c), nil)
}

func reset(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	cd, err := service.BnjSvc.TimeReset(c, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]int64{"cd": cd}, nil)
}

func reward(c *bm.Context) {
	v := new(struct {
		Step int `form:"step" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.Reward(c, mid, v.Step))
}

func delTime(c *bm.Context) {
	v := new(struct {
		Key string `form:"key" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.BnjSvc.DelTime(c, v.Key))
}

func fail(c *bm.Context) {
	c.JSON(nil, nil)
}

func bnj20Main(c *bm.Context) {
	var loginMid int64
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.BnjSvc.Bnj20Main(c, loginMid))
}

func bnj20Reward(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.BnjSvc.Bnj20Reward(c, mid, v.ID))
}

func bnj20Material(c *bm.Context) {
	var loginMid int64
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.BnjSvc.Bnj20Material(c, loginMid), nil)
}

func bnj20MaterialUnlock(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1,oneof=8 9"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.BnjSvc.Bnj20MaterialUnlock(c, mid, v.ID))
}

func bnj20MaterialRedDot(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.BnjSvc.Bnj20MaterialRedDot(c, mid))
}

func bnj20HotpotIncrease(c *bm.Context) {
	v := new(struct {
		Count int64 `form:"count" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.BnjSvc.Bnj20HotpotIncrease(c, mid, v.Count))
}

func bnj20HotpotDecrease(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	cd, num, msg, err := service.BnjSvc.Bnj20HotpotDecrease(c, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"cd": cd, "decrease": num, "toast": msg}, nil)
}
