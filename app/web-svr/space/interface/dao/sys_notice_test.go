package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSysNoticelist(t *testing.T) {
	convey.Convey("SysNoticelist", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			sysNotice, err := d.SysNoticelist(c)
			ctx.Convey("Then err should be nil.sysNotice should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(sysNotice, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSysNoticeUIDlist(t *testing.T) {
	convey.Convey("SysNoticeUIDlist", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			sysNoticeUID, err := d.SysNoticeUIDlist(c)
			ctx.Convey("Then err should be nil.sysNoticeUID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(sysNoticeUID, convey.ShouldNotBeNil)
			})
		})
	})
}
