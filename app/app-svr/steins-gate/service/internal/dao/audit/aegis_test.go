package audit

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoAddAegisMsg(t *testing.T) {
	var (
		ctx      = context.Background()
		graphID  = int64(0)
		mid      = int64(0)
		aid      = int64(0)
		state    = int64(0)
		title    = ""
		diffMsg  = ""
		varsName = ""
	)
	convey.Convey("AddAegisMsg", t, func(c convey.C) {
		err := d.AddAegisMsg(ctx, graphID, mid, aid, state, title, diffMsg, varsName)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoCancelAegisMsg(t *testing.T) {
	var (
		ctx     = context.Background()
		graphID = int64(0)
		reason  = ""
	)
	convey.Convey("CancelAegisMsg", t, func(c convey.C) {
		err := d.CancelAegisMsg(ctx, graphID, reason)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
