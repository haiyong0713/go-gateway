package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/show"
)

func tabs(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		res    = map[string]interface{}{}
	)
	channel := params.Get("channel")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	buvid := header.Get(_headerBuvid)
	device := params.Get("device")
	language := params.Get("lang")
	platform := params.Get("platform")
	plat := model.Plat2(mobiApp, device)
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	slocale := params.Get("s_locale")
	clocale := params.Get("c_locale")
	data, config, ab, err := showSvc.Tabs(c, plat, build, teenagersMode, lessonsMode, buvid, mobiApp, platform, language, channel, mid, slocale, clocale)
	if ab != nil {
		res["abtest"] = ab
	}
	res["data"] = data
	res["config"] = config
	c.JSONMap(res, err)
}

func tabBubble(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		err    error
	)
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	buvid := header.Get(_headerBuvid)
	device := params.Get("device")
	language := params.Get("lang")
	plat := model.Plat2(mobiApp, device)
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	c.JSON(showSvc.TabBubble(c, plat, build, teenagersMode, lessonsMode, buvid, language, mid))
}

func skin(c *bm.Context) {
	var (
		params = c.Request.Form
	)
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	device := params.Get("device")
	isFreeTheme, _ := strconv.ParseBool(params.Get("is_free_theme"))
	plat := model.Plat(mobiApp, device)
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(showSvc.Skin(c, plat, build, mid, mobiApp, isFreeTheme))
}

func clickTab(c *bm.Context) {
	var (
		header = c.Request.Header
	)
	buvid := header.Get(_headerBuvid)
	var arg struct {
		ID   int64  `form:"id" validate:"required"`
		Ver  string `form:"ver" validate:"required"`
		Type string `form:"type" validate:"required"`
	}
	if err := c.Bind(&arg); err != nil {
		return
	}
	c.JSON(nil, showSvc.ClickTab(c, arg.ID, buvid, arg.Ver, arg.Type))
}

func topActivity(c *bm.Context) {
	params := &show.TopActivityReq{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid := c.Request.Header.Get(_headerBuvid)
	ua := c.Request.Header.Get(_userAgent)
	c.JSON(showSvc.TopActivity(c, params, mid, buvid, ua))
}

func tabsV2(c *bm.Context) {
	var (
		params = &show.TabsV2Params{}
		header = c.Request.Header
		res    = map[string]interface{}{}
	)
	if err := c.Bind(params); err != nil {
		return
	}
	buvid := header.Get(_headerBuvid)
	plat := model.Plat2(params.MobiApp, params.Device)
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data, config, err := showSvc.TabsV2(c, plat, buvid, mid, params)
	res["data"] = data
	res["config"] = config
	c.JSONMap(res, err)
}

func vivoPopularBadge(ctx *bm.Context) {
	ctx.JSON(showSvc.VIVOPopularBadge(ctx))
}
