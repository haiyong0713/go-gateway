package fawkes

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-feed")
		flag.Set("conf_token", "OC30xxkAOyaH9fI6FRuXA0Ob5HL0f3kc")
		flag.Set("tree_id", "2686")
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
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

// TestFawkesVersion dao ut.
func TestFawkesVersion(t *testing.T) {
	Convey("get fawkes version", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.getVersion).Reply(200).JSON(`{
			"code": 0,
			"message": "0",
			"ttl": 1,
			"data": {
				"prod": {
					"5s67": {
						"config": 70,
						"ff": 73
					},
					"android": {
						"config": 226,
						"ff": 265
					},
					"biliLink": {
						"config": 142,
						"ff": 172
					},
					"ipad": {
						"config": 202,
						"ff": 171
					},
					"iphone": {
						"config": 227,
						"ff": 274
					},
					"iphone_b": {
						"config": 140,
						"ff": 170
					},
					"w19e": {
						"config": 73,
						"ff": 91
					}
				},
				"test": {
					"5s67": {
						"config": 229,
						"ff": 84
					},
					"android": {
						"config": 234,
						"ff": 271
					},
					"iphone": {
						"config": 233,
						"ff": 273
					},
					"w19e": {
						"config": 230,
						"ff": 260
					}
				}
			}
		}`)
		res, err := d.FawkesVersion(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldBeEmpty)
	})
}
