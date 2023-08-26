package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func coinWins(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LolSvc.UserWinCoinsV2(c, mid))
}

func coinPredictList(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LolSvc.UserPredictList(c, mid))
}

func pointList(c *bm.Context) {
	v := new(struct {
		Ts     int64 `form:"ts"`
		Ps     int64 `form:"ps" validate:"min=1"`
		LastTs int64 `form:"last_ts"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	var ts int64
	if v.LastTs > 0 {
		ts = v.LastTs
	} else if v.Ts > 0 {
		ts = v.Ts
	}
	c.JSON(service.LolSvc.PointList(c, mid, ts, v.Ps))
}
