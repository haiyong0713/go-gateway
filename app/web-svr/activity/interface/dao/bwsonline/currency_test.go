package bwsonline

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBws_onlineRawUserCurrency(t *testing.T) {
	Convey("RawUserCurrency", t, func() {
		var (
			ctx = context.Background()
			mid = int64(1)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawUserCurrency(ctx, mid, 8)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDao_RawLastAutoEnergy(t *testing.T) {
	Convey("RawLastAutoEnergy", t, func() {
		var (
			ctx = context.Background()
			mid = int64(27515430)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawLastAutoEnergy(ctx, mid, 7)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				Printf("%d", p1)
			})
		})
	})
}
