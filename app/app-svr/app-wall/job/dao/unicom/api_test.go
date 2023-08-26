package unicom

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestUnicomFlowQry(t *testing.T) {
	var (
		c          = context.Background()
		phone      = int(0)
		requestNo  = int64(0)
		outorderid = ""
		orderid    = ""
		ts         = time.Now()
	)
	convey.Convey("FlowQry", t, func(ctx convey.C) {
		httpMock("GET", d.unicomFlowExchangeURL).Reply(200).JSON(`{"respcode":"0000"}`)
		orderstatus, msg, err := d.FlowQry(c, phone, requestNo, outorderid, orderid, ts)
		ctx.Convey("Then err should be nil.orderstatus,msg should not be nil.", func(ctx convey.C) {
			fmt.Println(err)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(msg, convey.ShouldNotBeNil)
			ctx.So(orderstatus, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomUnicomIP(t *testing.T) {
	var (
		c   = context.Background()
		now = time.Now()
	)
	convey.Convey("UnicomIP", t, func(ctx convey.C) {
		httpMock("POST", d.unicomIPURL).Reply(200).JSON(`{}`)
		_, err := d.UnicomIP(c, now)
		ctx.Convey("Then err should be nil.unicomIPs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomunicomHTTPGet(t *testing.T) {
	var (
		c      = context.Background()
		urlStr = "http://www.bilibili.com"
		params url.Values
		res    = interface{}(0)
	)
	convey.Convey("unicomHTTPGet", t, func(ctx convey.C) {
		httpMock("GET", urlStr).Reply(200).JSON(`{}`)
		err := d.unicomHTTPGet(c, urlStr, params, res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomunicomHTTPPost(t *testing.T) {
	var (
		c      = context.Background()
		urlStr = "http://www.bilibili.com"
		params url.Values
		res    = interface{}(0)
	)
	convey.Convey("unicomHTTPPost", t, func(ctx convey.C) {
		httpMock("POST", urlStr).Reply(200).JSON(`{}`)
		err := d.unicomHTTPPost(c, urlStr, params, res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomurlParams(t *testing.T) {
	var (
		v url.Values
	)
	convey.Convey("urlParams", t, func(ctx convey.C) {
		p1 := d.urlParams(v)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomsign(t *testing.T) {
	var (
		params = ""
	)
	convey.Convey("sign", t, func(ctx convey.C) {
		p1 := d.sign(params)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}
