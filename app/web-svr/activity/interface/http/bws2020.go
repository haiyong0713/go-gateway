package http

import (
	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/model/bws"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"
	"time"
)

func user2020(ctx *bm.Context) {
	v := new(struct {
		Key string `form:"key"`
		Bid int64  `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsSvc.User2020(ctx, v.Bid, mid, v.Key))
}

func adminUser2020(ctx *bm.Context) {
	v := new(struct {
		Key string `form:"key"`
		Bid int64  `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsSvc.AdminUser2020(ctx, v.Bid, mid, v.Key))
}

func user2020Internal(ctx *bm.Context) {
	v := new(struct {
		Key string `form:"key"`
		Mid int64  `form:"mid"`
		Bid int64  `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.BwsSvc.User2020(ctx, v.Bid, v.Mid, v.Key))
}

func awardList(ctx *bm.Context) {
	ctx.JSON(service.BwsSvc.AwardList(ctx))
}

func bwsUserPoints(ctx *bm.Context) {
	v := new(struct {
		Bid       int64  `form:"bid"`
		PointType int64  `form:"point_type"`
		Day       string `form:"day"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Day == "" {
		v.Day = time.Now().Format("20060102")
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsSvc.UserPoints(ctx, v.Bid, mid, v.PointType, v.Day))
}

func bwsUserPointsAdmin(ctx *bm.Context) {
	v := new(struct {
		Bid       int64 `form:"bid"`
		PointType int64 `form:"point_type"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsSvc.UserPointAdmin(ctx, v.Bid, mid, v.PointType))
}

func taskAward(ctx *bm.Context) {
	v := new(struct {
		TaskID int64 `form:"task_id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.AwardTask(ctx, mid, v.TaskID))
}

func lottery2020(ctx *bm.Context) {
	v := new(struct {
		BID int64 `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsSvc.Lottery2020(ctx, mid, v.BID))
}

func bwsGamePlayable2020(ctx *bm.Context) {
	v := new(struct {
		Mid    int64 `form:"mid" validate:"min=1"`
		Bid    int64 `form:"bid" validate:"min=1"`
		GameID int64 `form:"game_id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	_, _, _, err := service.BwsSvc.UserPlayable(ctx, v.Mid, v.Bid, v.GameID)
	ctx.JSON(nil, err)
}

func bwsGamePlay2020(ctx *bm.Context) {
	v := new(struct {
		Mid    int64 `form:"mid" validate:"min=1"`
		Bid    int64 `form:"bid" validate:"min=1"`
		GameID int64 `form:"game_id" validate:"min=1"`
		Star   int64 `form:"star"`
		Pass   bool  `form:"pass"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.BwsSvc.UserPlayGame(ctx, v.Mid, v.Bid, v.GameID, v.Star, v.Pass))
}

func getMidDate(ctx *bm.Context, mid int64, day string) (int64, string) {
	if service.BwsSvc.IsTest(ctx) {
		if service.BwsSvc.IsVip(ctx) {
			mid, day = service.BwsSvc.GetVipMidDate(ctx)
		} else {
			mid, day = service.BwsSvc.GetNormalMidDate(ctx)
		}
	}
	return mid, day
}

func bwsMember2020(ctx *bm.Context) {
	v := new(struct {
		Mid int64  `form:"mid" validate:"min=1"`
		Bid int64  `form:"bid" validate:"min=1"`
		Day string `form:"day"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Day == "" {
		v.Day = time.Now().Format("20060102")
	}
	var isVip bool

	if service.BwsSvc.IsWhiteMid(ctx, v.Mid) {
		_, err := service.BwsSvc.GetUserToken(ctx, v.Bid, v.Mid)
		if err != nil {
			ctx.JSON(nil, err)
			return
		}
		isVip = true
	} else {
		mid, d := getMidDate(ctx, v.Mid, v.Day)
		today, _ := strconv.ParseInt(d, 10, 64)

		res, err := service.BwsOnlineSvc.HasVipTickets(ctx, mid, today)
		if err != nil {
			ctx.JSON(nil, err)
			return
		}
		if len(res) > 0 {
			isVip = true
		}
	}
	ctx.JSON(service.BwsSvc.UserDetail(ctx, v.Mid, v.Bid, v.Day, isVip))

}

func bwsDelMemberRank2020(ctx *bm.Context) {
	v := new(struct {
		Mid int64  `form:"mid" validate:"min=1"`
		Bid int64  `form:"bid" validate:"min=1"`
		Day string `form:"day" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.BwsSvc.DelUserRank(ctx, v.Mid, v.Bid, v.Day))

}

func bwsAddMemberRank2020(ctx *bm.Context) {
	v := new(struct {
		Mid int64  `form:"mid" validate:"min=1"`
		Bid int64  `form:"bid" validate:"min=1"`
		Day string `form:"day" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.BwsSvc.AddUserRankInternal(ctx, v.Mid, v.Bid, v.Day))

}

func hasVipTicket(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Day int64 `form:"day"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.BwsOnlineSvc.HasVipTickets(ctx, v.Mid, v.Day))
}

func bws2020Member(ctx *bm.Context) {
	v := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Day string `form:"day"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Day == "" {
		v.Day = time.Now().Format("20060102")
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	var isVip bool
	if service.BwsSvc.IsWhiteMid(ctx, loginMid) {
		_, err := service.BwsSvc.GetUserToken(ctx, v.Bid, loginMid)
		if err != nil {
			ctx.JSON(nil, err)
			return
		}
		isVip = true
	} else {
		mid, d := getMidDate(ctx, loginMid, v.Day)
		day, _ := strconv.ParseInt(d, 10, 64)
		res, err := service.BwsOnlineSvc.HasVipTickets(ctx, mid, day)
		if err != nil {
			ctx.JSON(nil, err)
			return
		}
		if len(res) > 0 {
			isVip = true
		}
	}

	ctx.JSON(service.BwsSvc.UserDetail(ctx, loginMid, v.Bid, v.Day, isVip))
}

func adminBws2020Member(ctx *bm.Context) {
	v := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Day string `form:"day"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Day == "" {
		v.Day = time.Now().Format("20060102")
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)

	ctx.JSON(service.BwsSvc.AdminUserDetail(ctx, loginMid, v.Bid, v.Day))
}

func bws2020PlayGame(ctx *bm.Context) {
	v := new(struct {
		Bid    int64  `form:"bid" validate:"min=1"`
		GameID int64  `form:"game_id" validate:"min=1"`
		Key    string `form:"key" validate:"required"`
		Star   int64  `form:"star"`
		Pass   bool   `form:"pass"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.AdminAddStar(ctx, loginMid, v.Bid, v.GameID, v.Star, v.Key, v.Pass))
}

func bws2020AddHeart(ctx *bm.Context) {
	v := new(struct {
		Bid       int64  `form:"bid" validate:"min=1"`
		Heart     int64  `form:"heart" validate:"required"`
		Key       string `form:"key" validate:"required"`
		TimeStamp int64  `form:"timestamp" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.AdminAddHeart(ctx, loginMid, v.Bid, v.Heart, v.Key, v.TimeStamp))
}

func bws2020VipAddHeart(ctx *bm.Context) {
	v := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Key string `form:"vip_key" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	day := time.Now().Format("20060102")
	ctx.JSON(nil, service.BwsSvc.VipAddHeart(ctx, loginMid, v.Bid, v.Key, day))
}

func bws2020InternalAddHeart(ctx *bm.Context) {
	v := new(struct {
		Bid     int64  `form:"bid" validate:"min=1"`
		Mid     int64  `form:"mid" validate:"required"`
		Date    string `form:"day" validate:"required"`
		Heart   int64  `form:"heart" validate:"required"`
		OrderNo string `form:"orderno" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.BwsSvc.InternalAddHeart(ctx, v.Bid, v.Mid, v.Heart, v.Date, v.OrderNo))
}

func bwsUserReserve(ctx *bm.Context) {
	v := new(struct {
		Bid int64 `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	_, err := service.BwsSvc.MidToKey(ctx, v.Bid, loginMid)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(service.BwsOnlineSvc.OfflineMyReserveList(ctx, loginMid))
}

func bwsCheckReserve(ctx *bm.Context) {
	v := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Pid int64  `form:"pid" validate:"min=1"`
		Mid int64  `form:"mid"`
		Key string `form:"key"`
	})

	if err := ctx.Bind(v); err != nil {
		return
	}
	var err error
	if v.Mid == 0 && v.Key == "" {
		err = xecode.RequestErr
		ctx.JSON(nil, err)
		return
	}
	if v.Mid == 0 {
		v.Mid, err = service.BwsSvc.KeyToMid(ctx, v.Bid, v.Key)
		if err != nil {
			ctx.JSON(nil, err)
			return
		}
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	err = service.BwsSvc.IsOwner(ctx, loginMid, v.Pid, v.Bid)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(service.BwsOnlineSvc.CheckedReserve(ctx, v.Mid, v.Pid))

}

func awardSend(ctx *bm.Context) {
	v := new(struct {
		Mid     int64  `form:"mid"`
		Key     string `form:"key"`
		AwardID int64  `form:"award_id" validate:"min=1"`
		ID      int64  `form:"id" validate:"min=1"`
		BID     int64  `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Mid <= 0 && v.Key == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.AwardSend(ctx, v.BID, loginMid, v.Mid, v.ID, v.AwardID, v.Key))
}

func offlineAwardSend(ctx *bm.Context) {
	v := new(struct {
		AwardID int64 `form:"award_id" validate:"min=1"`
		ID      int64 `form:"id" validate:"min=1"`
		BID     int64 `form:"bid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	loginMid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.OfflineAwardSend(ctx, v.BID, loginMid, v.ID, v.AwardID))
}

func offlineUserRank(ctx *bm.Context) {
	v := new(struct {
		Bid int64  `form:"bid" validate:"min=1"`
		Day string `form:"day"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Day == "" {
		v.Day = time.Now().Format("20060102")
	}
	ctx.JSON(service.BwsSvc.UserRankList(ctx, v.Bid, v.Day))
}

func bws2020Store(ctx *bm.Context) {
	ctx.JSON(service.BwsSvc.GetStore(ctx))
}

func unlock2020(ctx *bm.Context) {
	v := new(bws.ParamUnlock20)
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if v.Mid <= 0 && v.Key == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.BwsSvc.Unlock2020(ctx, mid, true, v))
}

func bwsVote(ctx *bm.Context) {
	v := new(struct {
		Pid    int64 `form:"pid" validate:"min=1"`
		Ts     int64 `form:"ts" validate:"min=1"`
		Result int64 `form:"result" validate:"min=1,max=2"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.AddVote(ctx, mid, v.Pid, v.Result, v.Ts))
}

func bwsVoteClear(ctx *bm.Context) {
	v := new(struct {
		Pid    int64 `form:"pid" validate:"min=1"`
		Result int64 `form:"result" validate:"min=1,max=2"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsSvc.VoteClear(ctx, mid, v.Pid, v.Result))
}
