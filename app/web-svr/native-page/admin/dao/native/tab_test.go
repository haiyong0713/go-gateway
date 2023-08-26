package native

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/native-page/admin/model/native"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_GetTabById(t *testing.T) {
	convey.Convey("GetTabById", t, func(conveyCtx convey.C) {
		var (
			c        = context.Background()
			id int32 = 1
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			_, err := d.GetTabById(c, id)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SearchTab(t *testing.T) {
	convey.Convey("SearchTab", t, func(conveyCtx convey.C) {
		var (
			c   = context.Background()
			req = &native.SearchTabReq{
				ID:    1,
				State: -1,
				Pn:    1,
				Ps:    5,
			}
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			_, err := d.SearchTab(c, req)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CreateTab(t *testing.T) {
	convey.Convey("CreateTab", t, func(conveyCtx convey.C) {
		var (
			c   = context.Background()
			tab = &native.Tab{
				TabData: native.TabData{
					Title:         "活动底栏",
					Stime:         1586310000,
					Etime:         1586312000,
					State:         native.TabStateValid,
					Operator:      "unit_test",
					BgType:        1,
					BgImg:         "http://www.baidu.com?bg_img",
					BgColor:       "",
					IconType:      native.IconTypeWord,
					ActiveColor:   "#111222",
					InactiveColor: "#444555",
				},
				Creator: "unit_test",
			}
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			_, err := d.CreateTab(c, nil, tab)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_GetTabModuleByTabIds(t *testing.T) {
	convey.Convey("GetTabModuleByTabIds", t, func(conveyCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int32{1, 2}
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			_, err := d.GetTabModuleByTabIds(c, ids)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CreateTabModule(t *testing.T) {
	convey.Convey("CreateTabModule", t, func(conveyCtx convey.C) {
		var (
			c         = context.Background()
			tabModule = &native.TabModule{
				TabModuleData: native.TabModuleData{
					Title:       "广场",
					TabId:       1,
					State:       native.TabModuleStateValid,
					Operator:    "unit_test",
					ActiveImg:   "http://www.baidu.com/active_img",
					InactiveImg: "http://www.baidu.com/inactive_img",
					Category:    native.CategoryPage,
					Pid:         1,
					Url:         "",
					Rank:        1,
				},
			}
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			_, err := d.CreateTabModule(c, nil, tabModule)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_UpdateTabModulesById(t *testing.T) {
	convey.Convey("UpdateTabModulesById", t, func(conveyCtx convey.C) {
		var (
			c                  = context.Background()
			id           int32 = 1
			tabModuleMap       = map[string]interface{}{
				"title":        "我的",
				"operator":     "unit_test_modify",
				"state":        native.TabModuleStateValid,
				"active_img":   "http://www.baidu.com?active_img_modify",
				"inactive_img": "http://www.baidu.com?inactive_img_modify",
				"category":     native.CategoryLink,
				"pid":          0,
				"url":          "http://www.baidu.com?category_url_modify",
				"rank":         2,
			}
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			err := d.UpdateTabModulesById(c, nil, id, tabModuleMap)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_UpdateTabModulesByTabId(t *testing.T) {
	convey.Convey("UpdateTabModulesById", t, func(conveyCtx convey.C) {
		var (
			c                  = context.Background()
			tabId        int32 = 1
			tabModuleMap       = map[string]interface{}{
				"title":        "我的",
				"operator":     "unit_test_modify",
				"state":        native.TabModuleStateValid,
				"active_img":   "http://www.baidu.com?active_img_modify",
				"inactive_img": "http://www.baidu.com?inactive_img_modify",
				"category":     native.CategoryLink,
				"pid":          0,
				"url":          "http://www.baidu.com?category_url_modify",
				"rank":         2,
			}
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			err := d.UpdateTabModulesByTabId(c, nil, tabId, tabModuleMap)
			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_GetTabModuleByPid(t *testing.T) {
	convey.Convey("GetTabModuleByPid", t, func(conveyCtx convey.C) {
		var (
			c         = context.Background()
			pid int32 = 10
		)
		conveyCtx.Convey("When everything goes positive", func(conveyCtx convey.C) {
			tabModule, err := d.GetTabModuleByPid(c, pid)

			conveyCtx.Convey("Then err should be nil.", func(conveyCtx convey.C) {
				fmt.Printf("%v\n", tabModule)
				conveyCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
