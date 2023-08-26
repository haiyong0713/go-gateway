package white

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

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
	os.Exit(m.Run())
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestWhiteVerify(t *testing.T) {
	Convey("get WhiteVerify all", t, func() {
		var (
			urlStr = "http://api.vc.bilibili.co/promo_svr/v0/promo_svr/inner_user_check?appkey=0e9b9fcce22daaf1&ts=1576570497&uid=44452073"
			mid    = int64(44452073)
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", urlStr).Reply(200).JSON(`{
			"code": 0,
			"Data": {
				"status": 0
			}
		}`)
		res, err := d.WhiteVerify(ctx(), mid, urlStr)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
