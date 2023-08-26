package http

import (
	"context"
	"strconv"
	"strings"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

// ActIndex .
// nolint:gomnd
func ActIndex(c *bm.Context) {
	arg := &actmdl.ParamActIndex{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	arg.Buvid = header.Get(_headerBuvid)
	arg.TfIsp = header.Get(_headerTfIsp)
	arg.UserAgent = header.Get(_headerUserAgent)
	if arg.Ps > 50 { //一次最大请求50个组件
		arg.Ps = 50
	}
	fixMiaokai(c, "index", arg.MobiApp, arg.Build, arg.VideoMeta)
	c.JSON(actSvc.ActIndex(c, arg, mid))
}

func inlineTab(c *bm.Context) {
	arg := &actmdl.ParamInlineTab{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	arg.Buvid = header.Get(_headerBuvid)
	arg.TfIsp = header.Get(_headerTfIsp)
	arg.UserAgent = header.Get(_headerUserAgent)
	arg.Ps = 50
	fixMiaokai(c, "inline", arg.MobiApp, arg.Build, arg.VideoMeta)
	c.JSON(actSvc.InlineTab(c, arg, mid))
}

func menuTab(c *bm.Context) {
	arg := &actmdl.ParamMenuTab{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	arg.Buvid = header.Get(_headerBuvid)
	arg.TfIsp = header.Get(_headerTfIsp)
	arg.UserAgent = header.Get(_headerUserAgent)
	arg.Ps = 50 //一页最多取50个组件
	c.JSON(actSvc.MenuTab(c, arg, mid))
}

func ActLiked(c *bm.Context) {
	arg := &actmdl.ParamActLike{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(actSvc.ActLiked(c, arg, mid))
}

func actFollow(c *bm.Context) {
	arg := &actmdl.ParamActFollow{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	header := c.Request.Header
	arg.Buvid = header.Get(_headerBuvid)
	arg.UserAgent = header.Get(_headerUserAgent)
	c.JSON(actSvc.ActFollow(c, arg, mid))
}

func ActDetail(c *bm.Context) {
	arg := &actmdl.ParamActDetail{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	c.JSON(actSvc.ActDetail(c, arg))
}

func baseDetail(c *bm.Context) {
	arg := &actmdl.ParamBaseDetail{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	c.JSON(actSvc.BaseDetail(c, arg))
}

func LikeList(c *bm.Context) {
	arg := &actmdl.ParamLike{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	arg.Buvid = header.Get(_headerBuvid)
	arg.TfIsp = header.Get(_headerTfIsp)
	fixMiaokai(c, "like_list", arg.MobiApp, arg.Build, arg.VideoMeta)
	c.JSON(actSvc.LikeList(c, arg, mid))
}

func supernatant(c *bm.Context) {
	arg := &actmdl.ParamSupernatant{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(actSvc.Supernatant(c, arg, mid))
}

func actTab(c *bm.Context) {
	arg := &actmdl.ParamActTab{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSvc.ActTab(c, arg))
}

func actReceive(c *bm.Context) {
	arg := &actmdl.ParamReceive{}
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	data, msg, err := actSvc.ActReceive(c, arg, mid)
	c.JSON(struct {
		State int    `json:"state,omitempty"`
		Msg   string `json:"msg,omitempty"`
	}{State: data, Msg: msg}, err)
}

func fixMiaokai(c context.Context, from string, mobiApp string, build int64, rawVideoMeta string) {
	allow := false
	if feature.GetBuildLimit(c, cfg.Feature.FeatureBuildLimit.Miaokai, &feature.OriginResutl{
		BuildLimit: mobiApp == "iphone" && build <= 61500200,
	}) {
		allow = true
	}
	if feature.GetBuildLimit(c, cfg.Feature.FeatureBuildLimit.MiaokaiAndroid, &feature.OriginResutl{
		BuildLimit: mobiApp == "android" && build <= 6150400,
	}) && (from == "inline" || from == "like_list") {
		allow = true
	}
	if !allow {
		return
	}
	batchArg, ok := arcmid.FromContext(c)
	if !ok {
		return
	}
	videoMeta := parseVideoMeta(rawVideoMeta)
	batchArg.Qn, _ = strconv.ParseInt(videoMeta["qn"], 10, 64)
	batchArg.Fnver, _ = strconv.ParseInt(videoMeta["fnver"], 10, 64)
	batchArg.Fnval, _ = strconv.ParseInt(videoMeta["fnval"], 10, 64)
}

// nolint:gomnd
func parseVideoMeta(videoMeta string) map[string]string {
	kvs := strings.Split(videoMeta, ",")
	m := make(map[string]string)
	for _, kv := range kvs {
		data := strings.Split(kv, ":")
		if len(data) < 2 {
			continue
		}
		m[data[0]] = data[1]
	}
	return m
}
