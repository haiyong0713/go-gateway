package http

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/interface/model"
)

func tags(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	// get params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	ridStr := params.Get("rid")
	ver := params.Get("ver")
	// check params
	rid, err := strconv.ParseInt(ridStr, 10, 64)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	data, version, err := regionSvc.HotTags(c, mid, rid, ver, plat, time.Now())
	c.JSONMap(map[string]interface{}{"data": data, "ver": version}, err)
}

func subTags(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func addTag(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func cancelTag(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}
