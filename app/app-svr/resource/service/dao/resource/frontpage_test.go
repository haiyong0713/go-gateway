package resource

import (
	"context"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFrontpageRawDefaultPage(t *testing.T) {
	Convey("RawDefaultPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{}
		)
		Convey("When everything goes positive", func() {
			ret, err := d.RawDefaultPage(c, req)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
				log.Info("Default frontPage: %+v", ret)
			})
		})
	})
}

func TestFrontpageRawOnlinePage(t *testing.T) {
	Convey("RawOnlinePage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 3129,
			}
		)
		Convey("When everything goes positive", func() {
			ret, err := d.RawOnlinePage(c, req)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
				log.Info("Online frontPage: %+v", ret)
			})
		})
	})
}

func TestFrontpageRawHiddenPage(t *testing.T) {
	Convey("RawHiddenPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 3129,
			}
		)
		Convey("When everything goes positive", func() {
			ret, err := d.RawHiddenPage(c, req)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				log.Info("Hidden frontPage: %+v", ret)
			})
		})
	})
}

func TestFrontpageGetFrontPage(t *testing.T) {
	Convey("GetFrontPage", t, func() {
		var (
			c        = context.Background()
			querySql = _queryOnlineSQL
			req      = &pb.FrontPageReq{
				ResourceId: 3129,
			}
		)
		Convey("When everything goes positive", func() {
			ret, err := d.GetFrontPage(c, querySql, req)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
				log.Info("select Online: %+v", ret)
			})
		})
	})
}

func TestFrontpageGetEffectiveFrontPage(t *testing.T) {
	Convey("GetEffectiveFrontPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 3129,
			}
		)
		Convey("When everything goes positive", func() {
			ret, err := d.GetEffectiveFrontPage(c, req)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
				log.Info("FrontPage reply: %+v", ret)
			})
			Convey("Test ret.Style", func() {
				idx := len(ret.Online)
				So(ret.Online[idx-1].IsSplitLayer, ShouldNotBeZeroValue)
				log.Info("Online[%d].is_split_layer = %d", idx-1, ret.Online[idx-1].IsSplitLayer)
			})
		})
	})
}
