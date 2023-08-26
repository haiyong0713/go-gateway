package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model/search"
)

func wxHot(c *bm.Context) {
	v := new(struct {
		Pn          int    `form:"pn" default:"1" validate:"min=1"`
		Ps          int    `form:"ps" default:"100" validate:"min=1,max=100"`
		Platform    string `form:"platform"`
		TeenageMode int    `form:"teenage_mode" default:"0"` // 是否是青少年模式
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var (
		mid   int64
		buvid string
	)
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	list, count, err := webSvc.WxHot(c, v.Pn, v.Ps, v.Platform, v.TeenageMode, mid, buvid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"pn":    v.Pn,
		"ps":    v.Ps,
		"count": count,
	}
	data["data"] = list
	data["page"] = page
	c.JSONMap(data, nil)
}

func wxSearchAll(c *bm.Context) {
	var (
		mid   int64
		buvid string
		err   error
	)
	v := new(search.SearchAllArg)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Pn <= 0 {
		v.Pn = 1
	}
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.SearchAll(c, mid, v, buvid, c.Request.Header.Get("User-Agent"), search.WxSearchType))
}
