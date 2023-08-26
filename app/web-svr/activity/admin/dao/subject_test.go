package dao

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoAddTags(t *testing.T) {
	convey.Convey("AddTags", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			tags = "keai"
			ip   = "10.256.36.68"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddTags(c, tags, ip)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				fmt.Printf("%+v", err)
			})
		})
	})
}

func TestDaoUserAwardLog(t *testing.T) {
	convey.Convey("UserAwardLog", t, func(ctx convey.C) {
		var (
			c         = context.Background()
			oid int64 = 1
			mid int64 = 0
			pn  int64 = 1
			ps  int64 = 10
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, total, err := d.UserAwardLog(c, oid, mid, pn, ps)
			ctx.So(err, convey.ShouldBeNil)
			ctx.Println(total)
			if len(list) > 0 {
				ctx.Printf("%+v", list[0])
			}
		})
	})
}

func TestDao_AddNativePage(t *testing.T) {
	convey.Convey("AddNativePage", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddNativePage(context.Background(), "测试视频数据源", "")
			fmt.Printf("native error(%+v)", err)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
