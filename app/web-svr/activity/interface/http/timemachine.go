package http

import (
	"fmt"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
)

func timemachine2018(c *bm.Context) {
	c.JSON(nil, ecode.ActivityHasOffLine)
}

func timemachine2019(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(service.TmSvc.Timemachine(c, loginMid, v.Vmid))
}

func tmRaw2019(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(service.TmSvc.Timemachine2019Raw(c, loginMid, v.Vmid))
}

func tmCache2019(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(service.TmSvc.Timemachine2019Cache(c, loginMid, v.Vmid))
}

func tmReset2019(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(nil, service.TmSvc.Timemachine2019Reset(c, loginMid, v.Vmid))
}

func startTmProc(c *bm.Context) {
	c.JSON(nil, service.TmSvc.StartTmproc(c))
}

func stopTmProc(c *bm.Context) {
	c.JSON(nil, service.TmSvc.StopTmproc(c))
}

func userReport2020Filter(c *bm.Context) {
	v := new(struct {
		Aids  []int64 `form:"aids,split"`
		Arts  []int64 `form:"arts,split"`
		Cards []int32 `form:"cards,split"`
		Cover bool
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.TmSvc.UserReport2020Filter(c, v.Aids, v.Arts, v.Cards, v.Cover))
}

func userReport2020Cache(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.TmSvc.UserReport2020Cache(c, v.Vmid))
}

func userReport2020(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(service.TmSvc.UserReport2020(c, loginMid, v.Vmid))
}

func beforePublish2020(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(nil, service.TmSvc.BeforePublish2020(c, loginMid, v.Vmid))
}

func publish2020(c *bm.Context) {
	v := new(struct {
		AID  int64 `form:"aid"`
		Vmid int64 `form:"vmid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	info, err := service.TmSvc.Publish2020(c, loginMid, v.Vmid, v.AID)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if info.LotteryID == "" {
		c.JSON(map[string]interface{}{
			"lottery_id": info.LotteryID,
			"gift": struct {
			}{},
		}, nil)
		return
	}
	riskParams := risk(c, loginMid, riskmdl.ActionLottery)
	ret, err := service.LotterySvc.DoLottery(c, info.LotteryID, info.Mid, riskParams, 1, false, fmt.Sprint(info.Mid))
	if err != nil {
		log.Errorc(c, "lotterySvc.DoLottery err[%v]", err)
	}
	// 抽奖兜底逻辑
	if len(ret) > 0 && ret[0].GiftID == 0 {
		ret[0].GiftType = conf.Conf.Timemachine.Gift.GiftType
		ret[0].GiftName = conf.Conf.Timemachine.Gift.GiftName
		ret[0].ImgURL = conf.Conf.Timemachine.Gift.ImgURL
	}
	c.JSON(map[string]interface{}{
		"lottery_id": info.LotteryID,
		"gift":       ret,
	}, nil)
}
