package bangumi

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"github.com/golang/mock/gomock"
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

func TestUpdates(t *testing.T) {
	Convey("get Updates all", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.updates).Reply(200).JSON(`{
			"code": 0,
			"message":"success",
			"result": {
				"square_cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
				"title": "标题标题",
				"updates": 88
			}
		}`)
		var (
			mid int64
			now time.Time
		)
		res, err := d.Updates(ctx(), mid, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestPullSeasons(t *testing.T) {
	Convey("get PullSeasons all", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.pullSeasons).Reply(200).JSON(`{
			"code": 0,
			"message":"success",
			"result": [
				{
					"bgm_type": 1,
					"season_id": "1587",
					"title": "Fate/Stay Night: Unlimited Blade Works 第二季",
					"cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
					"is_finish": "1",
					"ts": 1490058000,
					"new_ep": {
						"cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
						"episode_id": 123,
						"index": "123",
						"index_title": "长标题",
						"play": 123,
						"dm": 132,
						"url": "http://bangumi.bilibili.com/anime/1587/play#123"
					},
					"total_count": "13"
				},
				{
					"bgm_type": 1,
					"season_id": "1587",
					"title": "Fate/Stay Night: Unlimited Blade Works 第二季",
					"cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
					"is_finish": "1",
					"ts": 1490058000,
					"new_ep": {
						"cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
						"episode_id": 123,
						"index": "123",
						"index_title": "长标题",
						"play": 123,
						"dm": 132,
						"url": "http://bangumi.bilibili.com/anime/1587/play#123"
					},
					"total_count": "13"
				}
			]
		}`)
		var (
			seasonIDs []int64
			now       time.Time
		)
		res, err := d.PullSeasons(ctx(), seasonIDs, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestTestFollowPull(t *testing.T) {
	Convey("get PullSeasons all", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.followPull).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"id": 123,
				"title": "2018国萌四强",
				"cover": "http://i0.hdslb.com/bfs/bangumi/s7xgc76cgu.jpg",
				"link": "http://api.bilibili.com/pgc/moe/2018/cn/index",
				"desc": "四强公开!!!",
				"badge": "日萌场"
			}
		}`)
		var (
			mid             int64
			mobiApp, device string
			now             time.Time
		)
		res, err := d.FollowPull(ctx(), mid, mobiApp, device, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestRemind(t *testing.T) {
	Convey("Remind", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.remind).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"updates": 88,
				"list": [
					{
						"cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
						"square_cover": "http://i0.hdslb.com/bfs/bangumi/test.jpg",
						"update_desc": "《紫罗兰永恒花园》更新至第25话（完结）",
						"update_title": "你追的系列新作更新啦~",
						"season_id": 11211,
						"uri": "bilibili://main/favorite?tab=bangumi"
					}
				]
			}
		}`)
		var mid int64
		res, err := d.Remind(ctx(), mid)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestEpPlayer(t *testing.T) {
	Convey("EpPlayer", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.epPlayer).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"115317": {
					"aid": 10100590,
					"cid": 10100591,
					"season": {
						"type": 2,
						"cover": "http://uat-i0.hdslb.com/bfs/bangumi/0f6c282872757e3e72e9334fbb84ff4c2a9d27e5.jpg",
						"square_cover": "http://uat-i0.hdslb.com/bfs/bangumi/0f6c282872757e3e72e9334fbb84ff4c2a9d27e5.jpg",
						"is_finish": 0,
						"season_id": 33419,
						"title": "动态电影ceshi",
						"total_count": -1,
						"type_name": "番剧",
						
					},
					"stat": {
						"danmaku": 1,
						"play": 2,
						"reply": 3
					},
					"region_uri": "bilibili://pgc/bangumi",
					"cover": "http://i0.hdslb.com/bfs/archive/496ea8899680d4a80d163d2edb401b23.jpg",
					"episode_id": 115317,
					"short_title": "01",
					"index_title": "长标题",
					"is_finish": 1,
					"duration": 21939,
					"dimension": {
						"width": 1920,
						"height": 1080,
						"rotate": 0
					},
					"new_desc": "第1话 长标题",
					"url": "https://www.bilibili.com/bangumi/play/ep115317?season_cover=http%3a%2f%2fuat-i0.hdslb.com%2fbfs%2fbangumi%2f0f6c282872757e3e72e9334fbb84ff4c2a9d27e5.jpg&index_title=%e9%95%bf%e6%a0%87%e9%a2%98&player_info=%7b%22cid%22%3a+20000000%2c%22expire_time%22%3a+1536118637%2c%22file_info%22%3a+%7b%2216%22%3a+%5b%7b%22filesize%22%3a+1221999%2c%22timelength%22%3a+21939%7d%5d%2c%2232%22%3a+%5b%7b%22ahead%22%3a+%22EhA%3d%22%2c%22filesize%22%3a+2777010%2c%22timelength%22%3a+21910%2c%22vhead%22%3a+%22AWQAIP%2fhAB5nZAAgrNlA2D3n%2f%2fAoACfxAAADAAEAAAMAMA8YMZYBAAVo6%2bzyPA%3d%3d%22%7d%5d%2c%2264%22%3a+%5b%7b%22filesize%22%3a+4111788%2c%22timelength%22%3a+21910%7d%5d%2c%2280%22%3a+%5b%7b%22filesize%22%3a+4566257%2c%22timelength%22%3a+21907%7d%5d%7d%2c%22fnval%22%3a+8%2c%22fnver%22%3a+0%2c%22quality%22%3a+32%2c%22support_description%22%3a+%5b%22%e9%ab%98%e6%b8%85+1080P%22%2c+%22%e9%ab%98%e6%b8%85+720P%22%2c+%22%e6%b8%85%e6%99%b0+480P%22%2c+%22%e6%b5%81%e7%95%85+360P%22%5d%2c%22support_formats%22%3a+%5b%22flv%22%2c+%22flv720%22%2c+%22flv480%22%2c+%22mp4%22%5d%2c%22support_quality%22%3a+%5b80%2c+64%2c+32%2c+16%5d%2c%22url%22%3a+%22http%3a%2f%2fupos-hz-mirrorkodo.acgvideo.com%2fupgcxcode%2f00%2f00%2f20000000%2f20000000-1-32.flv%3fe%3dig8euxZM2rNcNbN17WKjnoMMhzdHhzTEto8g5X10ugNcXBlqNxHxNEVE5XREto8KqJZHUa6m5J0SqE85tZvEuENvNC8xNEVE9EKE9IMvXBvE2ENvNCImNEVEK9GVqJIwqa80WXIekXRE9IB5QK%3d%3d%26deadline%3d1536122237%26dynamic%3d1%26gen%3dplayurl%26oi%3d2871790504%26os%3dkodo%26platform%3dpc%26rate%3d215468%26trid%3db1365e41f42b4aaa93cdcbb95f5094b9%26uipk%3d5%26uipv%3d5%26um_deadline%3d1536122237%26um_sign%3d8b5b2a0a94f2210905ad0dfa926400ed%26upsig%3d26027d524be443827d71d6b5d3cc8e63%22%2c%22video_codecid%22%3a+7%2c%22video_project%22%3a+false%7d",
					"is_preview": 1,
					"player_info": {
						"cid": 20000000,
						"expire_time": 1536118637,
						"file_info": {
							"16": {
								"infos": [
									{
										"filesize": 1221999,
										"timelength": 21939
									}
								]
							},
							"32": {
								"infos": [
									{
										"ahead": "EhA=",
										"filesize": 2777010,
										"timelength": 21910,
										"vhead": "AWQAIP/hAB5nZAAgrNlA2D3n//AoACfxAAADAAEAAAMAMA8YMZYBAAVo6+zyPA=="
									}
								]
							},
							"64": {
								"infos": [
									{
										"filesize": 4111788,
										"timelength": 21910
									}
								]
							},
							"80": {
								"infos": [
									{
										"filesize": 4566257,
										"timelength": 21907
									}
								]
							}
						},
						"fnval": 8,
						"fnver": 0,
						"quality": 32,
						"support_description": [
							"高清 1080P",
							"高清 720P",
							"清晰 480P",
							"流畅 360P"
						],
						"support_formats": [
							"flv",
							"flv720",
							"flv480",
							"mp4"
						],
						"support_quality": [
							80,
							64,
							32,
							16
						],
						"url": "http://upos-hz-mirrorkodo.acgvideo.com/upgcxcode/00/00/20000000/20000000-1-32.flv?e=ig8euxZM2rNcNbN17WKjnoMMhzdHhzTEto8g5X10ugNcXBlqNxHxNEVE5XREto8KqJZHUa6m5J0SqE85tZvEuENvNC8xNEVE9EKE9IMvXBvE2ENvNCImNEVEK9GVqJIwqa80WXIekXRE9IB5QK==&deadline=1536122237&dynamic=1&gen=playurl&oi=2871790504&os=kodo&platform=pc&rate=215468&trid=b1365e41f42b4aaa93cdcbb95f5094b9&uipk=5&uipv=5&um_deadline=1536122237&um_sign=8b5b2a0a94f2210905ad0dfa926400ed&upsig=26027d524be443827d71d6b5d3cc8e63",
						"video_codecid": 7,
						"video_project": false
					}
				}
			}
		}`)
		var (
			epIDs                     []int64
			mobiApp, platform, device string
			build, fnver, fnval       int
		)
		res, err := d.EpPlayer(ctx(), epIDs, mobiApp, platform, device, build, fnver, fnval)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestCardsInfoReply(t *testing.T) {
	Convey("CardsInfoReply", t, func() {
		var (
			mockCtrl   = gomock.NewController(t)
			res        map[int32]*episodegrpc.EpisodeCardsProto
			err        error
			episodeIds []int32
		)
		defer mockCtrl.Finish()
		mockArc := episodegrpc.NewMockEpisodeClient(mockCtrl)
		d.rpcClient = mockArc
		mockArc.EXPECT().Cards(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.CardsInfoReply(ctx(), episodeIds)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestCardsByAids(t *testing.T) {
	Convey("CardsByAids", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      map[int32]*episodegrpc.EpisodeCardsProto
			err      error
			aids     []int32
		)
		defer mockCtrl.Finish()
		mockArc := episodegrpc.NewMockEpisodeClient(mockCtrl)
		d.rpcClient = mockArc
		mockArc.EXPECT().CardsByAids(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.CardsByAids(ctx(), aids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
