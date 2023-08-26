package article

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestReplyAdd(t *testing.T) {
	c := context.Background()
	convey.Convey("ReplyAdd", t, func(ctx convey.C) {
		err := d.ReplyAdd(c, 1, 1, "ceshiceshi")
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSignature(t *testing.T) {
	convey.Convey("signature", t, func(ctx convey.C) {
		params := url.Values{
			"ts":     []string{strconv.FormatInt(1617948975, 10)},
			"appkey": []string{d.c.HTTPClient.Key},
		}
		sign := signature(params, d.c.HTTPClient.Secret)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(len(sign), convey.ShouldBeGreaterThan, 0)
		})
	})
}
