package native

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeAddPage(t *testing.T) {
	convey.Convey("AddPage", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			natPage = &natmdl.PageParam{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.AddPage(c, natPage)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeModifyPage(t *testing.T) {
	convey.Convey("AddPage", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.ModifyPage(c, 1, map[string]interface{}{})
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelPage(t *testing.T) {
	convey.Convey("DelPage", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			id   = int64(1)
			user = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelPage(c, id, user, "system")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeUpdatePage(t *testing.T) {
	convey.Convey("UpdatePage", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			natPage = &natmdl.EditParam{
				ID:         1,
				Stime:      1556121600,
				ShareTitle: "",
				ShareImage: "",
				UserName:   "lx",
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.UpdatePage(c, natPage, 0)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDynExtByPid(t *testing.T) {
	convey.Convey("DynExtByPid", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			rly, _ := d.DynExtByPid(c, 100)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				fmt.Printf("%v", rly)
			})
		})
	})
}

func TestNativePageSkipUrl(t *testing.T) {
	convey.Convey("PageSkipUrl", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			param = &natmdl.EditParam{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.PageSkipUrl(c, param, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativePageByID(t *testing.T) {
	convey.Convey("PageByID", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.PageByID(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativePageByFID(t *testing.T) {
	convey.Convey("PageByFID", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			foreignID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			pageRes, err := d.PageByFID(c, foreignID, 1)
			convCtx.Convey("Then err should be nil.pageRes should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(pageRes, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestModulesInfo(t *testing.T) {
	convey.Convey("modulesInfo", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			nativeID = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			moduleID, err := d.ModulesInfo(c, nativeID, []int{0})
			convCtx.Convey("Then err should be nil.moduleID should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", moduleID)
			})
		})
	})
}

func TestNativeSearchPage(t *testing.T) {
	convey.Convey("SearchPage", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			param = &natmdl.SearchParam{PageParam: natmdl.PageParam{ID: 4}, Ps: 5, Pn: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.SearchPage(c, param)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativeNatTagID(t *testing.T) {
	convey.Convey("NatTagID", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			title = "光头"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			tagID, err := d.NatTagID(c, title)
			convCtx.Convey("Then err should be nil.tagID should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(tagID)
			})
		})
	})
}

func TestNativeSaveModule(t *testing.T) {
	convey.Convey("SaveModule", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			nativeID = int64(1)
			portion  = &natmdl.JsonData{}
			str      = `{
  "structure": {
    "root": {
      "id": "root",
      "children": ["a", "b", "c", "i"]
    }
  },
  "modules": {
    "a": {
      "type": "click",
      "config": {
        "image": "http://i0.hdslb.com/bfs/article/833520114bd7f710b9568f932e31cd263b6461bf.jpg",
        "width": 800,
        "height": 800,
        "areas": [
          {
            "x": 20,
            "y": 20,
            "w": 100,
            "h": 100,
            "link": "http://www.bilibili.com/read/cv2241?from=category_1"
          }
        ]
      }
    },
    "b": {
      "type": "act",
      "config": {
        "title_imgae": "http://i0.hdslb.com/bfs/article/f70ea1a4348f39de935eb290a8d9189b486ff3b6.jpg",
        "acts": [{
           "page_id":1
          },
          {
            "page_id":6
          }
        ]
      }
    },
	"c": {
		"type": "carousel",
		"config": {
			"content_type": 1,
			"content_style": 1,
			"attribute": 1,     
			"background_color": "#FFFFFF",
			"indicator_color": "#AAAAAA",
			"title_image": "aaa",     
			"font_color": "aaa",   
			"scroll_type": 1,     
			"img_list": [
				{
					"img_url": "aaa", 
					"redirect_url": "aaa"
				}
			],
			"word_list": [
				{
					"content": "bbb"
				}
			]
		}  
	  },
	  "i": {
		"type": "icon",
		"config": {
			"background_color": "aaa",
			"font_color": "aaa",       
			"img_list": [
				{
					"img_url": "aaa",      
					"redirect_url": "aaa",
					"content": "bbb"       
				}
			]
		}
	  }
  }
}`
		)
		json.Unmarshal([]byte(str), portion)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SaveModule(c, nativeID, portion, "")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeSearchModule(t *testing.T) {
	convey.Convey("SearchModule", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			param = &natmdl.SearchModule{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.SearchModule(c, param)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestFindPage(t *testing.T) {
	convey.Convey("SaveModule", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.FindPage(c, "小豆芽", 0, 1, []int64{0})
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativebatchAddMixtureExt(t *testing.T) {
	convey.Convey("batchAddMixtureExt", t, func(convCtx convey.C) {
		var (
			tx = d.DB
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.batchAddMixtureExt(tx, 1, nil, nil)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
