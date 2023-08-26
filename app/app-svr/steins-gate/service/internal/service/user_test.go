package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestServiceLatestGraphList(t *testing.T) {
	convey.Convey("LatestGraphList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515257)
			aid = int64(10113611)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := s.LatestGraphList(c, mid, aid)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestServiceGraphShow(t *testing.T) {
	convey.Convey("GraphShow", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			mid     = int64(27515257)
			aid     = int64(10113611)
			graphID = int64(30)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := s.GraphShow(c, mid, aid, graphID)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestServicePlayurl(t *testing.T) {
	convey.Convey("Playurl", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515257)
			arg = &model.PlayurlParam{Aid: 10113611, Cid: 10162784}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := s.Playurl(c, mid, arg)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestServiceMsgCheck(t *testing.T) {
	convey.Convey("MsgCheck", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			msg = "测试"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := s.MsgCheck(c, msg)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestServiceGraphCheck(t *testing.T) {
	convey.Convey("GraphCheck", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(10113629)
			cid = int64(10162855)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			check, err := s.GraphCheck(c, aid, cid)
			convCtx.Convey("Then err should be nil.check should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(check, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestServicecheckGraphParam(t *testing.T) {
	convey.Convey("checkGraphParam", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			mid       = int64(27515257)
			isPreview = 0
			param     = &model.SaveGraphParam{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, _, _, err := s.checkGraphParam(c, mid, isPreview, param)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestServiceexclusionCond(t *testing.T) {
	convey.Convey("checkGraphParam", t, func(convCtx convey.C) {
		var (
			conds = map[string]int{
				model.EdgeConditionTypeGt: 1,
			}
		)
		convCtx.Convey("When conds 0 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeGt: 1,
			model.EdgeConditionTypeGe: 2,
		}
		convCtx.Convey("When conds 1 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeLt: 1,
			model.EdgeConditionTypeLe: 1,
		}
		convCtx.Convey("When conds 2 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeGt: 1,
			model.EdgeConditionTypeLe: 1,
		}
		convCtx.Convey("When conds 3 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeGt: 1,
			model.EdgeConditionTypeLt: 1,
		}
		convCtx.Convey("When conds 4 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeGe: 1,
			model.EdgeConditionTypeLt: 1,
		}
		convCtx.Convey("When conds 5 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeGe: 2,
			model.EdgeConditionTypeLe: 1,
		}
		convCtx.Convey("When conds 6 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeTrue)
		})
		conds = map[string]int{
			model.EdgeConditionTypeGe: 1,
			model.EdgeConditionTypeLe: 1,
		}
		convCtx.Convey("When conds 7 goes positive", func(convCtx convey.C) {
			check := exclusionCond(conds)
			convCtx.So(check, convey.ShouldBeFalse)
		})
	})
}

func TestService_CheckDimension(t *testing.T) {
	convey.Convey("TestService_CheckDimension", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			view = new(model.VideoUpView)
		)
		view.Videos = append(view.Videos, &model.Video{
			Cid:   10162482,
			Title: "哈哈",
		},
			&model.Video{
				Cid:   10162482,
				Title: "heihei",
			},
			&model.Video{
				Cid:   10162496,
				Title: "哈哈",
			})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, _, err := s.checkDimension(c, view, nil)
			str, _ := json.Marshal(res)
			fmt.Println(string(str), err)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestService_GraphDiff(t *testing.T) {
	convey.Convey("tt", t, func(c convey.C) {
		var ctx = context.Background()
		str := `{"graph":{"script":"{\"nodes\":{\"n-ZiohK1it#z\":{\"id\":\"n-ZiohK1it#z\",\"type\":\"videoNode\",\"data\":{\"type\":1,\"aid\":\"\",\"cid\":10164932,\"name\":\"2222\",\"index\":1,\"showTime\":5,\"dimension\":{\"width\":640,\"height\":360}},\"isRoot\":true,\"input\":[],\"output\":[\"l-ATpWqedL1T\",\"l-of@HzMOJu4\"],\"refInput\":[],\"refOutput\":[]},\"n-BcpBmRb1JI\":{\"id\":\"n-BcpBmRb1JI\",\"type\":\"videoNode\",\"data\":{\"type\":0,\"aid\":\"\",\"cid\":10164934,\"name\":\"212121我问清楚\",\"index\":3,\"showTime\":0,\"dimension\":{\"width\":640,\"height\":360}},\"isRoot\":false,\"input\":[\"l-ATpWqedL1T\"],\"output\":[],\"refInput\":[],\"refOutput\":[]},\"n-BFVbkuDWnq\":{\"id\":\"n-BFVbkuDWnq\",\"type\":\"videoNode\",\"data\":{\"type\":0,\"aid\":\"\",\"cid\":10164933,\"name\":\"随便看看\",\"index\":2,\"showTime\":0},\"isRoot\":false,\"input\":[\"l-of@HzMOJu4\"],\"output\":[],\"refInput\":[],\"refOutput\":[]}},\"links\":{\"l-ATpWqedL1T\":{\"id\":\"l-ATpWqedL1T\",\"type\":\"flowLink\",\"data\":{\"id\":\"l-K6VOmLeaA4\",\"text\":\"212121\",\"default\":true,\"conditions\":[],\"actions\":[],\"point\":{\"x\":0.8119122257053291,\"y\":0.7518221685246762,\"align\":3}},\"from\":\"n-ZiohK1it#z\",\"to\":\"n-BcpBmRb1JI\"},\"l-of@HzMOJu4\":{\"id\":\"l-of@HzMOJu4\",\"type\":\"flowLink\",\"data\":{\"id\":\"l-nsmc6G9lKw\",\"text\":\"随便试试\",\"default\":false,\"conditions\":[],\"actions\":[]},\"from\":\"n-ZiohK1it#z\",\"to\":\"n-BFVbkuDWnq\"}},\"hasGoto\":false,\"editorVersion\":\"0.1.0\",\"createdTime\":1564996877900,\"enableVariables\":true,\"variables\":[{\"id\":\"v-AND5rJny4R\",\"type\":1,\"name\":\"数值1\",\"initValue\":0},{\"id\":\"v-oyPR3VtzLQ\",\"type\":1,\"name\":\"数值2\",\"initValue\":0},{\"id\":\"v-y8zVORTTwS\",\"type\":1,\"name\":\"数值3\",\"initValue\":0},{\"id\":\"v-RTQ03hK9dF\",\"type\":1,\"name\":\"数值4\",\"initValue\":0},{\"id\":\"v-oWbjteRU#1\",\"type\":2,\"name\":\"随机值\",\"initValue\":1,\"initValue2\":100}]}","aid":10114207,"nodes":[{"id":"n-ZiohK1it#z","cid":10164932,"name":"2222","is_start":1,"show_time":5,"otype":1,"edges":[{"title":"212121","to_node_id":"n-BcpBmRb1JI","is_default":1,"condition":[],"attribute":[]},{"title":"随便试试","to_node_id":"n-BFVbkuDWnq","is_default":0,"condition":[],"attribute":[]}]},{"id":"n-BcpBmRb1JI","cid":10164934,"name":"212121我问清楚","is_start":0,"show_time":0,"otype":1,"edges":[]},{"id":"n-BFVbkuDWnq","cid":10164933,"name":"随便看看","is_start":0,"show_time":0,"otype":1,"edges":[]}],"regional_vars":[{"name":"数值1","init_min":0,"init_max":0,"type":1,"id":"v-AND5rJny4R"},{"name":"数值2","init_min":0,"init_max":0,"type":1,"id":"v-oyPR3VtzLQ"},{"name":"数值3","init_min":0,"init_max":0,"type":1,"id":"v-y8zVORTTwS"},{"name":"数值4","init_min":0,"init_max":0,"type":1,"id":"v-RTQ03hK9dF"},{"name":"随机值","init_min":1,"init_max":100,"type":2,"id":"v-oWbjteRU#1"}]}}`
		param := new(model.SaveGraphParam)
		if err := json.Unmarshal([]byte(str), &param); err != nil {
			fmt.Println(err)
			return
		}
		s.graphDiff(ctx, param, 0)
	})
}

func TestService_SaveGraph(t *testing.T) {
	convey.Convey("tt", t, func(c convey.C) {
		var (
			ctx       = context.Background()
			mid       = int64(12404946)
			isPreview = 1
		)
		str := `{"graph":{"script":"","aid":10114549,"nodes":[{"id":"n-OKeosfif#X","cid":10165988,"name":"在4在在在在在在在在","is_start":1,"show_time":-1,"otype":1,"edges":[{"id":"l-S5rHZbJntj","title":"test","to_node_id":"n-9FWdXjFAJ","is_default":1,"condition":[],"attribute":[]}]},{"id":"n-9FWdXjFAJ","cid":10165989,"name":"qeqeqeqea1","is_start":0,"show_time":-1,"otype":1,"edges":[{"id":"l-exBd1r68Zn","title":"11111111","to_node_id":"n-lK0a44xrI","is_default":1,"condition":[],"attribute":[]}]},{"id":"n-lK0a44xrI","cid":10165990,"name":"横屏41111","is_start":0,"show_time":-1,"otype":1,"edges":[]}],"regional_vars":[{"name":"数值数值数数值数值数","init_min":-61,"init_max":0,"type":1,"id":"v-iWRQZAa27","is_show":1},{"name":"数值2","init_min":1,"init_max":0,"type":1,"id":"v-vrU4lgsShd","is_show":1},{"name":"数值3","init_min":2,"init_max":0,"type":1,"id":"v-Zu7nUiWQrj","is_show":0},{"name":"数值4","init_min":3,"init_max":0,"type":1,"id":"v-Oybt6YGnhW","is_show":0},{"name":"随机值","init_min":1,"init_max":100,"type":2,"id":"v-jiKZNJc#WW","is_show":0}]}}`
		param := new(model.SaveGraphParam)
		if err := json.Unmarshal([]byte(str), &param); err != nil {
			fmt.Println(err)
			return
		}
		gid, _, err := s.SaveGraph(ctx, mid, isPreview, param)
		fmt.Println(gid, err)
	})
}

func TestService_GetDiffMsg(t *testing.T) {
	convey.Convey("GetDiffMsg", t, func(c convey.C) {
		newNodeNames := strings.Split(`结束_分流结果、你想的数字是在1-15之间、你想的数字是1到4之间、你想的数字是1、你想的数字是2、你想的数字是3、你想的数字是4、你想的数字是5到8之间、你想的是数字5、你想的是数字6、你想的是数字7、你想的是数字8、你想的数字是9到12之间、你想的是数字9、你想的是数字10、你想的是数字11、你想的是数字12、你想的数字是13到15之间、你想的是数字13、你想的是数字14、你想的是数字15、你想的数字是在16-30之间、你想的数字是16到19之间、你想的数字是16、你想的数字是17、你想的数字是18、你想的数字是19、你想的数字是20到23之间、你想的数字是20、你想的数字是21、你想的数字是22、你想的数字是23、你想的数字是24到27之间、你想的数字是24、你想的数字是25、你想的数字是26、你想的数字是27、你想的数字是28到30之间、你想的数字是28、你想的数字是29、你想的数字是30、你想的数字是在31-45之间、你想的数字是31到34之间、你想的数字是31、你想的数字是32、你想的数字是33、你想的数字是34、你想的数字是35到38之间、你想的数字是35、你想的数字是36、你想的数字是37、你想的数字是38、你想的数字是39到42之间、你想的数字是39、你想的数字是40、你想的数字是41、你想的数字是42、你想的数字是43到45之间、你想的数字是43、你想的数字是44、你想的数字是45、你想的数字是在46-60之间、你想的数字是46到49之间、你想的数字是46、你想的数字是47、你想的数字是48、你想的数字是49、你想的数字是50到53之间、你想的数字是50、你想的数字是51、你想的数字是52、你想的数字是53、你想的数字是54到57之间、你想的数字是54、你想的数字是55、你想的数字是56、你想的数字是57、你想的数字是58到60之间、你想的数字是58、你想的数字是59、你想的数字是60】 选项名称:【想好了，开始6次问答、开始猜测你心里的数值、开始猜测你心里的数值、开始猜测你心里的数值、开始猜测你心里的数值、还需5秒感应，点击加速！（猜对三连）、还需5秒感应，点击加速！（猜对三连）、还需5秒感应，点击加速！（猜对三连）、还需5秒感应，点击加速！（猜对三连）`, "、")
		convey.Println(newNodeNames)
		convey.Println(len(newNodeNames))
		varsNames := "你所想的数字是"
		str := s.getDiffMsg(newNodeNames, []string{"123"}, varsNames)
		convey.Println(utf8.RuneCountInString(str))
	})
}
