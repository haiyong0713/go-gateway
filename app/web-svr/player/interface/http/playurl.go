package http

import (
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/player/interface/model"
)

const _allowFnver = 0

// nolint:gomnd
var fnvalCheck int32 = ^(0 | 1 | 2 | 16 | 64 | 128 | 256 | 512 | 1024 | 2048)

func playurl(c *bm.Context) {
	arg := new(model.PlayurlArg)
	if err := c.Bind(arg); err != nil {
		return
	}
	var err error
	if arg.Aid, err = bvArgCheck(arg.Aid, arg.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if arg.Fnval&fnvalCheck != 0 || arg.Fnver != _allowFnver {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var data *model.PlayurlRes
	if arg.HTML5 > 0 || arg.Platform == _platformH5 || arg.Platform == _platformH5New {
		data, err = playSvr.PlayurlH5(c, mid, arg)
	} else {
		data, err = playSvr.Playurl(c, mid, arg)
	}
	if err != nil {
		if playSvr.SLBRetry(err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func playurlHls(c *bm.Context) {
	params := &model.ParamHls{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	//从header获取
	buvid := c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	params.Buvid = buvid
	c.JSON(playSvr.PlayurlHls(c, mid, params))
}

func hlsMaster(c *bm.Context) {
	params := &model.ParamHls{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	//从header获取
	buvid := c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	params.Buvid = buvid
	//发生错误时，返回空内容
	rly, err := playSvr.HlsMaster(c, mid, params)
	if err != nil || rly == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	hlsResponse(c, rly.M3u8Data)
}

func hlsStream(c *bm.Context) {
	params := &model.ParamHls{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	//从header获取
	buvid := c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	params.Buvid = buvid
	//发生错误时，返回空内容
	rly, err := playSvr.M3U8Scheduler(c, mid, params)
	if err != nil || rly == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	hlsResponse(c, rly.M3u8Data)
}

func hlsResponse(c *bm.Context, body []byte) {
	c.Writer.Header().Set("Content-Type", "text; charset=UTF-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	_, _ = c.Writer.Write(body)
}
