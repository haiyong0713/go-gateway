package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUpPre(t *testing.T) {
	convey.Convey("UpPre", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(10)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.UpPre(c, id)
			ctx.Convey("Then err should be nil.n should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestUpItemPre(t *testing.T) {
	convey.Convey("UpItemPre", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.UpItemPre(c, id)
			ctx.Convey("Then err should be nil.n should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPreSetUp(t *testing.T) {
	convey.Convey("PreSetUp", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			sid = int64(10434)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PreSetUp(c, id, sid, 1)
			ctx.Convey("Then err should be nil.n should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPreItemSetUp(t *testing.T) {
	convey.Convey("PreItemSetUp", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			sid = int64(10292)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PreItemSetUp(c, id, sid, 1)
			ctx.Convey("Then err should be nil.n should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
