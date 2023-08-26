package rank

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-intl/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-intl")
		flag.Set("conf_token", "02007e8d0f77d31baee89acb5ce6d3ac")
		flag.Set("tree_id", "64518")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-intl-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func TestAllRank(t *testing.T) {
	Convey(t.Name(), t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.allRank).Reply(200).JSON(`{
			"note": "统计3日内新投稿的数据综合得分，每十分钟更新一次。",
			"source_date": "2019-12-19",
			"code": 0,
			"num": 100,
			"list": [
				{
					"aid": 79526660,
					"mid": 16539048,
					"score": 1641655
				},
				{
					"aid": 79261306,
					"mid": 14583962,
					"score": 1570496
				},
				{
					"aid": 79663368,
					"mid": 927587,
					"score": 1387414
				},
				{
					"aid": 79753087,
					"mid": 9824766,
					"score": 1201603
				},
				{
					"aid": 79465773,
					"mid": 337521240,
					"score": 1197418,
					"others": [
						{
							"aid": 79703910,
							"score": 460515
						}
					]
				}
			]
		}`)
		res, err := d.AllRank(context.Background())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
