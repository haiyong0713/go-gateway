package ad

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
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

func TestSplashList(t *testing.T) {
	Convey("get SplashList all", t, func() {
		var (
			mobiApp, device, buvid, birth, adExtra string
			height, width, build                   int
			mid                                    int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.splashListURL).Reply(200).JSON(`{
			"code": 0,
			"max_time": 4,
			"min_interval": 14400,
			"pull_interval": 900,
			"request_id": "1576551036561q172a22a59a108q441",
			"data": [
				{
					"id": 3,
					"type": 2,
					"card_type": 14,
					"duration": 3,
					"begin_time": 1576944000,
					"end_time": 1579449599,
					"thumb": "https://i0.hdslb.com/bfs/sycp/tmaterial/201901/c3a0b2320ef185d277a48e7eea3aa671.jpg",
					"hash": "07b7fe1b4742599c27be0e4066100fbc",
					"logo_url": "",
					"logo_hash": "",
					"skip": 0,
					"uri": "",
					"uri_title": "",
					"source": 928,
					"ad_cb": "CAMQAxgDIAAwADigB0IfMTU3NjU1MTAzNjU2MXExNzJhMjJhNTlhMTA4cTQ0MUiR7Y6O8S1SAGgBcACAARayASD9qQJtK1krR/IrPXaJNqEF8V3elXY39S6apZjlGvC8MA==",
					"resource_id": 925,
					"is_ad": false,
					"schema_callup_white_list": [
						"dianping",
						"kaola",
						"vipshop",
						"ctrip",
						"bilibili",
						"mqq",
						"openapp.jdmobile",
						"taobao",
						"tbopen",
						"tmall",
						"weixin",
						"alipays",
						"sinaweibo",
						"pinduoduo",
						"SNKRS",
						"booking",
						"yohobuy",
						"eleme",
						"alipay"
					],
					"extra": {
						"use_ad_web_v2": false,
						"show_urls": [
							
						],
						"click_urls": [
							
						],
						"download_whitelist": [
							
						],
						"open_whitelist": [
							"dianping",
							"kaola",
							"vipshop",
							"ctrip",
							"bilibili",
							"mqq",
							"openapp.jdmobile",
							"taobao",
							"tbopen",
							"tmall",
							"weixin",
							"alipays",
							"sinaweibo",
							"pinduoduo",
							"SNKRS",
							"booking",
							"yohobuy",
							"eleme",
							"alipay"
						],
						"report_time": 0,
						"sales_type": 41,
						"special_industry": false,
						"preload_landingpage": 0,
						"upzone_entrance_type": 0,
						"upzone_entrance_report_id": 0
					}
				},
				{
					"id": 6,
					"type": 0,
					"card_type": 15,
					"duration": 3,
					"begin_time": 1514736000,
					"end_time": 2524579200,
					"thumb": "https://i0.hdslb.com/bfs/sycp/tmaterial/201805/4a694d49ac21864c1e17a23273230d40.png",
					"hash": "00b8bad3974c2be58aca005487138a8b",
					"logo_url": "",
					"logo_hash": "",
					"skip": 0,
					"uri": "",
					"uri_title": "",
					"source": 928,
					"ad_cb": "CAYQBhgGIAAwADigB0IfMTU3NjU1MTAzNjU2MXExNzJhMjJhNTlhMTA4cTQ0MUiR7Y6O8S1SAGgBcACAARayASCBpnrV637CX6QZXhLWRZ5fV5dWeQBsvuryHC6fLhlxZQ==",
					"resource_id": 925,
					"is_ad": false,
					"schema_callup_white_list": [
						"dianping",
						"kaola",
						"vipshop",
						"ctrip",
						"bilibili",
						"mqq",
						"openapp.jdmobile",
						"taobao",
						"tbopen",
						"tmall",
						"weixin",
						"alipays",
						"sinaweibo",
						"pinduoduo",
						"SNKRS",
						"booking",
						"yohobuy",
						"eleme",
						"alipay"
					],
					"extra": {
						"use_ad_web_v2": false,
						"show_urls": [
							
						],
						"click_urls": [
							
						],
						"download_whitelist": [
							
						],
						"open_whitelist": [
							"dianping",
							"kaola",
							"vipshop",
							"ctrip",
							"bilibili",
							"mqq",
							"openapp.jdmobile",
							"taobao",
							"tbopen",
							"tmall",
							"weixin",
							"alipays",
							"sinaweibo",
							"pinduoduo",
							"SNKRS",
							"booking",
							"yohobuy",
							"eleme",
							"alipay"
						],
						"report_time": 0,
						"sales_type": 41,
						"special_industry": false,
						"preload_landingpage": 0,
						"upzone_entrance_type": 0,
						"upzone_entrance_report_id": 0
					}
				},
				{
					"id": 1584,
					"type": 1,
					"card_type": 15,
					"duration": 3,
					"begin_time": 1576512000,
					"end_time": 1576598399,
					"thumb": "https://i0.hdslb.com/bfs/sycp/creative_img/201912/a2405851112a1535a47f786032b4aa2a.jpg",
					"hash": "c51f8dd6471f38f9779a81e1b105c616",
					"logo_url": "",
					"logo_hash": "",
					"skip": 1,
					"uri": "https://clickc.admaster.com.cn/c/a136622,b3848893,c3297,i0,m101,8a2,8b3,0a__OS__,n__MAC__,z__IDFA__,o__OPENUDID__,0d__ANDROIDID__,0c__IMEI__,f__IP__,t__TS__,q__OSVS__,r__TERM__,0i__MUDS__,0h__MUID__,0v__ISOFFLINE__,s__ADWH__,1b__CUSTOMV1__,1a__CUSTOMV2__,h",
					"uri_title": "开启你的奕世界",
					"source": 928,
					"cm_mark": 1,
					"ad_cb": "CLAMELAMGLAMIAAwADigB0IfMTU3NjU1MTAzNjU2MXExNzJhMjJhNTlhMTA4cTQ0MUiR7Y6O8S1SAGgBcACAARayASBO1TzGEmH7204Y+iCzV6SFdAIpPSlIC396XVxjJg0vow==",
					"resource_id": 925,
					"is_ad": true,
					"schema_callup_white_list": [
						"tmall",
						"taobao",
						"ctrip",
						"openapp.jdmobile",
						"newsapp",
						"dianping",
						"vipshop",
						"kaola",
						"weixin",
						"alipays",
						"tbopen",
						"qunaraphone",
						"qunariphone",
						"eleme",
						"airbnb"
					],
					"extra": {
						"use_ad_web_v2": true,
						"show_urls": [
							
						],
						"click_urls": [
							
						],
						"download_whitelist": [
							
						],
						"open_whitelist": [
							"tmall",
							"taobao",
							"ctrip",
							"openapp.jdmobile",
							"newsapp",
							"dianping",
							"vipshop",
							"kaola",
							"weixin",
							"alipays",
							"tbopen",
							"qunaraphone",
							"qunariphone",
							"eleme",
							"airbnb"
						],
						"report_time": 0,
						"sales_type": 41,
						"special_industry": false,
						"preload_landingpage": 0,
						"upzone_entrance_type": 0,
						"upzone_entrance_report_id": 0
					}
				},
				{
					"id": 1587,
					"type": 1,
					"card_type": 15,
					"duration": 3,
					"begin_time": 1576512000,
					"end_time": 1576598399,
					"thumb": "https://i0.hdslb.com/bfs/sycp/creative_img/201912/3d86fb8be3972d8cb638a6e97d8257b8.jpg",
					"hash": "dd31004bcaabb582740314d5153bf784",
					"logo_url": "",
					"logo_hash": "",
					"skip": 1,
					"uri": "https://wt.ictr.cn/t/ad?eid=00101317\u0026sdr=clt\u0026ac=1\u0026iesid=__IESID__\u0026ts=__TS__\u0026term=__TERM__\u0026os=__OS__\u0026ua=__UA__\u0026ip=__IP__\u0026mac=__MAC__\u0026mac1=__MAC1__\u0026imei=__IMEI__\u0026adid=__ANDROIDID__\u0026aaid=__AAID__\u0026idfa=__IDFA__\u0026udid=__OPENUDID__\u0026duid=__DUID__\u0026apn=__ANAME__\u0026apk=__AKEY__\u0026sdv=__SDKVS__\u0026ev=__EVNT__\u0026muds=__MUDS__\u0026muid=__MUID__\u0026lbs=__LBS__\u0026osv=__OSVS__\u0026wf=__WIFI__\u0026wfm=__WIFIBSSID__\u0026wfn=__WIFISSID__\u0026scd=__SCWH__\u0026add=__ADWH__\u0026rqid=__REQUESTID__\u0026rd=https%3a%2f%2fwapact.189.cn%3a9001%2f5Gyuyue%2f5GyuyueAll.html%3fcmpid%3djt-5gss-khd-hezo-bilibili-q4\u0026sgn=__SIGN__",
					"uri_title": "Hello，5G用电信！",
					"source": 928,
					"cm_mark": 1,
					"ad_cb": "CLMMELMMGLMMIAAwADigB0IfMTU3NjU1MTAzNjU2MXExNzJhMjJhNTlhMTA4cTQ0MUiR7Y6O8S1SAGgBcACAARayASCOY187HRHxIrMdAnmVso+sEms9byCW7/Kh6tQSChQkYw==",
					"resource_id": 925,
					"is_ad": true,
					"schema_callup_white_list": [
						"tmall",
						"taobao",
						"ctrip",
						"openapp.jdmobile",
						"newsapp",
						"dianping",
						"vipshop",
						"kaola",
						"weixin",
						"alipays",
						"tbopen",
						"qunaraphone",
						"qunariphone",
						"eleme",
						"airbnb"
					],
					"extra": {
						"use_ad_web_v2": true,
						"show_urls": [
							
						],
						"click_urls": [
							
						],
						"download_whitelist": [
							{
								"size": 0,
								"display_name": "电信营业厅-新人领豪华大礼包",
								"apk_name": "",
								"url": "https://itunes.apple.com/cn/app/zhong-guo-dian-xin-zhang-shang/id513836029",
								"md5": "",
								"icon": "https://i0.hdslb.com/bfs/sycp/app_icon/201911/cf8a6e767696282f939f7efabd2ed8c3.gif",
								"bili_url": ""
							}
						],
						"open_whitelist": [
							"tmall",
							"taobao",
							"ctrip",
							"openapp.jdmobile",
							"newsapp",
							"dianping",
							"vipshop",
							"kaola",
							"weixin",
							"alipays",
							"tbopen",
							"qunaraphone",
							"qunariphone",
							"eleme",
							"airbnb"
						],
						"report_time": 0,
						"sales_type": 41,
						"special_industry": false,
						"preload_landingpage": 0,
						"upzone_entrance_type": 0,
						"upzone_entrance_report_id": 0
					}
				},
				{
					"id": 1590,
					"type": 1,
					"card_type": 15,
					"duration": 3,
					"begin_time": 1576512000,
					"end_time": 1576598399,
					"thumb": "https://i0.hdslb.com/bfs/sycp/creative_img/201912/c4effb4e09a0f730379b9052d41259b6.jpg",
					"hash": "d94afca46fc183274d41562851d7cdaf",
					"logo_url": "",
					"logo_hash": "",
					"skip": 1,
					"uri": "http://e.cn.miaozhen.com/r/k=2148023\u0026p=7V3Km\u0026dx=__IPDX__\u0026rt=2\u0026ns=__IP__\u0026ni=__IESID__\u0026v=__LOC__\u0026xa=__ADPLATFORM__\u0026tr=__REQUESTID__\u0026mo=__OS__\u0026m0=__OPENUDID__\u0026m0a=__DUID__\u0026m1=__ANDROIDID1__\u0026m1a=__ANDROIDID__\u0026m2=__IMEI__\u0026m4=__AAID__\u0026m5=__IDFA__\u0026m6=__MAC1__\u0026m6a=__MAC__\u0026m11=__OAID__\u0026o=https://gxb.mmstat.com/gxb.gif?t=https%3A%2F%2Fequity.tmall.com%2Ftm%3FagentId%3D476419%26_bind%3Dtrue%26bc_fl_src%3Dtmall_market_llb_1_641043%26llbPlatform%3D_pube%26llbOsd%3D1%26mm_unid%3D1_3674707_560501015e6d575e0100055c6d02545e6f0d51050a0c\u0026v=39708f36a80a\u0026di=__IDFA__\u0026dim=__IMEI__\u0026bc_fl_src=tmall_market_llb_1_641043\u0026llbPlatform=_pubu\u0026llbOsd=1\u0026agentId=476419",
					"uri_title": "涂鸦国潮 限量发售",
					"source": 928,
					"cm_mark": 1,
					"ad_cb": "CLYMELYMGLYMIAAwADigB0IfMTU3NjU1MTAzNjU2MXExNzJhMjJhNTlhMTA4cTQ0MUiR7Y6O8S1SAGgBcACAARayASBLVpKPCcK5c41FXgRujHbxwUo28DRPIPOacgAwtwVHrw==",
					"resource_id": 925,
					"is_ad": true,
					"schema_callup_white_list": [
						"tmall",
						"taobao",
						"ctrip",
						"openapp.jdmobile",
						"newsapp",
						"dianping",
						"vipshop",
						"kaola",
						"weixin",
						"alipays",
						"tbopen",
						"qunaraphone",
						"qunariphone",
						"eleme",
						"airbnb"
					],
					"extra": {
						"use_ad_web_v2": true,
						"show_urls": [
							
						],
						"click_urls": [
							
						],
						"download_whitelist": [
							
						],
						"open_whitelist": [
							"tmall",
							"taobao",
							"ctrip",
							"openapp.jdmobile",
							"newsapp",
							"dianping",
							"vipshop",
							"kaola",
							"weixin",
							"alipays",
							"tbopen",
							"qunaraphone",
							"qunariphone",
							"eleme",
							"airbnb"
						],
						"report_time": 0,
						"sales_type": 41,
						"special_industry": false,
						"preload_landingpage": 0,
						"upzone_entrance_type": 0,
						"upzone_entrance_report_id": 0
					}
				},
				{
					"id": 1592,
					"type": 1,
					"card_type": 15,
					"duration": 3,
					"begin_time": 1576512000,
					"end_time": 1576598399,
					"thumb": "https://i0.hdslb.com/bfs/sycp/creative_img/201912/d4a39979ac77213e9ccd01f155c8e1ad.jpg",
					"hash": "b4660d4e275c17e6e8d445675c4ff3ee",
					"logo_url": "",
					"logo_hash": "",
					"skip": 1,
					"uri": "https://www.bilibili.com/blackboard/topic/activity-tvmaniac.html",
					"uri_title": "分享你的影视心头好，我们拿大会员和你换！",
					"source": 928,
					"ad_cb": "CLgMELgMGLgMIAAwADigB0IfMTU3NjU1MTAzNjU2MXExNzJhMjJhNTlhMTA4cTQ0MUiR7Y6O8S1SAGgBcACAARayASDoll/UX19Nx8m0ZPXsMoioz5nFh7FJewf7dcawM0bPKg==",
					"resource_id": 925,
					"is_ad": true,
					"extra": {
						"use_ad_web_v2": false,
						"show_urls": [
							
						],
						"click_urls": [
							
						],
						"download_whitelist": [
							
						],
						"open_whitelist": [
							
						],
						"report_time": 0,
						"sales_type": 41,
						"special_industry": false,
						"preload_landingpage": 0,
						"upzone_entrance_type": 0,
						"upzone_entrance_report_id": 0
					}
				}
			]
		}`)
		res, _, err := d.SplashList(ctx(), mobiApp, device, buvid, birth, adExtra, height, width, build, mid)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestSplashShow(t *testing.T) {
	Convey("get SplashShow all", t, func() {
		var (
			mobiApp, device, buvid, birth, adExtra string
			height, width, build                   int
			mid                                    int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.splashShowURL).Reply(200).JSON(`{
			"code": 0,
			"data": [
				{
					"id": 1592,
					"stime": 1576512000,
					"etime": 1576598399
				}
			]
		}`)
		res, err := d.SplashShow(ctx(), mobiApp, device, buvid, birth, adExtra, height, width, build, mid)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
