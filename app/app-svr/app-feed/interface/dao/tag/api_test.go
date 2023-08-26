package tag

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

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestHots(t *testing.T) {
	Convey("Hots", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.hot).Reply(200).JSON(`{
			"code": 0,
			"message": "0",
			"ttl": 1,
			"data": [
				{
					"rid": 195,
					"tags": [
						{
							"tag_id": 8061,
							"tag_name": "三国",
							"highlight": 0,
							"is_atten": 0
						},
						{
							"tag_id": 20230,
							"tag_name": "一人之下",
							"highlight": 0,
							"is_atten": 0
						},
						{
							"tag_id": 499,
							"tag_name": "音乐",
							"highlight": 0,
							"is_atten": 0
						},
						{
							"tag_id": 14095,
							"tag_name": "原创音乐",
							"highlight": 0,
							"is_atten": 0
						},
						{
							"tag_id": 10040,
							"tag_name": "测试1",
							"highlight": 0,
							"is_atten": 0
						},
						{
							"tag_id": 536,
							"tag_name": "原创",
							"highlight": 0,
							"is_atten": 0
						},
						{
							"tag_id": 600,
							"tag_name": "测试",
							"highlight": 0,
							"is_atten": 0
						}
					]
				}
			]
		}`)
		var (
			mid int64
			rid int16
			now time.Time
		)
		res, err := d.Hots(ctx(), mid, rid, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestAdd(t *testing.T) {
	Convey("Add", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.add).Reply(200).JSON(`{ "code": 0, "message": "ok" }`)
		var (
			mid, tid int64
			now      time.Time
		)
		err := d.Add(ctx(), mid, tid, now)
		So(err, ShouldBeNil)
	})
}

func TestCancel(t *testing.T) {
	Convey("Cancel", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.cancel).Reply(200).JSON(`{ "code": 0, "message": "ok" }`)
		var (
			mid, tid int64
			now      time.Time
		)
		err := d.Cancel(ctx(), mid, tid, now)
		So(err, ShouldBeNil)
	})
}

func TestTags(t *testing.T) {
	Convey("Tags", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.tags).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"111111": [
					{
						"tag_id": 1,
						"tag_name": "公告",
						"cover": "http://i2.hdslb.com/bfs/face/ab4a8402a2debb51ea7dcb73384f80b9df78d02d.jpg",
						"content": "公告的tag",
						"type": 0,
						"state": 0,
						"ctime": 1433151310,
						"count": {
							"view": 0,
							"use": 2,
							"atten": 1
						},
						"is_atten": 1,
						"likes": 36,
						"hates": 2,
						"liked": 0,
						"hated": 1,
						"attribute": 1
					},
					{
						"tag_id": 2,
						"tag_name": "金馆长",
						"cover": "http://i2.hdslb.com/bfs/face/ab4a8402a2debb51ea7dcb73384f80b9df78d02d.jpg",
						"content": "金馆长的tag",
						"type": 0,
						"state": 0,
						"ctime": 1433151310,
						"count": {
							"view": 0,
							"use": 2,
							"atten": 1
						},
						"is_atten": 1,
						"likes": 36,
						"hates": 2,
						"liked": 0,
						"hated": 1,
						"attribute": 1
					}
				],
				"message": "ok"
			}
		}`)
		var (
			mid  int64
			aids []int64
			now  time.Time
		)
		res, err := d.Tags(ctx(), mid, aids, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestDetail(t *testing.T) {
	Convey("Detail", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.detail).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"info": {
					"tag_id": 9222,
					"tag_name": "英雄联盟",
					"cover": "/sp/87/8732428c57ecc95acc4757ce4d241cf1{IMG}.jpg",
					"content": "《英雄联盟》是由美国Riot Games开发的3D大型竞技场战网游戏，其主创团队是由实力强劲的Dota-Allstars的核心人物，以及暴雪等著名游戏公司的美术、程序、策划人员组成，将DOTA的玩法从对战平台延伸到网络游戏世界。除了DotA",
					"type": 0,
					"is_atten": 1,
					"count": {
						"use": 45372,
						"atten": 4570
					},
					"ctime": 1433151310
				},
				"similar": [
					{
						"tid": 43457,
						"tname": "综漫"
					}
				],
				"news": {
					"count": 0,
					"archives": [
						{
							"aid": 4632649,
							"tid": 15,
							"tname": "连载剧集",
							"copyright": 2,
							"pic": "http://i0.hdslb.com/bfs/archive/1c6e30a4645ad80ffd69db3932ddd803c8c42533.jpg",
							"title": "【国产】我的奇妙男友 14",
							"pubdate": 1463060350,
							"desc": "基因突变人薛灵乔（金泰焕饰）沉睡百年后。",
							"attribute": 524288,
							"duration": 0,
							"tags": [
								"国产"
							],
							"rights": {
								"bp": 0,
								"elec": 0,
								"download": 0,
								"movie": 0
							},
							"owner": {
								"mid": 6655892,
								"name": "腾讯电视剧",
								"face": "http://i2.hdslb.com/bfs/face/7cb7b4cc9361dc3ae4621adc6941a40ebe33c066.jpg"
							},
							"stat": {
								"view": 36655,
								"danmaku": 2447,
								"reply": 87,
								"favorite": 683,
								"coin": 91,
								"share": 4
							}
						}
					]
				}
			},
			"message": "ok"
		}`)
		var (
			tagID, pn, ps int
			now           time.Time
		)
		res, err := d.Detail(ctx(), tagID, pn, ps, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
