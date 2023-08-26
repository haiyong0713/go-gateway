package audio

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

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestAudioByCids(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			cids = []int64{1, 2, 3, 4}
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.audioByCidsURL).Reply(200).JSON(`{
			"code": 0,
			"msg": "success",
			"data": {
				"1": {
					"title": "【hanser】被害妄想手机女子",
					"song_id": 102,
					"cover_url": "http://i0.hdslb.com/bfs/test/7d92b3f2ee5353611749da421d8f1fd831b2b41c.jpg",
					"play_count": 0,
					"reply_count": 0,
					"upper_id": 26609612,
					"entrance": "一键转音频",
					"song_attr": 1
				},
				"2": {
					"title": "【李蚊香×佑可猫】千梦",
					"song_id": 103,
					"cover_url": "http://i0.hdslb.com/bfs/test/7555f67e1225efb9cbd98cad04e1158cbb2368ec.jpg",
					"play_count": 1,
					"reply_count": 0,
					"upper_id": 26609612,
					"entrance": "前往音频",
					"song_attr": 0
				},
				"3": {
					"title": "【洛天依】线【纯白】",
					"song_id": 104,
					"cover_url": "http://i0.hdslb.com/bfs/test/77a55e5e986c8d9e6377c2ca1526781dceab27b4.jpg",
					"play_count": 98,
					"reply_count": 0,
					"upper_id": 26609612,
					"entrance": "前往音频",
					"song_attr": 0
				}
			}
		}`)
		res, err := d.AudioByCids(context.Background(), cids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
