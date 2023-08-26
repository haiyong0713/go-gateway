package http

import (
	"strconv"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/model"
)

// getNotice get notice data.
func getNotice(c *bm.Context) {
	params := c.Request.Form
	ver := params.Get("ver")
	buildStr := params.Get("build")
	mobiApp := params.Get("mobi_app")
	typeStr := params.Get("type")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("stronv.ParseInt(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	typeInt, _ := strconv.Atoi(typeStr)
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	// get
	data, version, err := ntcSvc.Notice(c, plat, build, typeInt, ver)
	res := map[string]interface{}{
		"data": data,
		"ver":  version,
	}
	c.JSONMap(res, err)
}

func getPackagePushMsg(c *bm.Context) {
	dev, ok := device.FromContext(c)
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(ntcSvc.GetPackagePushMsg(c, dev.Buvid, dev.Model))
}
