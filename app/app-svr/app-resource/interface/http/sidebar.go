package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/model"
)

func sidebar(c *bm.Context) {
	var (
		params = c.Request.Form
		module int
		mid    int64
	)
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	device := params.Get("device")
	language := params.Get("lang")
	plat := model.Plat(mobiApp, device)
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	switch plat {
	case model.PlatAndroid, model.PlatAndroidG, model.PlatAndroidB, model.PlatAndroidI:
		module = 1
	case model.PlatIPhone, model.PlatIPhoneI, model.PlatIPhoneB:
		module = 3
	}
	c.JSON(sideSvr.SideBar(c, plat, build, module, teenagersMode, lessonsMode, mid, language), nil)
}

func topbar(c *bm.Context) {
	var (
		params = c.Request.Form
		module int
		mid    int64
	)
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	device := params.Get("device")
	language := params.Get("lang")
	plat := model.Plat(mobiApp, device)
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	switch plat {
	case model.PlatAndroid, model.PlatAndroidG, model.PlatAndroidB, model.PlatAndroidI:
		module = 2
	case model.PlatIPhone, model.PlatIPhoneI, model.PlatIPhoneB:
		module = 4
	case model.PlatIPad, model.PlatIPadI:
		module = 5
	}
	c.JSON(sideSvr.SideBar(c, plat, build, module, 0, 0, mid, language), nil)
}
