package http

import (
	"net/http"
	"time"

	bm "go-common/library/net/http/blademaster"
	webmdl "go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const (
	_appName     = `哔哩哔哩`
	_packageName = `tv.danmaku.bili`
)

func fullshort(c *bm.Context) {
	var (
		err   error
		items []*webmdl.Mi
	)
	v := new(struct {
		Pn     int64  `form:"pn"     validate:"min=1"`
		Ps     int64  `form:"ps"     validate:"min=1,max=50"`
		Source string `form:"bsource"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if items, err = srvWeb.FullShort(c, v.Pn, v.Ps, v.Source); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 4)
	data["app_name"] = _appName
	data["package_name"] = _packageName
	data["update_time"] = time.Now().Format("2006-01-02 15:04:05")
	data["shortvideos"] = items
	c.JSONMap(data, nil)
}

func outArc(c *bm.Context) {
	v := new(struct {
		Pn int64 `form:"pn" default:"1" validate:"min=1"`
		Ps int64 `form:"ps" default:"10" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, total := srvWeb.OutArc(c, v.Pn, v.Ps)
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"pn":    v.Pn,
		"ps":    v.Ps,
		"count": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func baiduPushArcContent(ctx *bm.Context) {
	ctx.Bytes(http.StatusOK, "text/json; charset=utf-8", srvWeb.BaiduPushArcContent(ctx))
}
