package http

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"go-gateway/app/web-svr/activity/admin/model/s10"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func ackCostInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http ackCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	idstr := params.Get("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http ackCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.AckCostInfo(ctx, id, mid))
}

func updateUserCostState(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserCostState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	idstr := params.Get("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserCostState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.UpdateUserCostState(ctx, id, mid))
}

func redeliveryCostInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http redeliveryCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	idstr := params.Get("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http redeliveryCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.RedeliveryCostInfo(ctx, id, mid))
}

func ackGiftInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http ackGiftInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	idstr := params.Get("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http ackGiftInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.AckGiftInfo(ctx, id, mid))
}

func redeliveryGiftInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http redeliveryCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	idstr := params.Get("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http redeliveryCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.RedeliveryGiftInfo(ctx, id, mid))
}

func userCostInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http userCostInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(s10Svc.UserCostInfo(ctx, mid))
}

func lotteryByRobin(ctx *bm.Context) {
	params := ctx.Request.Form
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http lotteryByRobin error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.AllLotteryUser(ctx, int32(robin)))
}

func superUserImport(ctx *bm.Context) {
	var (
		err  error
		data []byte
	)
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Error("importDetailCSV upload err(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("importDetailCSV ioutil.ReadAll err(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("importDetailCSV r.ReadAll() err(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	var users []*s10.SuperLotteryUserInfo
	for i, row := range records {
		if i == 0 {
			continue
		}
		userInfo := new(s10.SuperLotteryUserInfo)
		for field, value := range row {
			value = strings.TrimSpace(value)
			if value == "" {
				log.Warn("importDetailCSV name provinceID(%s)", value)
				ctx.JSON(nil, ecode.RequestErr)
				return
			}
			switch field {
			case 0:
				mid, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					ctx.JSON(nil, ecode.RequestErr)
					return
				}
				userInfo.Mid = mid
			case 1:
				userInfo.Name = value
			case 3:
				robin, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					ctx.JSON(nil, ecode.RequestErr)
					return
				}
				userInfo.Robin = int32(robin)
			case 4:
				gid, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					ctx.JSON(nil, ecode.RequestErr)
					return
				}
				userInfo.Gid = int32(gid)
			case 5:
				userInfo.Gname = value
			}

		}
		users = append(users, userInfo)
	}

	ctx.JSON(nil, s10Svc.SuperLotteryUser(ctx, users))
}

func checkUserLottery(ctx *bm.Context) {
	var (
		err    error
		params = new(struct {
			Robin int32   `from:"robin" validate:"min=1"`
			Mids  []int64 `from:"mids,split" validate:"min=1"`
		})
	)
	if err = ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(s10Svc.NotExistLotteryUserByRobin(ctx, params.Mids, params.Robin))
}

func userGiftInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http userGiftInfo error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(s10Svc.UserGiftInfo(ctx, mid))
}

func userGiftInfoFlush(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http userGiftInfoFlush error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.UserLotteryInfoFlush(ctx, mid))
}

func userCostCacheFlush(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http userCostCacheFlush error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.UserCostCacheFlush(ctx, mid))
}

func realGoodsList(ctx *bm.Context) {
	params := ctx.Request.Form
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http realGoodsList error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	result, err := s10Svc.RealGoodsList(ctx, int32(robin))
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: fmt.Sprintf("第%d阶段实体商品中奖名单.csv", robin),
		Header:         []string{"记录id", "用户mid", "商品名称", "用户名称", "用户手机号", "用户地址"},
		Result:         result,
	}
	exportCsv(ctx, exportCsvParam)
}

func sentOutGoods(ctx *bm.Context) {
	var (
		err    error
		params = new(struct {
			IDs []int64 `from:"ids,split" validate:"min=1"`
		})
	)
	if err = ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(nil, s10Svc.SentOutGoods(ctx, params.IDs))
}

func delGoodsStock(ctx *bm.Context) {
	params := ctx.Request.Form
	gidstr := params.Get("gid")
	gid, err := strconv.ParseInt(gidstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, s10Svc.DelGoodsStockByGid(ctx, int32(gid)))
}

func backupUsers(ctx *bm.Context) {
	var (
		err   error
		param = new(struct {
			Gid  int64   `form:"gid" validate:"gt=0"`
			Mids []int64 `from:"mids,split" validate:"min=1"`
		})
	)
	if err = ctx.Bind(param); err != nil {
		return
	}
	go s10Svc.BackupUsers(param.Gid, param.Mids)
	ctx.JSON(nil, nil)
}

func genBackUsers(ctx *bm.Context) {
	go s10Svc.GenBackupUsers()
	ctx.JSON(nil, nil)
}
