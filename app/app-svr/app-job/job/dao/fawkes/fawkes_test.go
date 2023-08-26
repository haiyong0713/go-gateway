package fawkes

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLaserAll(t *testing.T) {
	convey.Convey("SendWeChart", t, func() {
		convey.Convey("When everything is correct", func(ctx convey.C) {
			httpMock("GET", "http://fawkes.bilibili.co/x/admin/fawkes/business/laser/all").Reply(200).JSON(`{}`)
			_, err := d.LaserAll(context.Background())
			ctx.Convey("Then err should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
