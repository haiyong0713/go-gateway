package http

import (
	bm "go-common/library/net/http/blademaster"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"
	"time"
)

const (
	_headerBuvid = "Buvid"
	_buvid       = "buvid3"
)

func doLottery(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid  string `form:"sid" validate:"required"`
		Type int    `form:"type" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}

	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.DoLottery(c, v.Sid, mid, v.Type, false))
		return
	}
	params := risk(c, mid, riskmdl.ActionLottery)
	c.JSON(service.LotterySvc.DoLottery(c, v.Sid, mid, params, v.Type, false, ""))

}

func lotteryGift(c *bm.Context) {
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LotterySvc.GiftRes(c, v.Sid))
}

func doSimpleLottery(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid        string `form:"sid" validate:"required"`
		Type       int    `form:"type" validate:"min=1"`
		AlreadyWin int    `form:"already_win"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params := risk(c, mid, riskmdl.ActionLottery)
	c.JSON(service.LotterySvc.SimpleLottery(c, v.Sid, mid, params, v.Type, v.AlreadyWin, false))

}

func addLotteryTimes(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid        string `form:"sid" validate:"required"`
		ActionType int    `form:"action_type" validate:"min=1"`
		OrderNo    string `form:"order_no"`
		Cid        int64  `form:"cid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	orderNo := v.OrderNo
	if orderNo == "" {
		orderNo = strconv.FormatInt(mid, 10) + strconv.Itoa(v.ActionType) + strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)

	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(nil, service.LikeSvc.AddLotteryTimes(c, v.Sid, mid, 0, v.ActionType, 0, orderNo, true))
		return
	}
	if err = checkEsportsArenaLimit(c, v.Sid, v.ActionType, v.Cid, mid, orderNo); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, service.LotterySvc.AddLotteryTimes(c, v.Sid, mid, v.Cid, v.ActionType, 0, orderNo, true))
}

func syncTask(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid        string `form:"sid" validate:"required"`
		Activity   string `form:"activity" validate:"required"`
		ActionType string `form:"action_type" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, service.LotterySvc.AddTimesByTask(c, v.Sid, mid, v.Activity, v.ActionType))
}

// taskInfo 任务完成情况
func taskInfo(c *bm.Context) {
	var mid int64
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}

	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LotterySvc.TaskInfo(c, mid, v.ActivityId))
}

func progressRate(c *bm.Context) {
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
	})

	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LotterySvc.ProgressRate(c, v.Sid))
}

func internalAddLotteryTimes(c *bm.Context) {
	v := new(struct {
		Sid        string `form:"sid" validate:"required"`
		Mid        int64  `form:"mid" validate:"min=1"`
		ActionType int    `form:"action_type" validate:"min=1"`
		OrderNo    string `form:"order_no"`
		Cid        int64  `form:"cid"`
		Nums       int    `form:"nums"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(nil, service.LikeSvc.AddLotteryTimes(c, v.Sid, v.Mid, v.Cid, v.ActionType, v.Nums, v.OrderNo, false))
		return
	}
	if err = checkEsportsArenaLimit(c, v.Sid, v.ActionType, v.Cid, v.Mid, v.OrderNo); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, service.LotterySvc.AddLotteryTimes(c, v.Sid, v.Mid, v.Cid, v.ActionType, v.Nums, v.OrderNo, false))

}

func checkEsportsArenaLimit(c *bm.Context, sid string, actionType int, cid int64, mid int64, orderNo string) (err error) {
	// 上B站看电竞活动, 战队订阅、赛事预约、关注，这3个行为都是收口的单日总数最高3次
	return service.EsportSvc.CheckLotteryTimeLimit(c, sid, actionType, cid, mid, orderNo)
}

func addLotteryAddress(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
		ID  int64  `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(nil, service.LikeSvc.AddLotteryAddress(c, v.Sid, v.ID, mid))
		return
	}
	c.JSON(nil, service.LotterySvc.AddLotteryAddress(c, v.Sid, v.ID, mid))

}

func lotteryAddress(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.LotteryAddress(c, v.Sid, mid))
		return
	}
	c.JSON(service.LotterySvc.LotteryAddress(c, v.Sid, mid))

}

func lotteryGetMyList(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
		Ps  int    `form:"ps" default:"15" validate:"min=1,max=50"`
		Pn  int    `form:"pn" default:"1" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.GetMyList(c, v.Sid, v.Pn, v.Ps, mid, true))
		return
	}
	c.JSON(service.LotterySvc.GetMyList(c, v.Sid, v.Pn, v.Ps, mid, true))

}

func lotteryMyWinList(c *bm.Context) {
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
		Ps  int    `form:"ps" default:"50" validate:"min=1,max=50"`
		Pn  int    `form:"pn" default:"1" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.GetMyWinList(c, v.Sid, mid, true))
		return
	}
	c.JSON(service.LotterySvc.GetMyWinList(c, v.Sid, mid, v.Pn, v.Ps, true))

}

func lotteryMyCount(c *bm.Context) {
	v := new(struct {
		Sid        string `form:"sid" validate:"required"`
		ActionType int    `form:"action_type" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(service.LotterySvc.LotteryCount(c, v.Sid, mid, v.ActionType))
}

func lotteryCanAddTimes(c *bm.Context) {
	v := new(struct {
		Sid        string `form:"sid" validate:"required"`
		ActionType int    `form:"action_type" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.GetIsCanAddTimes(c, v.Sid, 0, mid, v.ActionType, 0))
		return
	}
	c.JSON(service.LotterySvc.GetIsCanAddTimes(c, v.Sid, 0, mid, v.ActionType, 0))
}

func lotteryCouponWinList(c *bm.Context) {
	v := new(struct {
		Sid    string `form:"lottery_id" validate:"required"`
		GiftID int64  `form:"gift_id" validate:"required"`
		Ps     int    `form:"ps" default:"50" validate:"min=1,max=50"`
		Pn     int    `form:"pn" default:"1" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(service.LotterySvc.CouponWinList(c, v.Sid, mid, v.GiftID, v.Pn, v.Ps))
}

func lotteryOrderNo(c *bm.Context) {
	v := new(struct {
		ID      int64  `form:"id" validate:"required"`
		OrderNo string `form:"order_no" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LotterySvc.GetRecordByOrderNo(c, v.ID, v.OrderNo))

}

func lotteryGetUnusedTimes(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.GetUnusedTimes(c, v.Sid, mid))
		return
	}
	c.JSON(service.LotterySvc.GetUnusedTimes(c, v.Sid, mid))

}

func lotteryWinList(c *bm.Context) {
	v := new(struct {
		Sid       string `form:"sid" validate:"required"`
		Num       int64  `form:"num" default:"10"`
		NeedCache bool   `form:"need_cache" default:"true"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	newlottery, err := service.LotterySvc.InitLottery(c, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if !newlottery {
		c.JSON(service.LikeSvc.WinList(c, v.Sid, v.Num, v.NeedCache))
		return
	}
	c.JSON(service.LotterySvc.WinList(c, v.Sid, v.Num, v.NeedCache))

}

func addExtraTimes(c *bm.Context) {
	v := new(struct {
		Sid string `form:"sid" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(nil, service.LikeSvc.AddExtraTimes(c, v.Sid, mid))
}

func supplymentWin(c *bm.Context) {
	v := new(struct {
		Sid    string `form:"sid" validate:"required"`
		Mid    int64  `form:"mid" validate:"required"`
		GiftID int64  `form:"gift_id" validate:"required"`
		IP     string `form:"ip" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LotterySvc.SupplymentWin(c, v.Sid, v.Mid, v.GiftID, v.IP))
}
