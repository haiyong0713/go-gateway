package bangumi

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

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

func TestPGC(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			aid, mid        int64
			build           int
			mobiApp, device string
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.pgc).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"allow_download": "1",
				"cover": "http://i2.hdslb.com/sp/99/9969a9a988cba14e365cfe8a2b0d4115.jpg",
				"is_finish": "1",
				"newest_ep_id": "50185",
				"newest_ep_index": "10",
				"season_id": "5641",
				"title": "夏目友人帐 第四季",
				"total_count": "13",
				"weekday": "2",
				"is_jump": 0,
				"season_type": 1,
				"ogv_play_url": "https://www.bilibili.com/bangumi/play/ep33282",
				"user_season": {
					"attention": "1"
				},
				"player": {
					"aid": 10492,
					"vid": "11459795",
					"cid": 501284,
					"from": "vupload"
				}
			}
		}`)
		res, err := d.PGC(context.Background(), aid, mid, build, mobiApp, device)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestMovie(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			aid, mid        int64
			build           int
			mobiApp, device string
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.movie).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"background": {
					"cover": "http://i2.hdslb.com/sp/5d/5df7d96eef0f82cc8846f29e9884d7ab.jpg",
					"width": 1920,
					"height": 2000
				},
				"season": {
					"actor": [
						{
							"actor": "佐仓绫音",
							"actor_id": 0,
							"role": "保登心爱"
						},
						{
							"actor": "水濑祈",
							"actor_id": 0,
							"role": "香风智乃"
						}
					],
					"area": "日本",
					"season_id": 123,
					"cover": "http://i2.hdslb.com/sp/5d/5df7d96eef0f82cc8846f29e9884d7ab.jpg",
					"title": "ll",
					"evaluate": "这是一个评价",
					"total_duration": "3000",
					"pub_time": 1450941949,
					"video_length": "1230000",
					"tags": [
						{
							"tag_id": 123,
							"cover": "http://i0.hdslb.com/u_user/32b520371b1833bb9caa46f5dc46869e.jpg",
							"tag_name": "μ'sic forever!"
						},
						{
							"tag_id": 143,
							"cover": "http://i0.hdslb.com/u_user/32b520371b1833bb9caa46f5dc46869e.jpg",
							"tag_name": "μ'sic forever!"
						}
					]
				},
				"activity": {
					"activity_id": "1",
					"cover": "http://i2.hdslb.com/sp/5d/5df7d96eef0f82cc8846f29e9884d7ab.jpg",
					"link": "http://www.bilibili.com/bangumi/i/2762/",
					"script_src": "http://static.hdslb.com/js/jquery.min.js"
				},
				"movie_status": 1,
				"trailer_aid": 321,
				"aid": 123123,
				"allow_download": 1,
				"record": "国权像字161-2017-0493号\n新出像进字（2017）504号",
				"payment": {
					"price_ios": "6.0",
					"price": "5.0",
					"product_id": "tv.danmaku.pay_bangumi6bp",
					"pay_begin_time": "2015-12-11 12:00:00"
				},
				"pay_user": {
					"status": 1
				},
				"list": [
					{
						"page": 1,
						"type": "vupload",
						"part": "",
						"cid": 5398010,
						"vid": "vupload_5398010",
						"has_alias": false
					}
				]
			}
		}`)
		res, err := d.Movie(context.Background(), aid, mid, build, mobiApp, device)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestSeasonidAid(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			moiveID int64
			now     time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.seasonidAidURL).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"123456": {
					"aid": 123
				},
				"123457": {
					"aid": 456
				}
			}
		}`)
		res, err := d.SeasonidAid(context.Background(), moiveID, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestCard(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			mid  int64
			sids []int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.card).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"5641": {
					"season_id": 5641,
					"season_type": 1,
					"season_type_name": "番剧",
					"url": "https://www.bilibili.com/bangumi/play/ss26703",
					"is_follow": 0,
					"is_selection": 0,
					"badges": [
						{
							"text": "抢先",
							"text_color": "#BAFF7DE",
							"text_color_night": "#BAFF7DE",
							"bg_color": "#BAFF7DE",
							"bg_color_night": "#BAFF7DE",
							"border_color": "#BAFF7DE",
							"border_color_night": "#BAFF7DE",
							"bg_style": 1
						}
					],
					"episodes": [
						{
							"id": 30185,
							"url": "https://www.bilibili.com/bangumi/play/ep26703",
							"badges": [
								{
									"text": "抢先",
									"text_color": "#BAFF7DE",
									"text_color_night": "#BAFF7DE",
									"bg_color": "#BAFF7DE",
									"bg_color_night": "#BAFF7DE",
									"border_color": "#BAFF7DE",
									"border_color_night": "#BAFF7DE",
									"bg_style": 1
								}
							],
							"cover": "http://i0.hdslb.com/video/c8/c869945dfdad179772f052a73be04da5.jpg",
							"index": "13",
							"index_title": "【合集】夏目友人帐 肆【bilibili正版】"
						}
					]
				}
			}
		}`)
		res, err := d.Card(context.Background(), mid, sids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
