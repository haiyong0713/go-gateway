package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/model"
)

func serverList(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	c.JSON(broadcastSvc.ServerList(c, plat))
}
