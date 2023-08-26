package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/relate"
	"go-gateway/app/app-svr/app-car/interface/model/view"
)

const (
	_headerCookie = "Cookie"
)

func viewIndex(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &view.ViewParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := viewSvc.View(c, plat, mid, buvid, param)
	c.JSON(data, err)
}

func relateAll(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &relate.RelateParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	c.JSON(showSvc.Relate(c, plat, mid, buvid, param))
}

func relateWebAll(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	param := &relate.RelateParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.RelateWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
	}{
		Item: data,
	}, err)
}

func ViewWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &view.ViewWebParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	referer := c.Request.Referer()
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	cookie := c.Request.Header.Get(_headerCookie)
	data, err := viewSvc.ViewWeb(c, mid, cookie, buvid, referer, param)
	c.JSON(data, err)
}

func like(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &view.LikeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	err := viewSvc.Like(c, mid, buvid, c.Request.URL.Path, c.Request.UserAgent(), param)
	c.JSON(nil, err)
}

func communityPGC(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &view.CommunityParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	c.JSON(viewSvc.CommunityPGC(c, mid, buvid, param))
}

func likeWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &view.LikeWebParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	err := viewSvc.LikeWeb(c, mid, buvid, c.Request.URL.Path, c.Request.UserAgent(), param)
	c.JSON(nil, err)
}

func communityWebPGC(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &view.CommunityWebParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(viewSvc.CommunityWebPGC(c, mid, buvid, param))
}

func viewV2Detail(c *bm.Context) {
	var (
		req = new(commonmdl.ViewDetailReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	cookie := header.Get(_headerCookie)
	referer := c.Request.Referer()
	if req.DeviceInfo.MobiApp == "" {
		req.DeviceInfo.MobiApp = model.AndroidBilithings
	}
	if req.DeviceInfo.Platform == "" {
		req.DeviceInfo.Platform = "android"
	}
	c.JSON(commonSvc.ViewDetail(c, req, mid, buvid, cookie, referer))
}

func viewV2Rcmd(c *bm.Context) {
	var (
		req = new(commonmdl.ViewRcmdReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.ViewRcmd(c, req, mid, buvid))
}

func viewV2Serial(c *bm.Context) {
	req := new(commonmdl.ViewV2SerialReq)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		req.Mid = midInter.(int64)
	}
	header := c.Request.Header
	req.Buvid = header.Get(_headerBuvid)
	if req.Buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			req.Buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.ViewPlaylist(c, req))
}

// mediaParse 特斯拉资源解析.
func mediaParse(c *bm.Context) {
	param := &struct {
		Url string `form:"url"`
	}{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(commonSvc.MediaParse(c, param.Url))
}
