package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"strconv"

	"go-gateway/app/app-svr/app-resource/interface/model"
	resMdl "go-gateway/app/app-svr/app-resource/interface/model/resource"
)

func download(c *bm.Context) {
	params := c.Request.Form

	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	ver := params.Get("ver")
	resType := params.Get("type")

	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid := c.Request.Header.Get("Buvid")

	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("build(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}

	res := &resMdl.ResourceDownloadRequest{
		MobiApp:  mobiApp,
		Device:   device,
		Build:    build,
		Mid:      mid,
		Buvid:    buvid,
		Platform: model.Plat(mobiApp, device),
		Ver:      ver,
		Type:     resType,
	}
	c.JSON(staticSvc.Download(c, res))
}
