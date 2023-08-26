package ad

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

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

func TestADRequest(t *testing.T) {
	Convey("ADRequest", t, func() {
		var (
			mid                                                                             int64
			build                                                                           int
			buvid, resource, ip, country, province, city, network, mobiApp, device, adExtra string
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.adURL).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"data": {
				"request_id": "1569492427482q172a22a50a231q384",
				"ads_control": {
					"has_danmu": 0
				},
				"ads_info": {
					"1890": {
						"1898": {
							"index": 1,
							"is_ad": true,
							"cm_mark": 1,
							"card_index": 3,
							"ad_info": {
								"creative_id": 31242779,
								"creative_type": 2,
								"creative_content": {
									"title": "想学原画插画的进，教你0基础变大神！",
									"description": "绘画-公开课",
									"image_url": "https://i0.hdslb.com/bfs/sycp/creative_img/201909/51bdec660f214a579ba400cccdbf7016.jpg_640x400.jpg",
									"image_md5": "a9067204c0eed95d5385ea78201de205",
									"url": "https://cm.bilibili.com/mgk/page/276134143511162880?buvid=__BUVID__&mid=__MID__&imei=__IMEI__&duid=__DUID__&idfa=__IDFA__&android_id=__ANDROIDID__&os=0&request_id=1569492427482q172a22a50a231q384&source_id=1898&track_id=jlI4vLkvgrXjj0Hbab3d7fVa6bDPnLgCxCelKFniTt60tJKkciGKttA00ttYuepJ-9dyfEV_5_cR7-hREc7_pBm1cjaEZXmJWk6atvTiXgWN6E6bOsGlWk_gPk67bFAq&creative_id=31242779&adtype=CPC",
									"click_url": "",
									"show_url": ""
								},
								"ad_cb": "CPenCBDV/CcYm/TyDiBIKAEwmzQ46g5CHzE1Njk0OTI0Mjc0ODJxMTcyYTIyYTUwYTIzMXEzODRI2sWn6NYtUgBaCeS4iua1t+W4gmIG5Lit5Zu9aAFwAHiAgICA4ASAAQOIAQCSAQoxLjE0LjEyOC4wmgHCBWFsbDpjcGNfY3Jvd2RfdGFyZ2V0LGVjcG06ZGVmYXVsdCxjcGNUYWdGaWx0ZXI6dW5kZWZpbmVkLHBsYXlwYWdlY3RyOmVuYWJsZV9wbGF5cGFnZV9jb250ZXh0LG5vX2FkX2Zsb3dfY29udHJvbDp1bmRlZmluZWQsYnJ1c2hfZHVwbGljYXRlOmRlZmF1bHQscGN0cl9jcG06Y3BtLHBjdHJfdjI6bHJfY29uc3RhbnQsZHluYW1pY19mbG93X2NvbnRyb2w6c3BsaXQgdGhlIGZsb3cgYnkgbWlkLHBjdnI6ZGxkLGZyZXFMaW1pdDpkZWZhdWx0LHNtYWxsQ29uc3VtZVVuaXQ6ZGVmYXVsdCxvdXRlckJlYXRJbm5lcjpkZWZhdWx0LG91dGVyUXVpdDpkZWZhdWx0LGZkc19ydHQ6ZGVmYXVsdCxjcGE6Y3BhX2tleTEsaW5kaXNfVVY6dW5kZWZpbmVkLGhhc2hfaW5kaXNfVVY6dW5kZWZpbmVkLGZlZWRzX3JhbmRvbV9yZXN1bHQ6ZGVmYXVsdCxmbG93X3JhdGlvX2NoZWNrOnIyLGJhc2VfaGFzaF9pbmRpc19VVjp1bmRlZmluZWQsY3RyX3RydW5jYXRpb25fZXhwOnRydW5jYXRpb25fMTIwLGRtcF9hZ2VfZXhwOnVuZGVmaW5lZCxjb2xkX2Jvb3RfZXhwOmRlZmF1bHQsbnRoX2JydXNoX2V2ZW50OmRlZmF1bHQsbG93X3F0eV9jcmVhdGl2ZTpsb3dfcXR5LGZyZXFfc3R5bF9jYXJkOmRlZmF1bHQsZHVwX2FkX2NvbnRyb2w6ZGVmYXVsdCxicnVzaF9hZF9jb250cm9sOmRlZmF1bHQsbmVnYXRpdmVGZWRCYWNrOmRlZmF1bHQscGxhdGZvcm06ZGVmYXVsdF8zMDCgAUioASKyASCD1hX4ipQkTN8oGClKIJMGiSUaTQ6kD89skmiY5T1PqboB+wJodHRwczovL2NtLmJpbGliaWxpLmNvbS9tZ2svcGFnZS8yNzYxMzQxNDM1MTExNjI4ODA/YnV2aWQ9X19CVVZJRF9fJm1pZD1fX01JRF9fJmltZWk9X19JTUVJX18mZHVpZD1fX0RVSURfXyZpZGZhPV9fSURGQV9fJmFuZHJvaWRfaWQ9X19BTkRST0lESURfXyZvcz0wJnJlcXVlc3RfaWQ9MTU2OTQ5MjQyNzQ4MnExNzJhMjJhNTBhMjMxcTM4NCZzb3VyY2VfaWQ9MTg5OCZ0cmFja19pZD1qbEk0dkxrdmdyWGpqMEhiYWIzZDdmVmE2YkRQbkxnQ3hDZWxLRm5pVHQ2MHRKS2tjaUdLdHRBMDB0dFl1ZXBKLTlkeWZFVl81X2NSNy1oUkVjN19wQm0xY2phRVpYbUpXazZhdHZUaVhnV042RTZiT3NHbFdrX2dQazY3YkZBcSZjcmVhdGl2ZV9pZD0zMTI0Mjc3OSZhZHR5cGU9Q1BDwgG9ATQyMDdfMTEyOV83NjRfNzM3XzYzMF81OTVfNTc2XzUyNF81MjRfMzg0XzM4NF8zODRfMzc4XzM0Nl8yNTlfMjU5XzI1OF8yNDJfMjQyXzI0Ml8yNDJfMjQyXzI0Ml8yNDBfMjQwXzI0MF8yNDBfMjQwXzI0MF8yNDBfMjQwXzI0MF8yNDBfMjQwXzI0MF8yNDBfMjQwXzI0MF8yNDBfMjQwXzI0MF8xODZfMTg2XzE4NF8xODNfMTgzXzE4M9IBANgBJuABoMIe6AGgjQbwAQD4AUiAAgKIAgCSAjUzNDIzNTJfMTU2ODg5MDM5Miw0MjE4MTZfMTU2ODk5ODEwNCw1NjA5NDVfMTU2ODk5ODExNJgCyhWgAgOoAvsSsAICuAIAwAIAyAK4BdACANgCAOoCAPACAPgCAIgDAZIDAKgDALADALgDAMIDAMgDAdIDN3siMSI6IjMxMjQyNzc5IiwiMiI6IjY2ODMiLCI0IjoiMzgiLCI1IjoiNDciLCI2IjoiM18wIn0=",
								"card_type": 7,
								"extra": {
									"use_ad_web_v2": true,
									"show_urls": [],
									"click_urls": [],
									"download_whitelist": [],
									"open_whitelist": [],
									"card": {
										"card_type": 7,
										"title": "想学原画插画的进，教你0基础变大神！",
										"covers": [{
											"url": "https://i0.hdslb.com/bfs/sycp/creative_img/201909/51bdec660f214a579ba400cccdbf7016.jpg_640x400.jpg",
											"loop": 0,
											"image_height": 400,
											"image_width": 640
										}],
										"jump_url": "https://cm.bilibili.com/mgk/page/276134143511162880?buvid=__BUVID__&mid=__MID__&imei=__IMEI__&duid=__DUID__&idfa=__IDFA__&android_id=__ANDROIDID__&os=0&request_id=1569492427482q172a22a50a231q384&source_id=1898&track_id=jlI4vLkvgrXjj0Hbab3d7fVa6bDPnLgCxCelKFniTt60tJKkciGKttA00ttYuepJ-9dyfEV_5_cR7-hREc7_pBm1cjaEZXmJWk6atvTiXgWN6E6bOsGlWk_gPk67bFAq&creative_id=31242779&adtype=CPC",
										"desc": "绘画-公开课",
										"callup_url": "",
										"long_desc": "绘画-公开课",
										"ad_tag": "",
										"extra_desc": "绘画-公开课",
										"ad_tag_style": {
											"type": 2,
											"text": "广告",
											"text_color": "#999999FF",
											"bg_border_color": "#999999FF"
										},
										"feedback_panel": {
											"panel_type_text": "广告",
											"feedback_panel_detail": [{
												"text": "我不想看到这个广告",
												"module_id": 1,
												"jump_type": 1,
												"icon_url": "https://i0.hdslb.com/bfs/sycp/mng/201907/a53df8f189bb12666a39d10ad1babcf5.png",
												"jump_url": "",
												"secondary_panel": [{
													"text": "不感兴趣",
													"reason_id": 1
												},
												{
													"text": "相似内容过多",
													"reason_id": 2
												},
												{
													"text": "广告质量差",
													"reason_id": 5
												}]
											},
											{
												"text": "举报",
												"module_id": 2,
												"jump_type": 2,
												"icon_url": "https://i0.hdslb.com/bfs/sycp/mng/201907/2bc344ad3510da5cfdc7c7714abaeda4.png",
												"jump_url": "http://cm.bilibili.com/ldad/light/ad-complain.html",
												"secondary_panel": []
											},
											{
												"text": "我为什么会看到此广告",
												"module_id": 3,
												"jump_type": 2,
												"icon_url": "https://i0.hdslb.com/bfs/sycp/mng/201907/82480c4ef205c9b715d6e2ea7f5c4041.png",
												"jump_url": "http://cm.bilibili.com/ldad/light/ad-introduce.html",
												"secondary_panel": []
											}]
										}
									},
									"report_time": 2000,
									"sales_type": 12,
									"special_industry": false,
									"special_industry_tips": "",
									"preload_landingpage": 0,
									"enable_download_dialog": false,
									"enable_share": true,
									"share_info": {
										"title": "",
										"subtitle": "",
										"image_url": ""
									}
								},
								"creative_style": 1
							}
						},
						"1899": {
							"index": 2,
							"is_ad": true,
							"cm_mark": 1,
							"card_index": 7,
							"ad_info": {
								"creative_id": 279999866739535872,
								"creative_type": 2,
								"creative_content": {
									"title": "想尝试全新体验？那就成为我的玩具吧？",
									"description": "前往参与测试",
									"image_url": "https://i0.hdslb.com/bfs/sycp/creative_img/201909/76a4dcbd183b66de22be07e510472bce.jpg_640x400.jpg",
									"image_md5": "070e98de9dc61ae6716af611631fac80",
									"url": "bilibili://game_center/detail?id=102378&sourceFrom=883&sourceType=adPut&msource=1&source=afid_a1f04180e03011e9bd7a261c4f8f6e99",
									"click_url": "https://ad-bili-data.biligame.com/api/mobile/clickBili?ad_plan_id=44996&mid=__MID__&os=0&idfa=__IDFA__&buvid=__BUVID__&android_id=__ANDROIDID__&imei=__IMEI__&mac=__MAC__&duid=__DUID__&ip=1.14.128.0&request_id=1569492427482q172a22a50a231q384&ts=__TS__&ua=__UA__",
									"show_url": ""
								},
								"ad_cb": "CAAQABiA4KqR08yw8QMgoB8oATAkOOsOQh8xNTY5NDkyNDI3NDgycTE3MmEyMmE1MGEyMzFxMzg0SNrFp+jWLVIAWgnkuIrmtbfluIJiBuS4reWbvWgBcAB4gICAgOAEgAECiAGSRJIBCjEuMTQuMTI4LjCaAb4FYWxsOmNwY19jcm93ZF90YXJnZXQsZWNwbTpkZWZhdWx0LGNwY1RhZ0ZpbHRlcjp1bmRlZmluZWQscGxheXBhZ2VjdHI6ZW5hYmxlX3BsYXlwYWdlX2NvbnRleHQsbm9fYWRfZmxvd19jb250cm9sOnVuZGVmaW5lZCxicnVzaF9kdXBsaWNhdGU6ZGVmYXVsdCxwY3RyX2NwbTpjcG0scGN0cl92Mjpscl9jb25zdGFudCxkeW5hbWljX2Zsb3dfY29udHJvbDpzcGxpdCB0aGUgZmxvdyBieSBtaWQscGN2cjpkbGQsZnJlcUxpbWl0OmRlZmF1bHQsc21hbGxDb25zdW1lVW5pdDpkZWZhdWx0LG91dGVyQmVhdElubmVyOmRlZmF1bHQsb3V0ZXJRdWl0OmRlZmF1bHQsZmRzX3J0dDpkZWZhdWx0LGNwYTpjcGFfa2V5MSxpbmRpc19VVjp1bmRlZmluZWQsaGFzaF9pbmRpc19VVjp1bmRlZmluZWQsZmVlZHNfcmFuZG9tX3Jlc3VsdDpkZWZhdWx0LGZsb3dfcmF0aW9fY2hlY2s6cjIsYmFzZV9oYXNoX2luZGlzX1VWOnVuZGVmaW5lZCxjdHJfdHJ1bmNhdGlvbl9leHA6dHJ1bmNhdGlvbl8xMjAsZG1wX2FnZV9leHA6dW5kZWZpbmVkLGNvbGRfYm9vdF9leHA6ZGVmYXVsdCxudGhfYnJ1c2hfZXZlbnQ6ZGVmYXVsdCxsb3dfcXR5X2NyZWF0aXZlOmxvd19xdHksZnJlcV9zdHlsX2NhcmQ6ZGVmYXVsdCxkdXBfYWRfY29udHJvbDpkZWZhdWx0LGJydXNoX2FkX2NvbnRyb2w6ZGVmYXVsdCxuZWdhdGl2ZUZlZEJhY2s6ZGVmYXVsdCxwbGF0Zm9ybTpkZWZhdWx0oAGgH6gBALIBIMV2b3nrSekk+lS/o68diq7yaEnALQhxvKKGxInB3FghugF+YmlsaWJpbGk6Ly9nYW1lX2NlbnRlci9kZXRhaWw/aWQ9MTAyMzc4JnNvdXJjZUZyb209ODgzJnNvdXJjZVR5cGU9YWRQdXQmbXNvdXJjZT0xJnNvdXJjZT1hZmlkX2ExZjA0MTgwZTAzMDExZTliZDdhMjYxYzRmOGY2ZTk5wgEA0gEA2AEm4AEA6AEA8AEA+AEAgAICiAIAuAIAwAIAyAIA0AIA2AIA6gIA8ALKrwT4AgCIAwGSAwCoAwCwAwC4AwDCAwDIAwHSA0F7IjEiOiIyNzk5OTk4NjY3Mzk1MzU4NzIiLCIyIjoiMzYiLCI0IjoiOTAiLCI1IjoiMjkyIiwiNiI6IjM2XzAifQ==",
								"card_type": 3,
								"extra": {
									"use_ad_web_v2": false,
									"show_urls": [],
									"click_urls": ["https://ad-bili-data.biligame.com/api/mobile/clickBili?ad_plan_id=44996&mid=__MID__&os=0&idfa=__IDFA__&buvid=__BUVID__&android_id=__ANDROIDID__&imei=__IMEI__&mac=__MAC__&duid=__DUID__&ip=1.14.128.0&request_id=1569492427482q172a22a50a231q384&ts=__TS__&ua=__UA__"],
									"download_whitelist": [],
									"open_whitelist": [],
									"card": {
										"card_type": 3,
										"title": "想尝试全新体验？那就成为我的玩具吧？",
										"covers": [{
											"url": "https://i0.hdslb.com/bfs/sycp/creative_img/201909/76a4dcbd183b66de22be07e510472bce.jpg_640x400.jpg",
											"loop": 0,
											"image_height": 0,
											"image_width": 0
										}],
										"jump_url": "bilibili://game_center/detail?id=102378&sourceFrom=883&sourceType=adPut&msource=1&source=afid_a1f04180e03011e9bd7a261c4f8f6e99",
										"desc": "前往参与测试",
										"callup_url": "",
										"long_desc": "",
										"ad_tag": "",
										"extra_desc": "",
										"ad_tag_style": {
											"type": 2,
											"text": "广告",
											"text_color": "#999999FF",
											"bg_border_color": "#999999FF"
										},
										"feedback_panel": {
											"panel_type_text": "广告",
											"feedback_panel_detail": [{
												"text": "我不想看到这个广告",
												"module_id": 1,
												"jump_type": 1,
												"icon_url": "https://i0.hdslb.com/bfs/sycp/mng/201907/a53df8f189bb12666a39d10ad1babcf5.png",
												"jump_url": "",
												"secondary_panel": [{
													"text": "不感兴趣",
													"reason_id": 1
												},
												{
													"text": "相似内容过多",
													"reason_id": 2
												},
												{
													"text": "广告质量差",
													"reason_id": 5
												}]
											},
											{
												"text": "举报",
												"module_id": 2,
												"jump_type": 2,
												"icon_url": "https://i0.hdslb.com/bfs/sycp/mng/201907/2bc344ad3510da5cfdc7c7714abaeda4.png",
												"jump_url": "http://cm.bilibili.com/ldad/light/ad-complain.html",
												"secondary_panel": []
											},
											{
												"text": "我为什么会看到此广告",
												"module_id": 3,
												"jump_type": 2,
												"icon_url": "https://i0.hdslb.com/bfs/sycp/mng/201907/82480c4ef205c9b715d6e2ea7f5c4041.png",
												"jump_url": "http://cm.bilibili.com/ldad/light/ad-introduce.html",
												"secondary_panel": []
											}]
										}
									},
									"report_time": 2000,
									"sales_type": 21,
									"special_industry": false,
									"special_industry_tips": "",
									"preload_landingpage": 0,
									"share_info": {
	
									}
								},
								"creative_style": 1
							}
						}
					}
				}
			}
		}`)
		res, err := d.ADRequest(ctx(), mid, build, buvid, resource, ip, country, province, city, network, mobiApp, device, adExtra)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
