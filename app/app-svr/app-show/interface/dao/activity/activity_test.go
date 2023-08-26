package activity

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-show-test.toml")
	}
	flag.Parse()
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func ctx() context.Context {
	return context.Background()
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestActivitys(t *testing.T) {
	Convey("get Activitys all", t, func() {
		var (
			ids  []int64
			mold int
			ip   string
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.activitys).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"list": [
					{
						"id": 2628,
						"state": 1,
						"stime": "2016-11-15 17:29:33",
						"etime": "2016-11-15 17:29:32",
						"ctime": "2016-06-24 11:30:29",
						"mtime": "1971-01-01 00:00:00",
						"name": "我们的相遇是不可思议的奇迹",
						"author": "暮光闪闪闪",
						"pc_url": "http://www.bilibili.com/topic/1328.html",
						"rank": 13280,
						"h5_url": "",
						"pc_cover": "http://i0.hdslb.com/topic/201606/1466739029-668ed6c1275ae98d.jpg",
						"h5_cover": "",
						"page_name": "",
						"plat": 1,
						"desc": "又是一年毕业季，历历在目的曾经没有一样可以舍弃，感谢让我们相遇的奇迹！",
						"click": 254858,
						"type": 0,
						"mold": 1,
						"series": 0,
						"dept": 5,
						"reply_id": 0,
						"tp_id": 1328,
						"ptime": "0000-00-00 00:00:00"
					}
				]
			}
		}`)
		res, err := d.Activitys(ctx(), ids, mold, ip)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
