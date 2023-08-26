package dynamic

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
func TestDynamicHistory(t *testing.T) {
	Convey("DynamicHistory", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.host+_dynamicHistory).Reply(200).JSON(`{
			"code":0,
			"message":"ok",
			"data":{
				"has_more":1,
				"open_rcmd":1,
				"attentions":{
					"uids":[111,222,33,44],
					"bangumis":[{"type":1, "season_id":222},{"type":4, "season_id":3322}]
				},
				"next_offset":1212121212,
				"cards":[
					{
						"card":"业务方数据，服务端透传给client",
						"desc":{ },
						"extension": { }
					}
				],
				"folds":[
					{
						"dynamic_ids":[ 111, 2333, 3532 ]
					}
				],
				"fold_mgr":[
					{
						"fold_type":1,
						"folds":[{"dynamic_ids":[1,2,3]},{"dynamic_ids":[4,5]}]
					},
					{
						"fold_type":2,
						"folds":[{"dynamic_ids":[7,8]},{"dynamic_ids":[11,34]}]
					}
				],
				"inplace_fold":[
					{
						"statement":"3条动态被折叠",
						"dynamic_ids":[1,2,3]
					},
					{
						"statement":"1条动态被折叠",
						"dynamic_ids":[6]
					}
				]
			}
		}`)
		var param string
		res, err := d.DynamicHistory(ctx(), param)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestDynamicCount(t *testing.T) {
	Convey("DynamicCount", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.host+_dynamicHistory).Reply(200).JSON(`{
			"code": 0,
			"message": "ok",
			"data": {
			  "new_num": 2,
			  "exist_gap": 1,
			  "update_num": 3,
			  "open_rcmd": 1,
			  "archive_up_num": 3,
			  "max_dynamic_id": 1212122112,
			  "history_offset": 1212122112,
			  "up_num": {
				"archive_up_num": 3,
				"bangumi_up_num": 1
			  },
			  "attentions": {
				"uids": [
				  111,
				  222
				],
				"bangumis": [
				  {
					"type": 1,
					"season_id": 222
				  },
				  {
					"type": 4,
					"season_id": 3322
				  }
				]
			  },
			  "rcmd_cards": [
				{
				  "trackid": 1211211,
				  "type": 2,
				  "pos": 2,
				  "users": [
					{
					  "basic_profile": {},
					  "feed": {
						"fans_cnt": 1
					  },
					  "recommend": {
						"rec_reason": "游戏区热门up主",
						"tid": 4,
						"sub_tid": 121
					  }
					}
				  ]
				}
			  ],
			  "cards": [
				{
				  "card": "业务方数据，服务端透传给client",
				  "desc": {},
				  "extension": {}
				}
			  ],
			  "folds": [
				{
				  "dynamic_ids": [
					111,
					2333,
					3532
				  ]
				}
			  ],
			  "fold_mgr": [
				{
				  "fold_type": 1,
				  "folds": [
					{
					  "dynamic_ids": [
						1,
						2,
						3
					  ]
					},
					{
					  "dynamic_ids": [
						4,
						5
					  ]
					}
				  ]
				},
				{
				  "fold_type": 2,
				  "folds": [
					{
					  "dynamic_ids": [
						7,
						8
					  ]
					},
					{
					  "dynamic_ids": [
						11,
						14
					  ]
					}
				  ]
				},
				{
				  "fold_type": 3,
				  "folds": [
					{
					  "dynamic_ids": [
						10,
						18
					  ]
					},
					{
					  "dynamic_ids": [
						13,
						19
					  ]
					}
				  ]
				},
				{
				  "fold_type": 4,
				  "folds": [
					{
					  "dynamic_ids": [
						22,
						23
					  ]
					},
					{
					  "dynamic_ids": [
						24,
						25
					  ]
					}
				  ]
				}
			  ],
			  "inplace_fold": [
				{
				  "statement": "3条动态被折叠",
				  "dynamic_ids": [
					1,
					2,
					3
				  ]
				},
				{
				  "statement": "1条动态被折叠",
				  "dynamic_ids": [
					6
				  ]
				}
			  ]
			}
		  }`)
		var param string
		res, err := d.DynamicCount(ctx(), param)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestDynamicNew(t *testing.T) {
	Convey("DynamicNew", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.host+_dynamicHistory).Reply(200).JSON(`{
			"code": 0,
			"message": "ok",
			"data": {
			  "new_num": 2,
			  "exist_gap": 1,
			  "update_num": 3,
			  "open_rcmd": 1,
			  "archive_up_num": 3,
			  "max_dynamic_id": 1212122112,
			  "history_offset": 1212122112,
			  "up_num": {
				"archive_up_num": 3,
				"bangumi_up_num": 1
			  },
			  "attentions": {
				"uids": [
				  111,
				  222
				],
				"bangumis": [
				  {
					"type": 1,
					"season_id": 222
				  },
				  {
					"type": 4,
					"season_id": 3322
				  }
				]
			  },
			  "rcmd_cards": [
				{
				  "trackid": 1211211,
				  "type": 2,
				  "pos": 2,
				  "users": [
					{
					  "basic_profile": {},
					  "feed": {
						"fans_cnt": 1
					  },
					  "recommend": {
						"rec_reason": "游戏区热门up主",
						"tid": 4,
						"sub_tid": 121
					  }
					}
				  ]
				}
			  ],
			  "cards": [
				{
				  "card": "业务方数据，服务端透传给client",
				  "desc": {},
				  "extension": {}
				}
			  ],
			  "folds": [
				{
				  "dynamic_ids": [
					111,
					2333,
					3532
				  ]
				}
			  ],
			  "fold_mgr": [
				{
				  "fold_type": 1,
				  "folds": [
					{
					  "dynamic_ids": [
						1,
						2,
						3
					  ]
					},
					{
					  "dynamic_ids": [
						4,
						5
					  ]
					}
				  ]
				},
				{
				  "fold_type": 2,
				  "folds": [
					{
					  "dynamic_ids": [
						7,
						8
					  ]
					},
					{
					  "dynamic_ids": [
						11,
						14
					  ]
					}
				  ]
				},
				{
				  "fold_type": 3,
				  "folds": [
					{
					  "dynamic_ids": [
						10,
						18
					  ]
					},
					{
					  "dynamic_ids": [
						13,
						19
					  ]
					}
				  ]
				},
				{
				  "fold_type": 4,
				  "folds": [
					{
					  "dynamic_ids": [
						22,
						23
					  ]
					},
					{
					  "dynamic_ids": [
						24,
						25
					  ]
					}
				  ]
				}
			  ],
			  "inplace_fold": [
				{
				  "statement": "3条动态被折叠",
				  "dynamic_ids": [
					1,
					2,
					3
				  ]
				},
				{
				  "statement": "1条动态被折叠",
				  "dynamic_ids": [
					6
				  ]
				}
			  ]
			}
		  }`)
		var param string
		res, err := d.DynamicNew(ctx(), param)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
