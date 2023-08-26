package bwsonline

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBws_onlineRawPrintList(t *testing.T) {
	Convey("RawPrintList", t, func() {
		var (
			ctx = context.Background()
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawPrintList(ctx, 8)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestBws_onlineRawPrint(t *testing.T) {
	Convey("RawPrint", t, func() {
		var (
			ctx = context.Background()
			id  = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawPrint(ctx, id)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestBws_onlineRawPrintByIDs(t *testing.T) {
	Convey("RawPrintByIDs", t, func() {
		var (
			ctx = context.Background()
			ids = []int64{1, 2}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawPrintByIDs(ctx, ids)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
