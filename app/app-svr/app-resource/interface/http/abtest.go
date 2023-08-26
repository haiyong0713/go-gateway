package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-resource/interface/model"
	abTestMdl "go-gateway/app/app-svr/app-resource/interface/model/abtest"
)

func abTest(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(abSvc.Experiment(c, plat, build), nil)
}

func abTestV2(c *bm.Context) {
	params := c.Request.Form
	header := c.Request.Header
	buvid := params.Get("buvid")
	if buvid == "" {
		buvid = header.Get(_headerBuvid)
	}
	if buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(abSvc.TemporaryABTests(c, buvid, mid), nil)
}

func abserver(c *bm.Context) {
	params := c.Request.Form
	buvid := params.Get("buvid")
	device := params.Get("device")
	mobiAPP := params.Get("mobi_app")
	buildStr := params.Get("build")
	filteredStr := params.Get("filtered")
	if buvid == "" || device == "" || mobiAPP == "" || buildStr == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	model := params.Get("model")
	brand := params.Get("brand")
	osver := params.Get("osver")
	c.JSON(abSvc.AbServer(c, buvid, device, mobiAPP, filteredStr, model, brand, osver, build, mid))
}

func abTestList(c *bm.Context) {
	var param = new(abTestMdl.AbTestListParam)
	if err := c.Bind(param); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	var (
		buvid = c.Request.Header.Get("Buvid")
	)
	c.JSON(abSvc.AbTestList(c, param, mid, buvid))
}

func tinyAbtest(c *bm.Context) {
	c.JSON(abSvc.TinyAbtest(c), nil)
}
