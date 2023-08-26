package tag

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

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

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestTagMsg(t *testing.T) {
	Convey("Hots", t, func() {
		res, err := d.TagMsg(ctx(), 174002, 1)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestTagInfo(t *testing.T) {
	Convey("TagInfo", t, func() {
		var (
			mid   int64
			tagID int
			now   time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.tagURL).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"tag_id": 9222,
				"tag_name": "英雄联盟",
				"cover": "/sp/87/8732428c57ecc95acc4757ce4d241cf1{IMG}.jpg",
				"content": "《英雄联盟》是由美国Riot Games开发的3D大型竞技场战网游戏，其主创团队是由实力强劲的Dota-Allstars的核心人物，以及暴雪等著名游戏公司的美术、程序、策划人员组成，将DOTA的玩法从对战平台延伸到网络游戏世界。除了DotA",
				"type": 0,
				"state": 0,
				"ctime": 1433151310,
				"count": {
					"view": 0,
					"use": 45372,
					"atten": 4570
				},
				"is_atten": 1,
				"likes": 12,
				"hates": 0,
				"attribute": 0,
				"liked": 0,
				"hated": 1
			},
			"message": "ok"
		}`)
		res, err := d.TagInfo(ctx(), mid, tagID, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestHots(t *testing.T) {
	Convey("Hots", t, func() {
		var (
			rid, tagID, pn, ps int
			now                time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", fmt.Sprintf(d.tagHotURL, rid, tagID)).Reply(200).JSON(`{
			"code": 0,
			"date": [
				123
			]
		}`)
		res, err := d.Hots(ctx(), rid, tagID, pn, ps, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestNewArcs(t *testing.T) {
	Convey("NewArcs", t, func() {
		var (
			rid, tagID, pn, ps int
			now                time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.tagNewURL).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"archives": [
					{
						"aid": 5761575,
						"tid": 75,
						"tname": "动物圈",
						"copyright": 2,
						"pic": "http://i0.hdslb.com/bfs/archive/df0e8b1ecbba781ff152a6123a4bc82232ffb29d.jpg",
						"title": "【汪星人】这只汪和喵简直就是汪喵界的罗密欧和茱莉叶啊",
						"pubdate": 1470896635,
						"ctime": 1470896635,
						"desc": "优酷 http://v.youku.com/v_show/id_XMTYyNDYyNDY0MA==.html?firsttime=0#paction 优酷",
						"state": 0,
						"attribute": 540672,
						"duration": 67,
						"tags": [
							"喵星人",
							"汪星人",
							"汪星人被玩坏",
							"喵星人的日常"
						],
						"rights": {
							"bp": 0,
							"elec": 0,
							"download": 0,
							"movie": 0,
							"pay": 0
						},
						"owner": {
							"mid": 17804516,
							"name": "债殿",
							"face": "http://i1.hdslb.com/bfs/face/f8a55017b6ad70ca3f4da229b99d44d8f7403352.jpg"
						},
						"stat": {
							"view": 306,
							"danmaku": 1,
							"reply": 0,
							"favorite": 6,
							"coin": 2,
							"share": 0,
							"now_rank": 0,
							"his_rank": 0
						}
					}
				],
				"page": {
					"count": 1000,
					"num": 1,
					"size": 20
				}
			},
			"message": "ok"
		}`)
		res, err := d.NewArcs(ctx(), rid, tagID, pn, ps, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestSimilarTag(t *testing.T) {
	Convey("SimilarTag", t, func() {
		var (
			rid, tagID int
			now        time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.similarTagURL).Reply(200).JSON(`{
			"code": 0,
			"data": [
				{
					"rid": 27,
					"rname": "综合",
					"tid": 43457,
					"tname": "综漫"
				}
			],
			"message": "ok"
		}`)
		res, err := d.SimilarTag(ctx(), rid, tagID, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestSimilarTagChange(t *testing.T) {
	Convey("SimilarTagChange", t, func() {
		var (
			tagID int
			now   time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.similarTagChangeURL).Reply(200).JSON(`{
			"code": 0,
			"data": [
				{
					"rid": 0,
					"rname": "",
					"tid": 2450,
					"cover": "",
					"atten": 100,
					"tname": "官方"
				},
				{
					"rid": 0,
					"rname": "",
					"tid": 7007,
					"cover": "http://i1.hdslb.com/sp/cc/cc8ada3d624447b48d65ddeacd5ac319_s.jpg",
					"atten": 1,
					"tname": "手机"
				}
			],
			"message": ""
		}`)
		res, err := d.SimilarTagChange(ctx(), tagID, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestDetail(t *testing.T) {
	Convey("Detail", t, func() {
		var (
			tagID, pn, ps int
			now           time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.tagDetailURL).Reply(200).JSON(`{
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
		res, err := d.Detail(ctx(), tagID, pn, ps, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestDetailRanking(t *testing.T) {
	Convey("DetailRanking", t, func() {
		var (
			reid, tagID, pn, ps int
			now                 time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.tagRankingURL).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"archives": [
					{
						"aid": 5761575,
						"tid": 75,
						"tname": "动物圈",
						"copyright": 2,
						"pic": "http://i0.hdslb.com/bfs/archive/df0e8b1ecbba781ff152a6123a4bc82232ffb29d.jpg",
						"title": "【汪星人】这只汪和喵简直就是汪喵界的罗密欧和茱莉叶啊",
						"pubdate": 1470896635,
						"ctime": 1470896635,
						"desc": "优酷 http://v.youku.com/v_show/id_XMTYyNDYyNDY0MA==.html?firsttime=0#paction 优酷",
						"state": 0,
						"attribute": 540672,
						"duration": 67,
						"tags": [
							"喵星人",
							"汪星人",
							"汪星人被玩坏",
							"喵星人的日常"
						],
						"rights": {
							"bp": 0,
							"elec": 0,
							"download": 0,
							"movie": 0,
							"pay": 0
						},
						"owner": {
							"mid": 17804516,
							"name": "债殿",
							"face": "http://i1.hdslb.com/bfs/face/f8a55017b6ad70ca3f4da229b99d44d8f7403352.jpg"
						},
						"stat": {
							"view": 306,
							"danmaku": 1,
							"reply": 0,
							"favorite": 6,
							"coin": 2,
							"share": 0,
							"now_rank": 0,
							"his_rank": 0
						}
					}
				],
				"page": {
					"count": 1000,
					"num": 1,
					"size": 20
				}
			},
			"message": "ok"
		}`)
		res, err := d.DetailRanking(ctx(), reid, tagID, pn, ps, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestTagArchive(t *testing.T) {
	Convey("TagArchive", t, func() {
		var (
			aid int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.tagArchiveURL).Reply(200).JSON(`{
			"code": 0,
			"data": [
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
		}`)
		res, err := d.TagArchive(ctx(), aid)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
