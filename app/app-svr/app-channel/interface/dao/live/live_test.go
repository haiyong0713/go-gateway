package live

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-channel")
		flag.Set("conf_token", "a920405f87c5bbcca15f3ffebf169c04")
		flag.Set("tree_id", "7852")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
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

func init() {
	dir, _ := filepath.Abs("../../cmd/app-channel-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestAppMRoom(t *testing.T) {
	Convey("AppMRoom", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.appMRoom).Reply(200).JSON(`{
			"code": 0,
			"data": [{
				"uid": 6336952,
				"room_id": 100628,
				"title": "今天是栓王超清直播！！",
				"cover": "http://i0.hdslb.com/bfs/live/79d826bce6bd8ac61af16494b80f3bd6920a3d2c.jpg",
				"uname": "老李船长",
				"face": "http://i2.hdslb.com/bfs/face/4581a9522ebc04a78d00e4605960a786d82ca5c9.jpg",
				"online": 4266,
				"live_status": 1,
				"area_v2_parent_id": 6,
				"area_v2_parent_name": "单机",
				"area_v2_id": 221,
				"area_v2_name": "战地5",
				"playurl_h264": "http://ws.live-play.acgvideo.com/live-ws/546895/live_6336952_3740226.flv?wsSecret=a5c5f606ab6106eeeef21c87cc1a3bfd\u0026wsTime=1552911946\u0026trid=f09799d19abc4a97aa5a0f6471f7bc04\u0026sig=no",
				"accept_quality": [4],
				"current_quality": 4,
				"current_qn": 4,
				"quality_description": [{
					"qn": 4,
					"desc": "原画"
				}]
			}, {
				"uid": 11153765,
				"room_id": 23058,
				"title": "哔哩哔哩音悦台",
				"cover": "http://i0.hdslb.com/bfs/live/6029764557e3cbe91475faae26e6e244de8c1d3c.jpg",
				"uname": "3号直播间",
				"face": "http://i0.hdslb.com/bfs/face/5d35da6e93fbfb1a77ad6d1f1004b08413913f9a.jpg",
				"online": 10020,
				"live_status": 1,
				"area_v2_parent_id": 1,
				"area_v2_parent_name": "娱乐",
				"area_v2_id": 34,
				"area_v2_name": "音乐台",
				"playurl_h264": "http://ws.live-play.acgvideo.com/live-ws/987958/live_11153765_9369560.flv?wsSecret=cdd104d73f4ded63e1188568e51db7e4\u0026wsTime=1552911946\u0026trid=f09799d19abc4a97aa5a0f6471f7bc04\u0026sig=no",
				"accept_quality": [4, 3],
				"current_quality": 4,
				"current_qn": 4,
				"quality_description": [{
					"qn": 4,
					"desc": "原画"
				}, {
					"qn": 3,
					"desc": "高清"
				}]
			}]
		}`)
		res, err := d.AppMRoom(ctx(), []int64{11}, "ios")
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestFeedList(t *testing.T) {
	Convey("FeedList", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.feedList).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"rooms": [{
					"room_id": 100628,
					"face": "http://i2.hdslb.com/bfs/face/4581a9522ebc04a78d00e4605960a786d82ca5c9.jpg"
				}]
			}
		}`)
		res, _, err := d.FeedList(ctx(), 1, 1, 1)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestCard(t *testing.T) {
	Convey("Card", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.card).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"1111": [{
					"room_id": 100628,
					"title": "xxxxxxx"
				}]
			}
		}`)
		res, err := d.Card(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
