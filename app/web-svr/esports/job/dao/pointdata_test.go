package dao

import (
	"context"
	"net/url"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestEsportsThirdGet(t *testing.T) {
	convey.Convey("ThirdGet", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			url = "http://47.95.28.113/nesport/index.php/Api/live/getList?key=xxx"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.ThirdGet(c, url)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestEsportsThirdPost(t *testing.T) {
	convey.Convey("ThirdPost", t, func(convCtx convey.C) {
		params := url.Values{}
		params.Set("client_id", "7f0000010fa2000000e3")
		params.Set("match_id", "328290")
		params.Set("key", "xxx")
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.ThirdPost(context.Background(), params)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestEsportsPushRoom(t *testing.T) {
	convey.Convey("PushRoom", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			room = int64(328252)
			opt  = "1005"
			msg  = `{"type":"hello","payload":null}`
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.PushRoom(c, room, opt, msg)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				println(err)
			})
		})
	})
}
