package show

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

//func TestPopLiveCardAdd(t *testing.T) {
//	convey.Convey("PopLiveCardAdd", t, func(ctx convey.C) {
//		var (
//			param = &show.PopLiveCardAD{
//				CardType: "live_card",
//				RID:      123456,
//				CreateBy: "Yunzhan",
//				Cover:    "test1234",
//			}
//		)
//		ctx.Convey("When everything gose positive", func(ctx convey.C) {
//			//err, _ := d.PopLiveCardAdd(context.Background(), param)
//			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
//				//ctx.So(err, convey.ShouldBeNil)
//			})
//		})
//	})
//}

func TestPopLiveCardUpdate(t *testing.T) {
	convey.Convey("PopLiveCardUpdate", t, func(ctx convey.C) {
		var (
			param = &show.PopLiveCardUP{
				ID:    100,
				Cover: "",
				RID:   1234568,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopLiveCardUpdate(context.Background(), param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopLargeCardOperate(t *testing.T) {
	convey.Convey("PopLargeCardOperate", t, func(ctx convey.C) {
		var (
			ID    = int64(100)
			State = 1
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopLargeCardOperate(context.Background(), ID, State)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopLiveCardList(t *testing.T) {
	convey.Convey("PopLiveCardList", t, func(ctx convey.C) {
		var (
			id       = int64(0)
			createby = ""
			state    = 0
			pn       = 0
			ps       = 10
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.PopLiveCardList(context.Background(), id, state, createby, pn, ps)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
