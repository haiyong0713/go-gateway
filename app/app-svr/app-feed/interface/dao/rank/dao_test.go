package rank

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

func TestAllRank(t *testing.T) {
	Convey("AllRank", t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.allRank).Reply(200).JSON(`{
			"note": "统计3日内新投稿的数据综合得分，每十分钟更新一次。",
			"source_date": "2019-12-06",
			"code": 0,
			"num": 100,
			"list": [
				{
					"aid": 77959619,
					"mid": 54992199,
					"score": 1734095,
					"others": [
						{
							"aid": 78152802,
							"score": 379704
						},
						{
							"aid": 77927231,
							"score": 367470
						},
						{
							"aid": 78045688,
							"score": 341577
						}
					]
				},
				{
					"aid": 78128190,
					"mid": 6574487,
					"score": 1515992
				},
				{
					"aid": 78005909,
					"mid": 193262326,
					"score": 1373310
				},
				{
					"aid": 77963556,
					"mid": 14110780,
					"score": 1350489
				},
				{
					"aid": 78192525,
					"mid": 129240403,
					"score": 1119060
				},
				{
					"aid": 77938993,
					"mid": 203337614,
					"score": 1061084
				},
				{
					"aid": 77935473,
					"mid": 65538469,
					"score": 1001802
				},
				{
					"aid": 77829019,
					"mid": 2920960,
					"score": 966594
				},
				{
					"aid": 78143483,
					"mid": 196356191,
					"score": 965265
				},
				{
					"aid": 78014605,
					"mid": 546195,
					"score": 922379
				},
				{
					"aid": 77988803,
					"mid": 176037767,
					"score": 890178
				},
				{
					"aid": 77905674,
					"mid": 1958342,
					"score": 870383
				},
				{
					"aid": 77934749,
					"mid": 437316738,
					"score": 829701,
					"others": [
						{
							"aid": 78160466,
							"score": 789983
						}
					]
				},
				{
					"aid": 78172723,
					"mid": 279991456,
					"score": 792430,
					"others": [
						{
							"aid": 78073306,
							"score": 258343
						}
					]
				},
				{
					"aid": 77887412,
					"mid": 390461123,
					"score": 756796,
					"others": [
						{
							"aid": 78005298,
							"score": 622436
						},
						{
							"aid": 78113248,
							"score": 547331
						}
					]
				},
				{
					"aid": 78004733,
					"mid": 37439823,
					"score": 738086
				},
				{
					"aid": 77934020,
					"mid": 7487399,
					"score": 720250,
					"others": [
						{
							"aid": 78035986,
							"score": 468088
						}
					]
				},
				{
					"aid": 77985902,
					"mid": 23400436,
					"score": 694164
				},
				{
					"aid": 77874289,
					"mid": 29329085,
					"score": 649143
				},
				{
					"aid": 78101437,
					"mid": 398510,
					"score": 646016
				},
				{
					"aid": 77933933,
					"mid": 1565155,
					"score": 632186
				},
				{
					"aid": 78090940,
					"mid": 258457966,
					"score": 626155,
					"others": [
						{
							"aid": 78091866,
							"score": 292076
						}
					]
				},
				{
					"aid": 78094006,
					"mid": 26139491,
					"score": 609823
				},
				{
					"aid": 78135967,
					"mid": 37090048,
					"score": 598045
				},
				{
					"aid": 78052083,
					"mid": 209708163,
					"score": 594554
				},
				{
					"aid": 78146738,
					"mid": 10462362,
					"score": 589265
				},
				{
					"aid": 78031734,
					"mid": 222103174,
					"score": 575993
				},
				{
					"aid": 78031102,
					"mid": 386043247,
					"score": 569577
				},
				{
					"aid": 77962879,
					"mid": 179512321,
					"score": 559441
				},
				{
					"aid": 77830714,
					"mid": 7552204,
					"score": 552254
				},
				{
					"aid": 78048985,
					"mid": 471902481,
					"score": 531241
				},
				{
					"aid": 77930202,
					"mid": 294152720,
					"score": 518527
				},
				{
					"aid": 78079036,
					"mid": 7560829,
					"score": 517726,
					"others": [
						{
							"aid": 78003966,
							"score": 501919
						}
					]
				},
				{
					"aid": 77918627,
					"mid": 145149047,
					"score": 516831
				},
				{
					"aid": 77929401,
					"mid": 20165629,
					"score": 499000,
					"others": [
						{
							"aid": 77975146,
							"score": 395710
						},
						{
							"aid": 78108476,
							"score": 365347
						}
					]
				},
				{
					"aid": 77927663,
					"mid": 131452749,
					"score": 497523
				},
				{
					"aid": 78095203,
					"mid": 108572682,
					"score": 495489
				},
				{
					"aid": 78004916,
					"mid": 927587,
					"score": 494041
				},
				{
					"aid": 78106861,
					"mid": 482917999,
					"score": 493428
				},
				{
					"aid": 78003913,
					"mid": 3682229,
					"score": 491216,
					"others": [
						{
							"aid": 77943624,
							"score": 291117
						}
					]
				},
				{
					"aid": 77925899,
					"mid": 562197,
					"score": 482874
				},
				{
					"aid": 77938799,
					"mid": 10119428,
					"score": 482310
				},
				{
					"aid": 78036563,
					"mid": 431047137,
					"score": 477322
				},
				{
					"aid": 78030991,
					"mid": 414641554,
					"score": 472731
				},
				{
					"aid": 78100473,
					"mid": 268810504,
					"score": 470123
				},
				{
					"aid": 77989109,
					"mid": 13354765,
					"score": 461638
				},
				{
					"aid": 78020360,
					"mid": 10851726,
					"score": 460497
				},
				{
					"aid": 78028731,
					"mid": 128343100,
					"score": 460137
				},
				{
					"aid": 77920056,
					"mid": 324753357,
					"score": 448533
				},
				{
					"aid": 78083909,
					"mid": 384298638,
					"score": 441788,
					"others": [
						{
							"aid": 78207123,
							"score": 402972
						}
					]
				},
				{
					"aid": 78106598,
					"mid": 337521240,
					"score": 424052
				},
				{
					"aid": 78122450,
					"mid": 290526283,
					"score": 404353
				},
				{
					"aid": 77939471,
					"mid": 2072832,
					"score": 399702
				},
				{
					"aid": 78037395,
					"mid": 168598,
					"score": 394970,
					"others": [
						{
							"aid": 77983809,
							"score": 278207
						}
					]
				},
				{
					"aid": 77898292,
					"mid": 434716461,
					"score": 369639
				},
				{
					"aid": 78008105,
					"mid": 485118594,
					"score": 369454
				},
				{
					"aid": 78041651,
					"mid": 279583114,
					"score": 367105
				},
				{
					"aid": 77992753,
					"mid": 31731027,
					"score": 365595
				},
				{
					"aid": 78078278,
					"mid": 2378908,
					"score": 363744
				},
				{
					"aid": 78039675,
					"mid": 313950018,
					"score": 362991
				},
				{
					"aid": 78211736,
					"mid": 50329118,
					"score": 362981,
					"others": [
						{
							"aid": 78209907,
							"score": 261586
						}
					]
				},
				{
					"aid": 78002399,
					"mid": 32365949,
					"score": 359034,
					"others": [
						{
							"aid": 78120411,
							"score": 288596
						}
					]
				},
				{
					"aid": 78059508,
					"mid": 16720403,
					"score": 357351
				},
				{
					"aid": 77608305,
					"mid": 39627524,
					"score": 352823
				},
				{
					"aid": 77831804,
					"mid": 93890857,
					"score": 349875
				},
				{
					"aid": 77894117,
					"mid": 79061224,
					"score": 349425,
					"others": [
						{
							"aid": 78118341,
							"score": 292513
						},
						{
							"aid": 78029187,
							"score": 267170
						}
					]
				},
				{
					"aid": 78040470,
					"mid": 42230125,
					"score": 340232
				},
				{
					"aid": 78047539,
					"mid": 455876411,
					"score": 339983
				},
				{
					"aid": 77926554,
					"mid": 3957971,
					"score": 331237
				},
				{
					"aid": 78133226,
					"mid": 10451557,
					"score": 327410
				},
				{
					"aid": 78046470,
					"mid": 1935882,
					"score": 326442
				},
				{
					"aid": 78055015,
					"mid": 113362335,
					"score": 324677
				},
				{
					"aid": 78109086,
					"mid": 11403305,
					"score": 320857
				},
				{
					"aid": 77944847,
					"mid": 17819768,
					"score": 308864
				},
				{
					"aid": 77940438,
					"mid": 14333871,
					"score": 303302
				},
				{
					"aid": 77998652,
					"mid": 10330740,
					"score": 298910,
					"others": [
						{
							"aid": 78075030,
							"score": 256888
						}
					]
				},
				{
					"aid": 77927319,
					"mid": 2986310,
					"score": 298655
				},
				{
					"aid": 78074457,
					"mid": 47291,
					"score": 298478
				},
				{
					"aid": 77954817,
					"mid": 427841873,
					"score": 295421
				},
				{
					"aid": 77715541,
					"mid": 17411953,
					"score": 294042
				},
				{
					"aid": 78034506,
					"mid": 450979444,
					"score": 283485
				},
				{
					"aid": 77981134,
					"mid": 302312847,
					"score": 281366
				},
				{
					"aid": 77373228,
					"mid": 632887,
					"score": 280361
				},
				{
					"aid": 78044062,
					"mid": 6739643,
					"score": 275449
				},
				{
					"aid": 77991103,
					"mid": 154021609,
					"score": 275169
				},
				{
					"aid": 77960277,
					"mid": 258150656,
					"score": 273537
				},
				{
					"aid": 78173989,
					"mid": 476343250,
					"score": 266696
				},
				{
					"aid": 78004647,
					"mid": 250648682,
					"score": 266516
				},
				{
					"aid": 78060096,
					"mid": 287795639,
					"score": 265936
				},
				{
					"aid": 77937395,
					"mid": 415479453,
					"score": 265062,
					"others": [
						{
							"aid": 78130724,
							"score": 260983
						}
					]
				},
				{
					"aid": 78207759,
					"mid": 808171,
					"score": 262444
				},
				{
					"aid": 78051574,
					"mid": 446644785,
					"score": 254274
				},
				{
					"aid": 78034537,
					"mid": 456664753,
					"score": 252200
				},
				{
					"aid": 77943608,
					"mid": 27501395,
					"score": 252068
				},
				{
					"aid": 77907394,
					"mid": 91399769,
					"score": 250635
				},
				{
					"aid": 78011672,
					"mid": 39304265,
					"score": 247580
				},
				{
					"aid": 78141844,
					"mid": 2192108,
					"score": 244447
				},
				{
					"aid": 77940931,
					"mid": 19515012,
					"score": 243675
				},
				{
					"aid": 78046987,
					"mid": 396848107,
					"score": 243572
				},
				{
					"aid": 78025117,
					"mid": 168064909,
					"score": 243328
				}
			]
		}`)
		res, err := d.AllRank(ctx())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
