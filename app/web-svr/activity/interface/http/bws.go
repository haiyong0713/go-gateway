package http

import (
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/model/bws"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

func user(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Mid int64  `form:"mid"`
		Key string `form:"key"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	if v.Mid == 0 {
		v.Mid = loginMid
	}
	if v.Mid == 0 && v.Key == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, ecode.AccessDenied)
	// if service.BwsSvc.IsNewBws(c, v.Bid) {
	// 	c.JSON(service.BwsSvc.NewUser(c, v.Bid, v.Mid, v.Key))
	// } else {
	// 	c.JSON(service.BwsSvc.User(c, v.Bid, v.Mid, v.Key))
	// }
}

func points(c *bm.Context) {
	p := new(bws.ParamPoints)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(service.BwsSvc.Points(c, p))
}

func point(c *bm.Context) {
	p := new(bws.ParamID)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(service.BwsSvc.Point(c, p))
}

func achievements(c *bm.Context) {
	p := new(bws.ParamID)
	if err := c.Bind(p); err != nil {
		return
	}
	if p.Day != "" {
		var (
			day int64
			err error
		)
		if day, err = strconv.ParseInt(p.Day, 10, 64); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		if day < 20180719 || day > 20180722 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	c.JSON(service.BwsSvc.Achievements(c, p))
}

func achievement(c *bm.Context) {
	p := new(bws.ParamID)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(service.BwsSvc.Achievement(c, p))
}

func unlock(c *bm.Context) {
	c.JSON(nil, ecode.AccessDenied)
	// v := new(bws.ParamUnlock)
	// if err := c.Bind(v); err != nil {
	// 	return
	// }
	// midStr, _ := c.Get("mid")
	// mid := midStr.(int64)
	// if v.Mid == 0 && v.Key == "" {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// if service.BwsSvc.IsNewBws(c, v.Bid) {
	// 	achieves, err := service.BwsSvc.NewUnlock(c, mid, v)
	// 	if err != nil {
	// 		c.JSON(nil, err)
	// 	} else {
	// 		c.JSON(map[string]interface{}{"achievements": achieves}, nil)
	// 	}
	// } else {
	// 	c.JSON(nil, service.BwsSvc.Unlock(c, mid, v))
	// }
}

func binding(c *bm.Context) {
	p := new(bws.ParamBinding)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	achieves, err := service.BwsSvc.Binding(c, mid, p)
	if err != nil {
		c.JSON(nil, err)
	} else {
		c.JSON(map[string]interface{}{"achievements": achieves}, nil)
	}
}

func award(c *bm.Context) {
	p := new(bws.ParamAward)
	if err := c.Bind(p); err != nil {
		return
	}
	if p.Mid == 0 && p.Key == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.BwsSvc.Award(c, mid, p))
}

func lottery(c *bm.Context) {
	p := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Aid int64  `form:"aid" validate:"min=1"`
		Day string `form:"day"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if p.Day != "" {
		var (
			day int64
			err error
		)
		if day, err = strconv.ParseInt(p.Day, 10, 64); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		if day < 20180719 || day > 20180722 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.BwsSvc.Lottery(c, p.Bid, mid, p.Aid, p.Day))
}

func lotteryV1(c *bm.Context) {
	v := new(struct {
		Bid     int64 `form:"bid" validate:"min=1"`
		AwardID int64 `form:"award_id" validate:"min=1"`
		Type    int64 `form:"type"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.BwsSvc.LotteryV1(c, v.Bid, v.AwardID, v.Type))
}

func redisInfo(c *bm.Context) {
	v := new(struct {
		Bid      int64   `form:"bid"`
		Mid      int64   `form:"mid"`
		Key      string  `form:"key"`
		Type     string  `form:"type" validate:"required"`
		Day      string  `form:"day"`
		Del      int     `form:"del"`
		LockType int64   `form:"lock_type"`
		Pids     []int64 `form:"pids,split"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(service.BwsSvc.RedisInfo(c, v.Bid, loginMid, v.Mid, v.Key, v.Day, v.Type, v.Del, v.LockType, v.Pids))
}

func keyInfo(c *bm.Context) {
	v := new(struct {
		Bid  int64  `form:"bid"`
		ID   int64  `form:"id"`
		Mid  int64  `form:"mid"`
		Key  string `form:"key"`
		Type string `form:"type" validate:"required"`
		Del  int    `form:"del"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	loginMid := midStr.(int64)
	c.JSON(service.BwsSvc.KeyInfo(c, v.Bid, loginMid, v.ID, v.Mid, v.Key, v.Type, v.Del))
}

func lotteryCheck(c *bm.Context) {
	v := new(struct {
		Aid int64  `form:"aid" validate:"min=1"`
		Day string `form:"day"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.BwsSvc.LotteryCheck(c, mid, v.Aid, v.Day))
}

func adminInfo(c *bm.Context) {
	v := new(struct {
		Bid int64 `form:"bid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.BwsSvc.AdminInfo(c, v.Bid, mid))
}

func rechargeAward(c *bm.Context) {
	v := new(bws.ParamRechargeAward)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.BwsSvc.RechargeAward(c, v))
}

func achieveRank(c *bm.Context) {
	var loginMid int64
	args := new(struct {
		Bid  int64 `form:"bid" validate:"min=1"`
		Type int   `form:"type" default:"0"`
		Ps   int   `form:"ps" default:"22" validate:"max=50"`
	})
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(service.BwsSvc.AchieveRank(c, args.Bid, loginMid, args.Ps, args.Type))
}

func fields(c *bm.Context) {
	args := new(struct {
		Bid int64 `form:"bid" validate:"min=1"`
	})
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(service.BwsSvc.Fields(c, args.Bid))
}

func addAchieve(c *bm.Context) {
	v := new(struct {
		Bid int64 `form:"bid" validate:"min=1"`
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.BwsSvc.AddAchieve(c, v.Bid, v.Mid))
}

func gradeEnter(c *bm.Context) {
	v := new(struct {
		Bid    int64  `form:"bid" validate:"min=1"`
		Pid    int64  `form:"pid" validate:"min=1"`
		Amount int64  `form:"amount" validate:"min=1,max=100"`
		Key    string `form:"key"`
		Mid    int64  `form:"mid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.BwsSvc.GradeEnter(c, v.Bid, v.Pid, mid, v.Mid, v.Amount, v.Key))
}

func gradeShow(c *bm.Context) {
	v := new(struct {
		Bid int64 `form:"bid" validate:"min=1"`
		Pid int64 `form:"pid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	loginMid := int64(0)
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.BwsSvc.GradeShow(c, v.Bid, v.Pid, loginMid))
}

func gradeFix(c *bm.Context) {
	v := new(struct {
		Bid int64 `form:"bid" validate:"min=1"`
		Pid int64 `form:"pid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.BwsSvc.GradeFix(c, v.Bid, v.Pid, mid))
}

func catchUp(c *bm.Context) {
	param := &bwsmdl.CatchUpper{}
	if err := c.Bind(param); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data, err := service.BwsSvc.InCatchUp(c, mid, param)
	c.JSON(struct {
		List []*bwsmdl.BluetoothUpInfo `json:"list"`
	}{List: data}, err)
}

func catchList(c *bm.Context) {
	param := &bwsmdl.CatchUpper{}
	if err := c.Bind(param); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data, count, err := service.BwsSvc.CatchList(c, mid, param)
	c.JSON(struct {
		List  []*bwsmdl.BluetoothUpInfo `json:"list"`
		Count int                       `json:"count"`
	}{List: data, Count: count}, err)
}

func catchBluetooth(c *bm.Context) {
	param := &bwsmdl.CatchUpper{}
	if err := c.Bind(param); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data, err := service.BwsSvc.CatchBluetoothList(c, mid, param)
	c.JSON(struct {
		List []*bwsmdl.BluetoothUpInfo `json:"list"`
	}{List: data}, err)
}

func bluetoothUps(c *bm.Context) {
	param := &bwsmdl.CatchUpper{}
	if err := c.Bind(param); err != nil {
		return
	}
	data := service.BwsSvc.BluetoothUpsAll(c, param)
	c.JSON(struct {
		List []*bwsmdl.BluetoothUpInfo `json:"list"`
	}{List: data}, nil)
}

func createUserToken(ctx *bm.Context) {
	v := new(struct {
		Pid int64 `form:"pid" validate:"min=1"`
		BID int64 `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var loginMid int64
	if midInter, ok := ctx.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	token, err := service.BwsSvc.CreateUserToken(ctx, loginMid, v.Pid, v.BID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(map[string]interface{}{"token": token}, nil)
}
