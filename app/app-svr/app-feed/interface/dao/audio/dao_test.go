package audio

import (
	"context"
	"flag"
	"os"
	"path/filepath"
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
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-feed-test.toml")
		flag.Set("conf", dir)
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

func TestAudios(t *testing.T) {
	Convey("Audios", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.getAudios).Reply(200).JSON(`{
			"code":0,
			"msg":"success",
			"data":{
				"39850":{
					"title":"心【改】",
					"cover_url":"http://uat-i0.hdslb.com/bfs/static/151123640138855e2c1bc6bb184b40fe8b5440fc8f258011055.jpg.webp",
					"play_num":0,
					"record_num":4,
					"favorite_num":0,
					"menu_id":39850,
					"is_off":0,
					"author":"哔哩哔哩音频",
					"face":"https://i2.hdslb.com/bfs/face/83e562832e65d3eb3b8db27e43263fe8f36a3f64.jpg",
					"pa_time":0,
					"type":4,
					"chn_tieup":null,
					"intro":null,
					"is_pay":null,
					"ctgs":[
						{
							"item_val":"其他有声",
							"item_id":8,
							"cate_id":1,
							"schema":"bilibili://music/menus/missevan?itemId=8&cateId=1&itemVal=其他有声"
						}
					],
					"songs":[
						{
							"title":"声优名作演绎《心》【1-2】宫野真守；速水奖；石田彰等",
							"author":"M站",
							"song_id":26902
						},
						{
							"title":"声优名作演绎《心》【3-4】宫野真守；速水奖；石田彰等",
							"author":"M站",
							"song_id":26903
						}
					]
				},
				"47157":{
					"title":"Various Artists - 海绵宝宝片尾曲的副本.mp3.flac",
					"cover_url":"http://uat-i0.hdslb.com/bfs/static/c08cb8ecf43300ab1fcfbb3894b88e72bcbeee2e.jpg",
					"play_num":0,
					"record_num":5,
					"favorite_num":0,
					"menu_id":47157,
					"is_off":1,
					"author":"哔哩哔哩音频",
					"face":"https://i2.hdslb.com/bfs/face/83e562832e65d3eb3b8db27e43263fe8f36a3f64.jpg",
					"pa_time":1522031617,
					"type":5,
					"chn_tieup":null,
					"intro":null,
					"is_pay":null,
					"ctgs":[
		 
					],
					"songs":[
		 
					]
				}
			}
		}`)
		var (
			ids = []int64{39850, 47157}
		)
		res, err := d.Audios(context.Background(), ids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
