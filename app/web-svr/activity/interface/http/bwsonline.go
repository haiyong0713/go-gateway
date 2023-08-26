package http

import (
	"context"
	"encoding/csv"
	"fmt"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
	"io"
	"os"
	"strconv"
	"strings"
)

func bwsOnlineGetStringBid(ctx *bm.Context) string {
	if ctx.Request.Form.Get("bid") == "" {
		if c, e := ctx.Request.Cookie("bid"); e == nil && c != nil {
			if c.Value != "" {
				return c.Value
			}
		}
	}
	return ctx.Request.Form.Get("bid")
}

func bwsOnlineGetBid(ctx *bm.Context) int64 {
	bid, _ := strconv.ParseInt(bwsOnlineGetStringBid(ctx), 10, 64)
	if bid == 0 {
		return service.BwsOnlineSvc.DefaultBid()
	}
	return bid
}

func bwsOnlineMain(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsOnlineSvc.Main(ctx, mid, bwsOnlineGetBid(ctx)))
}

func bwsOnlinePieceFind(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	level, num, err := service.BwsOnlineSvc.PieceFind(ctx, mid, bwsOnlineGetBid(ctx))
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(map[string]interface{}{"level": level, "num": num}, nil)
}

func bwsOnlinePieceFindFree(ctx *bm.Context) {
	v := new(struct {
		FromMid int64 `form:"from_mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	level, num, err := service.BwsOnlineSvc.PieceFreeFind(ctx, mid, v.FromMid, bwsOnlineGetBid(ctx))
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(map[string]interface{}{"level": level, "num": num}, nil)
}

func bwsOnlineAwardList(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.BwsOnlineSvc.AwardPackageList(ctx, mid, bwsOnlineGetBid(ctx)))
}

func bwsOnlineMyAwardList(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.BwsOnlineSvc.MyAwardList(ctx, mid, bwsOnlineGetBid(ctx)))
}

func bwsOnlineReward(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsOnlineSvc.AwardPackageReward(ctx, mid, v.ID, bwsOnlineGetBid(ctx)))
}

func bwsOnlineTicketReward(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsOnlineSvc.TicketReward(ctx, mid, v.ID, bwsOnlineGetBid(ctx)))
}

func bwsOnlineCurrencyFind(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsOnlineSvc.CurrencyFind(ctx, mid, bwsOnlineGetBid(ctx)))
}

func bwsOnlineMyDress(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.BwsOnlineSvc.MyDressList(ctx, mid))
}

func bwsOnlineDressUp(ctx *bm.Context) {
	v := new(struct {
		IDs []int64 `form:"ids,split" validate:"eq=3,dive,min=0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsOnlineSvc.DressUp(ctx, mid, v.IDs))
}

func bwsOnlinePrintList(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.BwsOnlineSvc.PrintList(ctx, mid, bwsOnlineGetBid(ctx)))
}

func bwsOnlinePrintDetail(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.BwsOnlineSvc.PrintDetail(ctx, mid, v.ID, bwsOnlineGetBid(ctx)))
}

func bwsOnlinePrintUnlock(ctx *bm.Context) {
	v := new(struct {
		ID     int64   `form:"id" validate:"min=1"`
		Counts []int64 `form:"counts,split" validate:"eq=3,dive,min=0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.BwsOnlineSvc.PrintUnlock(ctx, mid, v.ID, v.Counts, bwsOnlineGetBid(ctx)))
}

func bwsOnlineSendPiece(ctx *bm.Context) {
	v := new(struct {
		Mid   int64  `form:"mid" validate:"min=1"`
		ID    int64  `form:"id" validate:"min=1"`
		Token string `form:"token" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.BwsOnlineSvc.SendSpecialPiece(ctx, v.Mid, v.ID, v.Token, bwsOnlineGetBid(ctx)))
}

func bwsOnlineTabEntrance(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"uid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.BwsOnlineSvc.TabEntrance(ctx, v.Mid))
}

func bwsOnlineReserveAward(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.BwsOnlineSvc.ReserveAward(ctx, v.Mid))
}

func bwsOnlineImportOfflineHeart(ctx *bm.Context) {
	v := new(struct {
		File  string `form:"file" default:"/tmp/mid.txt"`
		Bid   int64  `form:"bid" default:"7"`
		Date  string `form:"date" default:"20201224"`
		Index int    `form:"index" default:"0"`
		Heart int64  `form:"heart" default:"1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	f, err := os.Open(v.File)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	reader := csv.NewReader(f)
	bgCtx := context.Background()
	for {
		p, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			ctx.JSON(nil, err)
			return
		}
		mid, err := strconv.ParseInt(p[v.Index], 10, 64)
		if err != nil {
			log.Errorc(ctx, "bwsOnlineImportOfflineHeart strconv.ParseInt(%s) err[%v]", p[v.Index], err)
			continue
		}
		err = service.BwsSvc.InternalAddHeart(bgCtx, v.Bid, mid, v.Heart, v.Date, fmt.Sprintf("imp_%s_%d", v.Date, mid))
		if err != nil {
			log.Errorc(ctx, "bwsOnlineImportOfflineHeart service.BwsSvc.InternalAddHeart(bgCtx, %d, %d, 1, %s, %s) err[%v]",
				bgCtx, v.Bid, mid, v.Date, fmt.Sprintf("imp_%s_%d", v.Date, mid), err)
			continue
		}
	}
	ctx.JSON(nil, nil)
}

func bwsOnlineImportOfflineAward(ctx *bm.Context) {
	v := new(struct {
		File    string `form:"file" default:"/tmp/mid.txt"`
		Index   int    `form:"index" default:"0"`
		Bid     int64  `form:"bid" default:"7"`
		AwardID int64  `form:"award_id" validate:"min=1"`
		State   string `form:"state" default:"init"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	go func() {
		f, err := os.Open(v.File)
		if err != nil {
			log.Errorc(ctx, "bwsOnlineImportOfflineAward err[%v]", err)
			return
		}
		reader := csv.NewReader(f)
		bgCtx := context.Background()
		for {
			p, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorc(ctx, "bwsOnlineImportOfflineAward err[%v]", err)
				return
			}
			mid, err := strconv.ParseInt(p[v.Index], 10, 64)
			if err != nil {
				log.Errorc(ctx, "bwsOnlineImportOfflineAward strconv.ParseInt(%s) err[%v]", p[v.Index], err)
				continue
			}
			token, err := service.BwsSvc.GetUserToken(bgCtx, v.Bid, mid)
			if err != nil || token == "" {
				log.Errorc(ctx, "bwsOnlineImportOfflineAward service.BwsSvc.GetUserToken(ctx, %d, %d) err[%v]", v.Bid, mid, err)
				continue
			}
			_, err = service.BwsSvc.AddUserAward(bgCtx, token, v.AwardID, v.State)
			if err != nil {
				log.Errorc(ctx, "bwsOnlineImportOfflineAward service.BwsSvc.AddUserAward(ctx, %s, %d, %s) err[%v]",
					ctx, token, v.AwardID, v.State, err)
				continue
			}
		}
	}()
	ctx.JSON(nil, nil)
}

func bwsOnlineTicketBind(ctx *bm.Context) {
	v := new(struct {
		UserName   string `json:"user_name" form:"user_name" validate:"required"`
		IdType     int    `json:"id_type" form:"id_type" `
		PersonalId string `json:"personal_id" form:"personal_id" validate:"min=1"`
		TicketNo   string `json:"ticket_no" form:"ticket_no" validate:"required,min=4"`
		Bid        int64  `form:"bid" default:"8"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	var (
		ef  int64
		err error
	)
	ef, err = service.BwsOnlineSvc.BindTicketReserve(ctx, mid, v.UserName, v.PersonalId, v.IdType, v.TicketNo)
	if err == nil && ef > 0 {
		for i := 0; i < 3; i++ {
			token, err := service.BwsSvc.GetUserToken(ctx, v.Bid, mid)
			log.Infoc(ctx, "bwsOnlineTicketBind GetUserToken , loop:%v , Bid:%d, mid:%d , err[%+v]", i, v.Bid, mid, err)
			if err == nil && token != "" {
				break
			}
		}
	}
	ctx.JSON(ef, err)
}

func bwsOnlineReserveInfo(ctx *bm.Context) {
	v := new(struct {
		ReserveDate string `json:"reserve_date" form:"reserve_date"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var screenDates []int64
	if v.ReserveDate != "" {
		checkRepeat := make(map[int64]struct{})
		for _, v := range strings.Split(v.ReserveDate, ",") {
			if date, err := strconv.ParseInt(v, 10, 64); err == nil && date > 0 {
				if _, ok := checkRepeat[date]; !ok {
					screenDates = append(screenDates, date)
					checkRepeat[date] = struct{}{}
				}
			}
		}
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	ctx.JSON(service.BwsOnlineSvc.InterReserveList(ctx, mid, screenDates))
}

func bwsOnlineReserveDo(ctx *bm.Context) {
	v := new(struct {
		TicketNo       string `json:"ticket_no" form:"ticket_no" validate:"min=1"`
		InterReserveId int64  `json:"inter_reserve_id" form:"inter_reserve_id" validate:"min=1"`
	})

	if err := ctx.Bind(v); err != nil {
		return
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)

	ctx.JSON(service.BwsOnlineSvc.ReserveDo(ctx, mid, v.InterReserveId, v.TicketNo))
}

func bwsOnlineReservedList(ctx *bm.Context) {
	if err := ctx.Bind(v); err != nil {
		return
	}

	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	myReserve, err := service.BwsOnlineSvc.MyReservedList(ctx, mid)
	ctx.JSON(map[string]interface{}{
		"reserve_list": myReserve,
	}, err)
}
