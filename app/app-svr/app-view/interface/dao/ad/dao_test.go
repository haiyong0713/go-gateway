package ad

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/goconvey/convey"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-view")
		flag.Set("conf_token", "3a4CNLBhdFbRQPs7B4QftGvXHtJo92xw")
		flag.Set("tree_id", "4575")
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
	m.Run()
	os.Exit(0)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestDao_Ad(t *testing.T) {
	convey.Convey("get TestDao_Ad", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.adURL).Reply(200).JSON(`{
			"code": 0,
			"message": "success",
			"data": {
				"request_id": "1551241997659q172a22a51a130q160",
				"ads_control": {
					"has_danmu": 0
				},
				"ads_info": {
					"2029": {
						"2333": {
							"index": 2,
							"is_ad": false,
							"cm_mark": 0,
							"card_index": 4
						},
						"2030": {
							"index": 1,
							"is_ad": true,
							"cm_mark": 1,
							"card_index": 1,
							"ad_info": {
								"creative_id": 7731958,
								"creative_type": 0,
								"creative_content": {
									"title": "天呐~这是什么神仙少女鞋呀！",
									"description": "送女友真的太合适了",
									"image_url": "https://i0.hdslb.com/bfs/sycp/creative_img/201902/919f4e10edf5941bc70b52ff126aad38.jpg",
									"image_md5": "1b640edd8195042e3852cfde79d304e7",
									"url": "https://itunes.apple.com/cn/app/id490655927?mt=8",
									"click_url": "",
									"show_url": "",
									"thumbnail_url": "",
									"thumbnail_url_md5": ""
								},
								"ad_cb": "COC9BBDtyAwY9vXXAyAUKAEwkAQ47g9CHzE1NTEyNDE5OTc2NTlxMTcyYTIyYTUxYTEzMHExNjBI29rp6ZItUgBaAGIAaAFwAXiAgICA4AKAAQOIAQCSAQCaAZYDYWxsOmRlZmF1bHQsZWNwbTpkZWZhdWx0LGNwY1RhZ0ZpbHRlcjp1bmRlZmluZWQsZW5oYW5jZUN0clFGYWN0b3I6c3F1YXJlLGFkTWVjaGFuaXNtTW9uaXRvcjpvdGhlcixwbGF5cGFnZWN0cjpkaXNhYmxlLHVwX3JlY19mbG93X2NvbnRyb2w6dW5kZWZpbmVkLGJydXNoX2R1cGxpY2F0ZTpkZWZhdWx0LHBjdHJfY3BtOmNwbSxkZnhfc3BlY2lmaWNfcmF0aW86dW5kZWZpbmVkLHBjdHJfdjI6bGFyYWNyb2Z0LGR5bmFtaWNfZmxvd19jb250cm9sOnNwbGl0IHRoZSBmbG93IGJ5IG1pZCxwY3ZyOmJvdGhfYV8xX2JfMC4wNV9jXzFfZl8xXzEuNSxmcmVxTGltaXQ6ZGVmYXVsdCxzbWFsbENvbnN1bWVVbml0OmRlZmF1bHQsb3V0ZXJCZWF0SW5uZXI6ZW5hYmxlLG91dGVyUXVpdDpkZWZhdWx0LGZkc19ydHQ6ZGVmYXVsdKABFKgBHLIBIAQdkesUalJP6jMSIyBjM364d/3ffHuTZmgnnmtrzrTSugEwaHR0cHM6Ly9pdHVuZXMuYXBwbGUuY29tL2NuL2FwcC9pZDQ5MDY1NTkyNz9tdD04wgFkNjk1XzIxMl8xNDlfMTQ2XzY5XzE5XzE3XzE3XzE3XzE3XzE3XzE3XzE3XzE3XzE3XzE1XzE1XzE1XzE1XzE1XzE1XzE1XzE0XzE0XzE0XzE0XzE0XzE0XzE0XzE0XzE0XzFfMcoBANIBANgBFuABgIl66AGAiXrwAc7xrxX4ATKAAgSIAgCSAucCMjEwMzE3XzE1NTExOTY1MjYsMjA2MTIyXzE1NTEyMjY5MjAsMjA3MjA5XzE1NTEyMjgzNTgsMjA3ODY0XzE1NTEyMjgzNTgsMjEwMzI1XzE1NTEyMjgzNTgsMjAxNDg0XzE1NTEyMzAwODYsMjA1MTAwXzE1NTEyMzI4MjYsMjA5OTAwXzE1NTEyMzM2MDYsMjA0NjY2XzE1NTEyMzQwMjIsMjEwMjY0XzE1NTEyMzQwMjIsMjA3MjA3XzE1NTEyMzQwMjIsMjA5ODkzXzE1NTEyMzQwMjIsMjA0NjcwXzE1NTEyMzQ2ODgsMjA1NTA4XzE1NTEyMzQ2ODgsMjA0NjU4XzE1NTEyMzU0MTYsMjA1NDQxXzE1NTEyMzU0MTYsMjEwMzQxXzE1NTEyNDE3NjcsMjExMDA2XzE1NTEyNDE3NjcsMjAxODE1XzE1NTEyNDE3NzAsMjA4MTk2XzE1NTEyNDE3NzCYAorqsQOgAuifAagC0LYisALDDLgCAMACAMgCANACANgCAOICAiws",
								"card_type": 5,
								"extra": {
									"use_ad_web_v2": true,
									"show_urls": [],
									"click_urls": [],
									"download_whitelist": [{
										"size": 128531456,
										"display_name": "Yoho!Buy 有货——潮流购物逛不停",
										"apk_name": "com.yoho.buy",
										"url": "https://itunes.apple.com/cn/app/id490655927?mt=8",
										"md5": "",
										"icon": "https://i0.hdslb.com/bfs/sycp/app_icon/201801/88350209586ac1c51191fe188f6ee2fa.gif",
										"bili_url": ""
									}],
									"open_whitelist": [],
									"card": {
										"card_type": 5,
										"title": "天呐~这是什么神仙少女鞋呀！",
										"covers": [{
											"url": "https://i0.hdslb.com/bfs/sycp/creative_img/201902/919f4e10edf5941bc70b52ff126aad38.jpg"
										}],
										"jump_url": "https://itunes.apple.com/cn/app/id490655927?mt=8",
										"desc": "送女友真的太合适了",
										"button": {
											"type": 2,
											"text": "详情",
											"jump_url": "https://itunes.apple.com/cn/app/id490655927?mt=8",
											"report_urls": [],
											"dlsuc_callup_url": ""
										},
										"callup_url": "",
										"ad_tag": ""
									},
									"report_time": 2000,
									"appstore_priority": 1,
									"sales_type": 12,
									"special_industry": false,
									"special_industry_tips": ""
								}
							}
						}
					}
				}
			}
		}`)
		res, err := d.Ad(ctx(), "iphone", "phone", "12312", 111, 111, 2222, 1, 1, []int64{1}, []int64{1}, "4g", "", "tm.recommend.0.0", "tm.recommend.0.0", "3")
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldNotBeEmpty)
	})
}

func ctx() context.Context {
	return context.Background()
}
