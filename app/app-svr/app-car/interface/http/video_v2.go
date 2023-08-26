package http

import (
	"encoding/json"

	bm "go-common/library/net/http/blademaster"
	commonModel "go-gateway/app/app-svr/app-car/interface/model/common"
)

func regionMeta(c *bm.Context) {
	c.JSON(commonSvc.RegionMeta(c))
}

func videoTabs(c *bm.Context) {
	req := &commonModel.VideoTabsReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	if midObj, ok := c.Get("mid"); ok {
		req.Mid = midObj.(int64)
	}
	if c.Request != nil && c.Request.Header != nil {
		req.Buvid = c.Request.Header.Get(_headerBuvid)
	}
	if req.Buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			req.Buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.VideoTabs(c, req))
}

func videoTabCards(c *bm.Context) {
	req := &commonModel.VideoTabCardReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	if midObj, ok := c.Get("mid"); ok {
		req.Mid = midObj.(int64)
	}
	if c.Request != nil && c.Request.Header != nil {
		req.Buvid = c.Request.Header.Get(_headerBuvid)
	}
	if req.Buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			req.Buvid = cookie.Value
		}
	}
	if req.PageNextStr != "" {
		pageNext := &commonModel.PageNext{}
		if err := json.Unmarshal([]byte(req.PageNextStr), &pageNext); err == nil {
			req.PageNext = pageNext
		}
	}
	c.JSON(commonSvc.VideoTabCards(c, req))
}

func cardPlaylist(c *bm.Context) {
	req := &commonModel.CardPlaylistReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	if midObj, ok := c.Get("mid"); ok {
		req.Mid = midObj.(int64)
	}
	if c.Request != nil && c.Request.Header != nil {
		req.Buvid = c.Request.Header.Get(_headerBuvid)
	}
	if req.Buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			req.Buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.CardPlaylist(c, req))
}
