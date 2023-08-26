package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDao_Bnj2019Conf(t *testing.T) {
	convey.Convey("Bnj2019Conf", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.bnjConfURL).Reply(200).JSON(`{"code":0,"data":{"grey_status":0,"grey_uids":""}}`)
			rs, err := d.Bnj2019Conf(c)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", rs)
			})
		})
	})
}

func TestDao_Bnj2020Conf(t *testing.T) {
	convey.Convey("Bnj2020Conf", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.bnjConfURL).Reply(200).JSON(`{"code":0,"data":{"grey_status":0,"grey_uids":""}}`)
			rs, cnt, err := d.Bnj2020Conf(c)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", rs)
				ctx.Printf("%+d", cnt)
			})
		})
	})
}
