package cpm

import (
	"context"
	"testing"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"
	"github.com/smartystreets/goconvey/convey"
)

func TestCpmCpmsAPP(t *testing.T) {
	convey.Convey("When cpm returns code = 0", t, func(ctx convey.C) {
		data := `{"code":0,"message":"successed","data":{}}`
		httpMock("GET", d.cpmAppURL).Reply(200).JSON(data)
		_, err := d.CpmsAPP(context.Background(), 0, 182504479, 6190, "457", "iphone", "phone", "222", "wifi", "", "", &location.InfoReply{Addr: "218.4.147.222"})
		ctx.Convey("Then Error should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
	convey.Convey("When cpm returns code != 0", t, func(ctx convey.C) {
		data := `{"code":-3,"message":"faild","data":{}}`
		httpMock("GET", d.cpmAppURL).Reply(200).JSON(data)
		_, err := d.CpmsAPP(context.Background(), 0, 182504479, 6190, "457", "iphone", "phone", "222", "wifi", "", "", &location.InfoReply{Addr: "218.4.147.222"})
		ctx.Convey("Then Error should not be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
	convey.Convey("When cpm http request gets 404", t, func(ctx convey.C) {
		httpMock("GET", d.cpmAppURL).Reply(404)
		_, err := d.CpmsAPP(context.Background(), 0, 182504479, 6190, "457", "iphone", "phone", "222", "wifi", "", "", &location.InfoReply{Addr: "218.4.147.222"})
		ctx.Convey("Then Error should not be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}
