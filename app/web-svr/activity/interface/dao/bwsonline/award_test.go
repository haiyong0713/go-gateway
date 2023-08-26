package bwsonline

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBws_onlineRawAwardPackageList(t *testing.T) {
	Convey("RawAwardPackageList", t, func() {
		var (
			ctx = context.Background()
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawAwardPackageList(ctx, 8)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestBws_onlineRawAwardByIDs(t *testing.T) {
	Convey("RawAwardByIDs", t, func() {
		var (
			ctx = context.Background()
			ids = []int64{1, 2}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawAwardByIDs(ctx, ids)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestBws_onlineRawAwardPackage(t *testing.T) {
	Convey("RawAwardPackage", t, func() {
		var (
			ctx = context.Background()
			id  = int64(1)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawAwardPackage(ctx, id)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
