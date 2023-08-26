package bwsonline

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBws_onlineRawDress(t *testing.T) {
	Convey("RawDress", t, func() {
		var (
			ctx = context.Background()
			id  = int64(1)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawDress(ctx, id)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestBws_onlineRawDressByIDs(t *testing.T) {
	Convey("RawDressByIDs", t, func() {
		var (
			ctx = context.Background()
			ids = []int64{1, 2}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawDressByIDs(ctx, ids)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestBws_onlineRawUserDress(t *testing.T) {
	Convey("RawUserDress", t, func() {
		var (
			ctx = context.Background()
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawUserDress(ctx, mid)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
