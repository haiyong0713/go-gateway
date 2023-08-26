package show

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestPopChannelResourceAdd(t *testing.T) {
	convey.Convey("PopChannelResourceAdd", t, func(ctx convey.C) {
		var (
			param = &show.PopChannelResourceAD{
				RID:           1,
				TopEntranceId: 1,
				Deleted:       0,
				State:         1,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopChannelResourceAdd(context.Background(), nil, param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopChannelResourceState(t *testing.T) {
	convey.Convey("PopChannelResourceState", t, func(ctx convey.C) {
		var (
			ID    = int64(1)
			aid   = int64(1)
			state = 2
			tagID = int64(3)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopChannelResourceState(context.Background(), ID, aid, tagID, state)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopCRFindByTEID(t *testing.T) {
	convey.Convey("PopCRFindByTEID", t, func(ctx convey.C) {
		var (
			id = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.PopCRFindByTEID(context.Background(), id)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
