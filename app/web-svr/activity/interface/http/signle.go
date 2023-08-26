package http

import (
	"go-gateway/app/web-svr/activity/interface/service"
	"net/http"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/render"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

func grantPid(c *bm.Context) {
	c.JSON(nil, ecode.ActivityHasOffLine)
}

func imageLottery(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.ImageLottery(c, mid))
}

func doImageTask(c *bm.Context) {
	arg := new(struct {
		TaskID int64 `form:"task_id"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.ImageDoTask(c, mid, arg.TaskID))
}

func upSpecial(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.UpSpecial(c, mid))
}

func fateData(c *bm.Context) {
	c.JSON(nil, nil)
}

func receiveCoupon(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.ReceiveCoupon(c, v.Sid, mid))
}

func starState(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.StarProjectState(c, mid, v.Sid))
}

func starArc(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.StarOneArc(c, mid, v.Sid))
}

func starSpring(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.StarSpring(c, mid))
}

func starMoreArc(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.StarMoreArc(c, mid))
}

func steinList(c *bm.Context) {
	c.JSON(service.LikeSvc.SteinList(c), nil)
}

func userMatch(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	sid := service.LikeSvc.UserMatchCheck(c, mid)
	c.JSON(map[string]interface{}{"sid": sid}, nil)
}

func singleAward(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.SingleAward(c, mid, v.Sid))
}

func singleAwardState(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	state, err := service.LikeSvc.SingleAwardState(c, mid, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"award_state": state}, nil)
}

func archiveList(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Tid int64 `form:"tid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.ArchiveList(c, mid, v.Sid, v.Tid))
}

func arcLists(c *bm.Context) {
	v := new(struct {
		Sid        int64 `form:"sid" validate:"min=1"`
		DefaultTid int64 `form:"default_tid" validate:"min=1"`
		SpecialTid int64 `form:"special_tid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.ArcLists(c, v.Sid, v.DefaultTid, v.SpecialTid))
}

func pointLottery(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	data, err := service.LikeSvc.PointLottery(c, mid)
	if data == nil {
		c.JSON(nil, err)
		return
	}
	if data.Code == like.NoLotteryCode {
		data := map[string]interface{}{
			"code":    like.NoLotteryCode,
			"message": "未中奖",
		}
		c.Render(http.StatusOK, render.MapJSON(data))
	} else {
		c.JSON(data.Data, err)
	}
}

func singleWebData(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.SingleWebData(c, v.Sid, v.Lid))
}

func rcmdInfo(c *bm.Context) {
	v := new(struct {
		Mids []int64 `form:"mids,split" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, ok := c.Get("mid")
	var mid int64
	if ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.RcmdData(c, v.Mids, mid))
}

func singleGroupWebData(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.SingleGroupWebData(c, mid))
}

func resourceAudit(c *bm.Context) {
	c.JSON(nil, nil)
}

func resourceIir(c *bm.Context) {
	v := new(struct {
		MobiApp string `form:"mobi_app" validate:"required"`
		Build   int64  `form:"build" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(map[string]interface{}{"iir": service.LikeSvc.Iir(c, v.MobiApp, v.Build)}, nil)
}

func specialArcList(c *bm.Context) {
	v := new(struct {
		ID  int64 `form:"id" validate:"min=0"`
		Sid int64 `form:"sid" validate:"min=1"`
		Pn  int   `form:"pn" default:"1" validate:"min=1"`
		Ps  int   `form:"ps" validate:"min=1,max=50" default:"15"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.SpecialArcList(c, v.ID, v.Sid, v.Pn, v.Ps))
}

func readDay(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.ReadDay(c, mid))
}

func bml20Follow(c *bm.Context) {
	v := new(struct {
		Sid   int64 `form:"sid" validate:"min=1"`
		ReSrc uint8 `form:"re_src" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.Bml20Follow(c, v.Sid, mid, v.ReSrc, c.Request.Header.Get("Cookie")))
}
func imageUserRank(c *bm.Context) {
	v := new(struct {
		Type int `form:"type" default:"1" validate:"min=1,max=2"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.ImageUserRank(c, mid, v.Type))
}

func childhoodList(c *bm.Context) {
	list, err := service.LikeSvc.ChildhoodRank(c)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"list": list}, nil)
}

func stupidList(ctx *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.LikeSvc.StupidList(ctx, v.Sid, mid))
}

func stupidStatus(ctx *bm.Context) {
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	status, err := service.LikeSvc.StupidStatus(ctx, v.Sid, mid)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(map[string]interface{}{"status": status.IsAfrican}, nil)
}

func channelArcs(c *bm.Context) {
	v := new(struct {
		Sid  int64 `form:"sid" validate:"min=1"`
		Tids []int `form:"tids,split" validate:"required,min=1,max=10,dive,gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.ChannelArcs(c, v.Sid, v.Tids))
}
