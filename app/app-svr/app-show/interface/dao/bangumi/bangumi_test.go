package bangumi

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

	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"github.com/golang/mock/gomock"
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

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestRecommend(t *testing.T) {
	Convey("Recommend", t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.rcmmd).Reply(200).JSON(`{
			"code": 0,
			"count": "41",
			"message": "success",
			"pages": "14",
			"result": [
				{
					"actor": [
						
					],
					"allow_bp": "0",
					"allow_download": "1",
					"area": "日本",
					"arealimit": 0,
					"bangumi_id": "2073",
					"bangumi_title": "历物语",
					"brief": "在和美丽吸血鬼遭遇的春假之后，主人公阿良良木历迎来新学期，从新学期的第1个月开始到大学入学考试为止，...",
					"copyright": "ugc",
					"cover": "http://x-img.hdslb.net/group1/M00/BC/83/oYYBAFacRvKAUBTxAAPTMmSMmdI876.jpg",
					"danmaku_count": "19867",
					"episodes": [
						
					],
					"evaluate": "",
					"favorites": "51278",
					"is_finish": "0",
					"last_time": "2016-02-14 17:04:11.0",
					"new_cover": "http://i1.hdslb.com/video/e9/e98b1f0ae2017723ca8024593ff16537.jpg",
					"new_ep": {
						"av_id": "3836965",
						"cover": "http://i1.hdslb.com/video/e9/e98b1f0ae2017723ca8024593ff16537.jpg",
						"danmaku": "6164598",
						"episode_id": "83887",
						"index": "6",
						"index_title": "【1月】历物语 06【极影】",
						"page": "1",
						"up": {
							
						},
						"update_time": "2016-02-14 17:04:11.0"
					},
					"newest_ep_id": "83887",
					"newest_ep_index": "6",
					"play_count": "981295",
					"pub_time": "2016-01-09 00:00:00",
					"related_seasons": [
						
					],
					"season_id": "3297",
					"season_title": "历物语",
					"seasons": [
						
					],
					"share_url": "http://www.bilibili.com/bangumi/i/3297/",
					"spid": "0",
					"squareCover": "http://x-img.hdslb.net/group1/M00/BC/83/oYYBAFacRw6AIbnYAAE_lEwX5PQ804.jpg",
					"staff": "",
					"tag2s": [
						
					],
					"tags": [
						{
							"cover": "http://x-img.hdslb.net/group1/M00/98/7E/oYYBAFaV0zSAdUhyAAChgFrpbZM069.jpg",
							"index": "10",
							"orderType": 0,
							"style_id": "0",
							"tag_id": "112",
							"tag_name": "16年一月新番",
							"type": "0"
						},
						{
							"cover": "http://i1.hdslb.com/u_user/9fc0e61cdba27a26ea969bac28d39606.png",
							"index": "110",
							"orderType": 0,
							"style_id": "26",
							"tag_id": "57",
							"tag_name": "奇幻",
							"type": "0"
						},
						{
							"cover": "http://i0.hdslb.com/group1/M00/5B/86/oYYBAFazCLCAP0XdAADUTHL-Xuc228.jpg",
							"index": "1",
							"orderType": 0,
							"style_id": "0",
							"tag_id": "84",
							"tag_name": "连载中",
							"type": "0"
						},
						{
							"cover": "http://i1.hdslb.com/u_user/b6ed574c94b249d990369d49eebb401b.jpg",
							"index": "120",
							"orderType": 0,
							"style_id": "24",
							"tag_id": "93",
							"tag_name": "校园",
							"type": "0"
						}
					],
					"title": "历物语",
					"total_count": "4",
					"trailerAid": "5",
					"user_season": {
						"attention": "0",
						"last_ep_index": "0",
						"last_time": "0"
					},
					"viewRank": 0,
					"watchingCount": "0",
					"weekday": "-1"
				}
			]
		}`)
		res, err := d.Recommend(time.Now())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestSeasonid(t *testing.T) {
	Convey("Seasonid", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.seasonidURL).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": {
				"123456": {
					"season_id": 123,
					"season_type": 4,
					"episode_id": 123
				},
				"123457": {
					"season_id": 123,
					"season_type": 1,
					"episode_id": 123
				},
				"123457": {
					"season_id": 123,
					"season_type": 2,
					"episode_id": 1234
				}
			}
		}`)
		var (
			aids []int64
			now  time.Time
		)
		res, err := d.Seasonid(aids, now)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
func TestBanners(t *testing.T) {
	Convey("Banners", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.bannerURL).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"result": [
				{
					"img": "http://i0.hdslb.com/group1/M00/94/1C/oYYBAFbfs1SAX7NSAACKuHxUZjk323.jpg",
					"link": "http://www.bilibili.com/video/av12345/",
					"title": "各种测"
				},
				{
					"img": "http://i0.hdslb.com/group1/M00/94/1C/oYYBAFbfs1SAX7NSAACKuHxUZjk323.jpg",
					"link": "http://bangumi.bilibili.com/anime/3333",
					"title": "测接口2"
				}
			]
		}`)
		res, err := d.Banners(context.TODO(), 13)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestCardsByAids(t *testing.T) {
	Convey("CardsByAids", t, func() {
		var (
			aids     []int64
			mockCtrl = gomock.NewController(t)
			res      map[int32]*seasongrpc.CardInfoProto
		)
		defer mockCtrl.Finish()
		mockArc := seasongrpc.NewMockSeasonClient(mockCtrl)
		d.rpcClient = mockArc
		mockArc.EXPECT().CardsByAids(context.TODO(), gomock.Any()).Return(res, nil)
		res, err := d.CardsByAids(context.TODO(), aids)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestEpPlayer(t *testing.T) {
	Convey("PosRecs", t, func() {
		res, err := d.EpPlayer(context.Background(), []int64{119526}, nil)
		fmt.Printf("%v", res)
		So(err, ShouldBeNil)
	})
}
