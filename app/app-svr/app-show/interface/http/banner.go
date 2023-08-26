package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-show/interface/model"
)

// banners get banners.
func banners(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	buildStr := params.Get("build")
	channel := params.Get("channel")
	module := params.Get("module")
	position := params.Get("position")
	// check param
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("build(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	data := bannerSvc.Display(c, plat, build, channel, module, position, mobiApp)
	returnJSON(c, data, nil)
}
