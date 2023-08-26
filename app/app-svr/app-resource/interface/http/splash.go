package http

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/splash"
)

const (
	_userAgent = "User-Agent"
)

// splashs splash handler
func splashs(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func birthSplash(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

// splashList ad splash handler
func splashList(c *bm.Context) {
	var (
		header = c.Request.Header
		params = c.Request.Form
	)
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	widthStr := params.Get("width")
	heightStr := params.Get("height")
	buildStr := params.Get("build")
	birth := params.Get("birth")
	adExtra := params.Get("ad_extra")
	device := params.Get("device")
	// check params
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		log.Error("width(%s) is invalid", widthStr)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		log.Error("height(%s) is invalid", heightStr)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	build, _ := strconv.Atoi(buildStr)
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	network := params.Get("network")
	buvid := header.Get(_headerBuvid)
	plat := model.Plat(mobiApp, device)
	userAgent := header.Get(_userAgent)
	loadedCreativeList := params.Get("loaded_creative_list")
	clientKeepIds := params.Get("client_keep_ids")
	result, err := splashSvc.AdList(c, plat, mobiApp, device, buvid, birth, adExtra, height, width, build, mid,
		userAgent, network, loadedCreativeList, clientKeepIds)
	c.JSON(result, err)
}

func splashRtShow(ctx *bm.Context) {
	req := &splash.SplashRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.MobiApp = model.MobiAPPBuleChange(req.MobiApp)
	midI, ok := ctx.Get("mid")
	if ok {
		req.Mid = midI.(int64)
	}
	req.Plat = model.Plat(req.MobiApp, req.Device)
	header := ctx.Request.Header
	req.UserAgent = header.Get(_userAgent)
	req.Buvid = header.Get(_headerBuvid)
	ctx.JSON(splashSvc.AdRtShow(ctx, req))
}

func splashState(c *bm.Context) {
	var (
		header = c.Request.Header
		params = c.Request.Form
	)
	buvid := header.Get(_headerBuvid)
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	state, _ := strconv.Atoi(params.Get("state"))
	result := splashSvc.State(c, buvid, mid, int8(state))
	c.JSON(result, nil)
}

func brandList(c *bm.Context) {
	param := &splash.SplashParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(splashSvc.BrandList(c, param))
}

func brandSet(c *bm.Context) {
	param := &splash.SplashParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(splashSvc.BrandSet(c, param))
}

func brandSave(c *bm.Context) {
	var (
		mid    int64
		header = c.Request.Header
		param  = &splash.SplashSaveParam{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	buvid := header.Get(_headerBuvid)
	splashSvc.BrandSave(c, param, "/x/v2/splash/brand/save", buvid, mid, time.Now())
	c.JSON(nil, nil)
}

func eventList(ctx *bm.Context) {
	req := &splash.EventSplashRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	midI, ok := ctx.Get("mid")
	if ok {
		req.Mid = midI.(int64)
	}
	ctx.JSON(splashSvc.EventSplashList(ctx, req))

}

func eventList2(ctx *bm.Context) {
	req := &splash.EventSplashRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	midI, ok := ctx.Get("mid")
	if ok {
		req.Mid = midI.(int64)
	}
	ctx.JSON(splashSvc.EventSplashList2(ctx, req))
}
