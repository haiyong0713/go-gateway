package bplus

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

func TestDynamicDetail(t *testing.T) {
	Convey("DynamicDetail", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.dynamicDetail).Reply(200).JSON(`{
			"code": 0,
			"message": "ok",
			"data": {
			  "list": [
				{
				  "dynamic_id": 1,
				  "publish_time": 12121212,
				  "audit_status": 0,
				  "delete_status": 0,
				  "mid": 121212,
				  "nick_name" : "nickname",
				  "face_img" : "https://1111.jpg",
				  "rid_type": 2,
				  "rid": 121212,
				  "view_count": 111,
				  "comment_count": 12,
				  "like_count":1212,
				  "dynamic_text": "dongtai动态e动态文案n",
				  "topics": [
					"topic1",
					"topic2"
				  ],
				  "img_count": 9,
				  "imgs": [
					"http://bfs.jpg",
					"http://bfs2.jpg"
				  ],
				  "jump_url":""
				},
				{
				  "dynamic_id": 2,
				  "publish_time": 12121212,
				  "publish_time_text": "12小时前",
				  "mid": 121211,
				  "nick_name" : "nickname",
				  "face_img" : "https://1111.jpg",
				  "rid_type": 2,
				  "audit_status": 0,
				  "delete_status": 0,
				  "rid": 121212,
				  "view_count": 111,
				  "comment_count": 12,
				  "dynamic_text": "dongtai动态e动态文案n",
				  "topics": [
							  "李光洙",
							  "小豆芽",
							  "光头",
							  "测试"
						  ],
				   "topic_infos": [
						 {
							  "topic_id": 12772,
							  "topic_name": "李光洙",
							  "is_activity":0,
							  "topic_link":"xxxxxxxxx"
						 },
						 {
							  "topic_id": 13527,
							  "topic_name": "小豆芽",
							  "is_activity":1,
							  "topic_link":"xxxxxx"
						  },
						  {
							  "topic_id": 914,
							  "topic_name": "光头",
							  "is_activity":0,
							  "topic_link":"xxxxxx"
						  },
						  {
							  "topic_id": 600,
							  "topic_name": "测试",
							  "is_activity":0,
							  "topic_link":"xxxxxx"
						  }
				   ]
				  "img_count": 3,
				  "imgs": [
					"http://bfs.jpg",
					"http://bfs2.jpg"
				  ],
				  "jump_url":"https://xxxxxx"
				}
			  ]
			}
		  }`)
		var (
			platfrom, mobiApp, device string
			build                     int
			ids                       []int64
		)
		res, err := d.DynamicDetail(ctx(), platfrom, mobiApp, device, build, ids...)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
