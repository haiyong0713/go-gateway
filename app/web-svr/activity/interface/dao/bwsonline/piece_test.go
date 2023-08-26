package bwsonline

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBws_onlineRawPiece(t *testing.T) {
	Convey("RawPiece", t, func() {
		var (
			ctx = context.Background()
			id  = int64(4)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawPiece(ctx, id)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				Printf("%+v", p1)
			})
		})
	})
}

func TestBws_onlineRawUserPiece(t *testing.T) {
	Convey("UserPiece", t, func() {
		var (
			ctx = context.Background()
			mid = int64(2089809)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.RawUserPiece(ctx, mid, 8)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				Printf("%+v", p1)
			})
		})
	})
}

func TestBws_onlineRawUsedTimes(t *testing.T) {
	Convey("RawUsedTimes", t, func() {
		var (
			ctx = context.Background()
			mid = int64(2089809)
			day = int64(20200624)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.UsedTimes(ctx, mid, day)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				Printf("%+v", p1)
			})
		})
	})
}
