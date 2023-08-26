package show

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

func ctx() context.Context {
	return context.Background()
}

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

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestCard(t *testing.T) {
	Convey("Card", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.getCard).Reply(200).JSON(`{
			"errno": 0,
			"errtag": 0,
			"msg": "",
			"data": [
				{
					"city_name": "",
					"district_name": "",
					"end_time": 0,
					"etime": "",
					"hide": 0,
					"id": 10004840,
					"is_sale": 1,
					"name": "万代 S.H.F 假面骑士 时王 Zi-O 成品模型 代理版",
					"performance_image": "",
					"performance_imagep": "//i0.hdslb.com/bfs/mall/mall/61/cc/61ccdafa062be4f207914a8faa05056b.png",
					"price_high": 369,
					"price_low": 369,
					"priceht": "￥369",
					"pricelt": "￥369",
					"province_name": "",
					"sale_flag": "",
					"sale_flag_num": 0,
					"start_time": 0,
					"status": 0,
					"stime": "",
					"subname": "",
					"tags": [
						
					],
					"type": 2,
					"url": "https://mall.bilibili.com/detail.html?msource=tianma_10004840&loadingShow=1&noTitleBar=1#itemsId=10004840",
					"venue_name": "",
					"want": "1204人想要"
				}
			]
		}`)
		var (
			ids []int64
		)
		res, err := d.Card(ctx(), ids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
