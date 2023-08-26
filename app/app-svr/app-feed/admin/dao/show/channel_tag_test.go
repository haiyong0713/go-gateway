package show

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestPopChannelTagAdd(t *testing.T) {
	convey.Convey("PopChannelTagAdd", t, func(ctx convey.C) {
		var (
			param = &show.PopChannelTagAD{
				TagID:         1,
				TopEntranceId: 1,
				Deleted:       0,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopChannelTagAdd(context.Background(), nil, param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopChannelTagDelete(t *testing.T) {
	convey.Convey("PopChannelTagDelete", t, func(ctx convey.C) {
		var (
			ID    = int64(1)
			tagid = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopChannelTagDelete(context.Background(), ID, tagid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopChannelTagNotDelete(t *testing.T) {
	convey.Convey("PopChannelTagNotDelete", t, func(ctx convey.C) {
		var (
			ID    = int64(1)
			tagid = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopChannelTagNotDelete(context.Background(), ID, tagid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopCTFindByTEID(t *testing.T) {
	convey.Convey("PopCTFindByTEID", t, func(ctx convey.C) {
		var (
			id = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.PopCTFindByTEID(context.Background(), id)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
