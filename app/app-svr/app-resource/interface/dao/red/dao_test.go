package red

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
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
	d.gameClient.SetTransport(gock.DefaultTransport)
	return r
}

func TestRedDot(t *testing.T) {
	var (
		c        = context.TODO()
		mid      = int64(1)
		url      = "http://line1-game-open-api.biligame.net/message/count"
		platform = "ios"
		business = "game"
	)
	convey.Convey("RedDot", t, func(ctx convey.C) {
		httpMock("GET", url).Reply(200).JSON(`{"code":0,"data":{"red_dot":true,"type":2}}`)
		red, err := d.RedDot(c, mid, url, platform, business)
		rr, _ := json.Marshal(red)
		fmt.Printf("res(%s)", rr)
		fmt.Printf("err(%+v)", err)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
