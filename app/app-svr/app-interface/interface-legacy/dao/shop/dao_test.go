package shop

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-interface")
		flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
		flag.Set("tree_id", "2688")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

// TestInfo test info
func TestInfo(t *testing.T) {
	var (
		c       = context.TODO()
		mid     = int64(27515399)
		mobiApp = "iphone"
		device  = "phone"
		build   = 8310
	)
	Convey("Ping", t, func(ctx C) {
		d.client.SetTransport(gock.DefaultTransport)
		ctx.Convey("When everthing goes positive", func(ctx C) {
			httpMock("GET", d.newShop).Reply(200).JSON(`{"code":0,"message":"success","data":{"shop":{"id":2444,"mid":11111,"name":"千笑的实体店14666","url":"https://mall.bilibili.com/shop/index.html?shopId=2244&noTitleBar=1&loadingShow=1","status":1,"logo":"//i0.hdslb.com/bfs/mall/mall/33/e2/33e23b37ceee7e80da9d4b7a7d3395d9.png"}}}`)
			res, err := d.Info(c, mid, mobiApp, device, build)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx C) {
				ctx.So(err, ShouldBeNil)
				ctx.So(res, ShouldNotBeEmpty)
			})
		})
		ctx.Convey("When data is null", func(ctx C) {
			httpMock("GET", d.newShop).Reply(200).JSON(`{"code":0,"message":"success","data":null}`)
			res, err := d.Info(c, mid, mobiApp, device, build)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx C) {
				ctx.So(err, ShouldBeNil)
				ctx.So(res, ShouldNotBeEmpty)
			})
		})
		ctx.Convey("When res.Code != ecode.OK.Code()", func(ctx C) {
			httpMock("GET", d.newShop).Reply(200).JSON(`{"code":-1,"message":"faild","data":null}`)
			_, err := d.Info(c, mid, mobiApp, device, build)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx C) {
				ctx.So(err, ShouldNotBeNil)
			})
		})
		ctx.Convey("When http request failed", func(ctx C) {
			httpMock("GET", d.newShop).Reply(500)
			_, err := d.Info(c, mid, mobiApp, device, build)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx C) {
				ctx.So(err, ShouldNotBeNil)
			})
		})
	})
}
