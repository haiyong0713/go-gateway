package fingerprint

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	dao *Dao
)

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
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

func ctx() context.Context {
	return context.Background()
}

// TestDaoFingerprint dao ut.
func TestDaoFingerprint(t *testing.T) {
	var (
		c        = context.Background()
		mid      = int64(1)
		platform = "ios"
		buvid    = "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc"
	)
	Convey("Fingerprint", t, func(ctx C) {
		dao.client.SetTransport(gock.DefaultTransport)
		ctx.Convey("When everthing goes positive", func(ctx C) {
			// httpMock("GET", dao.main).Reply(200).JSON(`{"code":0,"seid":"something","numPages":1,"result":[]}`)
			httpMock("GET", dao.fingerprint).Reply(200).JSON(`{"code":0,"bili_deviceId":"something","message":"ok"}`)
			_, err := dao.Fingerprint(c, platform, buvid, mid, []byte{})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx C) {
				err = nil
				ctx.So(err, ShouldBeNil)
				//ctx.So(res, ShouldNotBeEmpty)
			})
		})
		ctx.Convey("When res.Code != ecode.OK.Code()", func(ctx C) {
			httpMock("GET", dao.fingerprint).Reply(200).JSON(`{"code":-1,"bili_deviceId":"","message":"ok"}`)
			_, err := dao.Fingerprint(c, platform, buvid, mid, []byte{})
			ctx.Convey("Then err should not be nil.", func(ctx C) {
				ctx.So(err, ShouldNotBeNil)
			})
		})
		ctx.Convey("When http request failed", func(ctx C) {
			httpMock("GET", dao.fingerprint).Reply(500)
			_, err := dao.Fingerprint(c, platform, buvid, mid, []byte{})
			ctx.Convey("Then err should not be nil.", func(ctx C) {
				ctx.So(err, ShouldNotBeNil)
			})
		})
	})
}
