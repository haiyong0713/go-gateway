package lottery

import (
	"context"
	xtime "go-common/library/time"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLotteryListTotal(t *testing.T) {
	Convey("ListTotal", t, func() {
		var (
			c       = context.Background()
			state   = int(1)
			keyword = ""
		)
		Convey("When everything goes positive", func() {
			total, err := d.ListTotal(c, state, keyword)
			Convey("Then err should be nil.total should not be nil.", func() {
				So(err, ShouldBeNil)
				So(total, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryBaseList(t *testing.T) {
	Convey("BaseList", t, func() {
		var (
			c       = context.Background()
			pn      = int(0)
			ps      = int(0)
			state   = int(0)
			keyword = ""
			rank    = ""
		)
		Convey("When everything goes positive", func() {
			list, err := d.BaseList(c, pn, ps, state, keyword, rank)
			Convey("Then err should be nil.list should not be nil.", func() {
				So(err, ShouldBeNil)
				So(list, ShouldBeNil)
			})
		})
	})
}

func TestLotteryLotDetailByID(t *testing.T) {
	Convey("LotDetailByID", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			detail, err := d.LotDetailByID(c, id)
			Convey("Then err should be nil.detail should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(detail, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryUpdateLotInfo(t *testing.T) {
	Convey("UpdateLotInfo", t, func() {
		var (
			c        = context.Background()
			tx, _    = d.BeginTran(c)
			id       = int64(0)
			name     = ""
			operator = ""
			stime    xtime.Time
			etime    xtime.Time
		)
		Convey("When everything goes positive", func() {
			err := d.UpdateLotInfo(tx, c, id, name, operator, stime, etime)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryLotDetailBySID(t *testing.T) {
	Convey("LotDetailBySID", t, func() {
		var (
			c   = context.Background()
			sid = ""
		)
		Convey("When everything goes positive", func() {
			detail, err := d.LotDetailBySID(c, sid)
			Convey("Then err should be nil.detail should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(detail, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryInitLotDetail(t *testing.T) {
	Convey("InitLotDetail", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			lotID = ""
		)
		Convey("When everything goes positive", func() {
			err := d.InitLotDetail(tx, c, lotID)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryCreate(t *testing.T) {
	Convey("Create", t, func() {
		var (
			c        = context.Background()
			tx, _    = d.BeginTran(c)
			name     = ""
			operator = ""
			stime    xtime.Time
			etime    xtime.Time
			lotType  = int(0)
		)
		Convey("When everything goes positive", func() {
			id, err := d.Create(tx, name, operator, stime, etime, lotType)
			Convey("Then err should be nil.id should not be nil.", func() {
				So(err, ShouldBeNil)
				So(id, ShouldNotBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotterycreateAction(t *testing.T) {
	Convey("createAction", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			id    = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.createAction(tx, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotterycreateAddTimes(t *testing.T) {
	Convey("createAddTimes", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			id    = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.createAddTimes(tx, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotterycreateAddress(t *testing.T) {
	Convey("createAddress", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			id    = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.createAddress(tx, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotterycreateWin(t *testing.T) {
	Convey("createWin", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			id    = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.createWin(tx, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryDelete(t *testing.T) {
	Convey("Delete", t, func() {
		var (
			c        = context.Background()
			tx, _    = d.BeginTran(c)
			id       = int64(0)
			operator = ""
		)
		Convey("When everything goes positive", func() {
			err := d.Delete(tx, c, id, operator)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryGetLotRuleBySID(t *testing.T) {
	Convey("GetLotRuleBySID", t, func() {
		var (
			c   = context.Background()
			sid = ""
		)
		Convey("When everything goes positive", func() {
			result, err := d.GetLotRuleBySID(c, sid)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAllTimesConf(t *testing.T) {
	Convey("AllTimesConf", t, func() {
		var (
			c   = context.Background()
			sid = ""
		)
		Convey("When everything goes positive", func() {
			result, err := d.AllTimesConf(c, sid)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAllGift(t *testing.T) {
	Convey("AllGift", t, func() {
		var (
			c   = context.Background()
			sid = ""
		)
		Convey("When everything goes positive", func() {
			result, err := d.AllGift(c, sid)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRuleUpdate(t *testing.T) {
	Convey("RuleUpdate", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			rule  = &lotmdl.RuleInfo{}
		)
		Convey("When everything goes positive", func() {
			r, err := d.RuleUpdate(tx, c, rule)
			Convey("Then err should be nil.r should not be nil.", func() {
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryTimesAddBatch(t *testing.T) {
	Convey("TimesAddBatch", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			arr   = []*lotmdl.TimesConf{}
		)
		Convey("When everything goes positive", func() {
			r, err := d.TimesAddBatch(tx, c, arr)
			Convey("Then err should be nil.r should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(r, ShouldNotBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryTimesUpdateBatch(t *testing.T) {
	Convey("TimesUpdateBatch", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			arr   = []*lotmdl.TimesConf{}
		)
		Convey("When everything goes positive", func() {
			r, err := d.TimesUpdateBatch(tx, c, arr)
			Convey("Then err should be nil.r should not be nil.", func() {
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryGiftAdd(t *testing.T) {
	Convey("GiftAdd", t, func() {
		var (
			c          = context.Background()
			tx, _      = d.BeginTran(c)
			sid        = ""
			name       = ""
			source     = ""
			msgTitle   = ""
			msgContent = ""
			imgUrl     = ""
			num        = int(0)
			giftType   = int(0)
			timeLimit  xtime.Time
		)
		Convey("When everything goes positive", func() {
			r, err := d.GiftAdd(tx, c, sid, name, source, msgTitle, msgContent, imgUrl, num, giftType, timeLimit)
			Convey("Then err should be nil.r should not be nil.", func() {
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryGiftEdit(t *testing.T) {
	Convey("GiftEdit", t, func() {
		var (
			c          = context.Background()
			tx, _      = d.BeginTran(c)
			id         = int64(0)
			name       = ""
			source     = ""
			msgTitle   = ""
			msgContent = ""
			imgURL     = ""
			num        = int(0)
			giftType   = int(0)
			show       = int(0)
			leastMark  = int(0)
			effect     = int(0)
			timeLimit  xtime.Time
		)
		Convey("When everything goes positive", func() {
			r, err := d.GiftEdit(tx, c, id, name, source, msgTitle, msgContent, imgURL, num, giftType, show, leastMark, effect, timeLimit)
			Convey("Then err should be nil.r should not be nil.", func() {
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryGiftTotal(t *testing.T) {
	Convey("GiftTotal", t, func() {
		var (
			c        = context.Background()
			sid      = ""
			state    = int(0)
			giftType = int(0)
		)
		Convey("When everything goes positive", func() {
			total, err := d.GiftTotal(c, sid, state, giftType)
			Convey("Then err should be nil.total should not be nil.", func() {
				So(err, ShouldBeNil)
				So(total, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryGiftList(t *testing.T) {
	Convey("GiftList", t, func() {
		var (
			c        = context.Background()
			sid      = ""
			rank     = ""
			state    = int(0)
			giftType = int(0)
			pn       = int(0)
			ps       = int(0)
		)
		Convey("When everything goes positive", func() {
			result, err := d.GiftList(c, sid, rank, state, giftType, pn, ps)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGiftWinTotal(t *testing.T) {
	Convey("GiftWinTotal", t, func() {
		var (
			c      = context.Background()
			id     = int64(0)
			giftID = int64(0)
		)
		Convey("When everything goes positive", func() {
			total, err := d.GiftWinTotal(c, id, giftID)
			Convey("Then err should be nil.total should not be nil.", func() {
				So(err, ShouldBeNil)
				So(total, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryGiftWinList(t *testing.T) {
	Convey("GiftWinList", t, func() {
		var (
			c      = context.Background()
			id     = int64(0)
			giftID = int64(0)
			pn     = int(0)
			ps     = int(0)
		)
		Convey("When everything goes positive", func() {
			result, err := d.GiftWinList(c, id, giftID, pn, ps)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGiftUpload(t *testing.T) {
	Convey("GiftUpload", t, func() {
		var (
			c     = context.Background()
			tx, _ = d.BeginTran(c)
			lotID = int64(1)
			aid   = int64(1)
			keys  = []string{"323"}
		)
		Convey("When everything goes positive", func() {
			err := d.GiftUpload(tx, c, lotID, aid, keys)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
		Reset(func() {
			tx.Commit()
		})
	})
}

func TestLotteryGiftWinListAll(t *testing.T) {
	Convey("GiftWinListAll", t, func() {
		var (
			c      = context.Background()
			id     = int64(0)
			giftID = int64(0)
		)
		Convey("When everything goes positive", func() {
			result, err := d.GiftWinListAll(c, id, giftID)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGiftDetailByID(t *testing.T) {
	Convey("GiftDetailByID", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			giftInfo, err := d.GiftDetailByID(c, id)
			Convey("Then err should be nil.giftInfo should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(giftInfo, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryUpdateGiftEffect(t *testing.T) {
	Convey("UpdateGiftEffect", t, func() {
		var (
			c      = context.Background()
			id     = int64(0)
			effect = int(0)
		)
		Convey("When everything goes positive", func() {
			err := d.UpdateGiftEffect(c, id, effect)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGiftTaskCheck(t *testing.T) {
	Convey("GiftTaskCheck", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			task, err := d.GiftTaskCheck(c)
			Convey("Then err should be nil.task should not be nil.", func() {
				So(err, ShouldBeNil)
				So(task, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryUploadStatusUpdate(t *testing.T) {
	Convey("UploadStatusUpdate", t, func() {
		var (
			c      = context.Background()
			status = int(0)
			id     = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.UploadStatusUpdate(c, status, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLotteryCheckAction(t *testing.T) {
	Convey("CheckAction", t, func() {
		var (
			c          = context.Background()
			actionType = int(0)
			info       = ""
		)
		Convey("When everything goes positive", func() {
			result, err := d.CheckAction(c, actionType, info)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestLotteryCountUpload(t *testing.T) {
	Convey("CountUpload", t, func() {
		var (
			c      = context.Background()
			lotID  = int64(0)
			giftID = int64(0)
		)
		Convey("When everything goes positive", func() {
			count, err := d.CountUpload(c, lotID, giftID)
			Convey("Then err should be nil.count should not be nil.", func() {
				So(err, ShouldBeNil)
				So(count, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLeastMarkCheckList(t *testing.T) {
	Convey("LeastMarkCheckList", t, func() {
		var (
			c   = context.Background()
			sid = ""
		)
		Convey("When everything goes positive", func() {
			result, err := d.LeastMarkCheckList(c, sid)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestLotteryUpdateOperatorBySID(t *testing.T) {
	Convey("UpdateOperatorBySID", t, func() {
		var (
			c        = context.Background()
			sid      = ""
			operator = ""
		)
		Convey("When everything goes positive", func() {
			err := d.UpdateOperatorBySID(c, sid, operator)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGiftWinListWithoutAid(t *testing.T) {
	Convey("GiftWinListWithoutAid", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			result, err := d.GiftWinListWithoutAid(c, id)
			Convey("Then err should be nil.result should not be nil.", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestLotteryRawTimesByID(t *testing.T) {
	Convey("RawTimesByID", t, func() {
		var (
			c     = context.Background()
			lotID = int64(166)
		)
		Convey("When everything goes positive", func() {
			count, err := d.RawTimesByID(c, lotID)
			Convey("Then err should be nil.count should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(count)
			})
		})
	})
}

func TestLotteryBacthAddLotTimes(t *testing.T) {
	Convey("BacthAddLotTimes", t, func() {
		var (
			c     = context.Background()
			id    = int64(52)
			times = int64(1)
			cid   = int64(166)
			mids  = []int64{27543612}
		)
		Convey("When everything goes positive", func() {
			err := d.BacthAddLotTimes(c, id, times, cid, mids, strconv.FormatInt(time.Now().Unix(), 10))
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLotteryLotteryTimesKey(t *testing.T) {
	Convey("lotteryTimesKey", t, func() {
		var (
			sid    = int64(52)
			mid    = int64(27543612)
			remark = "add"
		)
		Convey("When everything goes positive", func() {
			str := lotteryTimesKey(sid, mid, remark)
			Convey("Then err should be nil.", func() {
				Print(str)
			})
		})
	})
}

func TestLotteryDeleteLotteryTimesCache(t *testing.T) {
	Convey("DeleteLotteryTimesCache", t, func() {
		var (
			c   = context.Background()
			sid = int64(52)
			mid = int64(27543612)
		)
		Convey("When everything goes positive", func() {
			err := d.DeleteLotteryTimesCache(c, sid, mid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGiftDetailBySid(t *testing.T) {
	Convey("GiftDetailByID", t, func() {
		var (
			c   = context.Background()
			sid = "cd828933-4c94-11ea-bfa0-246e9693a590"
		)
		Convey("When everything goes positive", func() {
			giftInfo, err := d.GiftDetailBySid(c, sid)
			Convey("Then err should be nil.giftInfo should not be nil.", func() {
				So(err, ShouldBeNil)
				Println(giftInfo)
			})
		})
	})
}
