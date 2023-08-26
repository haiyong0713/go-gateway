package show

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestPopLargeCardAdd(t *testing.T) {
	convey.Convey("PopLargeCardAdd", t, func(ctx convey.C) {
		var (
			param = &show.PopLargeCardAD{
				Deleted:   0,
				Title:     "Test",
				CardType:  "av_largecard",
				RID:       123456,
				WhiteList: "123456,124578",
				CreateBy:  "Yunzhan",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopLargeCardAdd(context.Background(), param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopLargeCardUpdate(t *testing.T) {
	convey.Convey("PopLargeCardUpdate", t, func(ctx convey.C) {
		var (
			param = &show.PopLargeCardUP{
				ID:        1,
				Title:     "Test1",
				CardType:  "av_largecard",
				RID:       123456,
				WhiteList: "123456,124578,1",
				CreateBy:  "Yunzhan1",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopLargeCardUpdate(context.Background(), param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopLargeCardDelete(t *testing.T) {
	convey.Convey("PopLargeCardDelete", t, func(ctx convey.C) {
		var (
			ID = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopLargeCardDelete(context.Background(), ID)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopLargeCardNotDelete(t *testing.T) {
	convey.Convey("PopLargeCardNotDelete", t, func(ctx convey.C) {
		var (
			ID = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopLargeCardNotDelete(context.Background(), ID)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopLargeCardList(t *testing.T) {
	convey.Convey("PopLargeCardList", t, func(ctx convey.C) {
		var (
			id       = int64(0)
			createby = ""
			rid      = int64(10099960)
			pn       = 0
			ps       = 10
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.PopLargeCardList(context.Background(), id, createby, rid, pn, ps)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
