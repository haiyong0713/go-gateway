package resource

import (
	"context"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFrontpageDefaultPage(t *testing.T) {
	Convey("DefaultPage", t, func() {
		var (
			c       = context.Background()
			request = &pb.FrontPageReq{}
		)
		Convey("When everything goes positive", func() {
			res, err := d.DefaultPage(c, request)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestFrontpageOnlinePage(t *testing.T) {
	Convey("OnlinePage", t, func() {
		var (
			c       = context.Background()
			request = &pb.FrontPageReq{}
		)
		Convey("When everything goes positive", func() {
			res, err := d.OnlinePage(c, request)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestFrontpageHiddenPage(t *testing.T) {
	Convey("HiddenPage", t, func() {
		var (
			c       = context.Background()
			request = &pb.FrontPageReq{}
		)
		Convey("When everything goes positive", func() {
			res, err := d.HiddenPage(c, request)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}
