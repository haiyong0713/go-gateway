package http

import (
	"net/http"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-player/interface/model"
)

const (
	_ugcType  = 1
	_pgcType  = 2
	_pugvType = 3
	_liveType = 4
)

func playurl(c *bm.Context) {
	params := &model.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	if params.Buvid == "" {
		params.Buvid = header.Get("Buvid")
	}
	plat := model.Plat(params.MobiApp, params.Device)
	params.NetType, params.TfType = model.TrafficFree(header.Get("X-Tf-Isp"))
	if params.Npcybs < 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if params.Otype != "json" && params.Otype != "xml" {
		params.Otype = "json"
	}
	if params.Dl == 1 {
		params.Download = model.DlDash
	} else if params.Npcybs == 1 {
		params.Download = model.DlFlv
	}
	params.FourkBool = params.Fourk == 1
	c.JSON(svr.PlayURLV2(c, mid, params, plat))
}

func dlNum(c *bm.Context) {
	params := &model.DlNumParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, svr.DlNum(c, mid, params))
}

func playurlOtt(c *bm.Context) {
	params := &model.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	if params.Buvid == "" {
		params.Buvid = header.Get("Buvid")
	}
	plat := model.Plat(params.MobiApp, params.Device)
	if params.Dl == 1 {
		params.Download = model.DlDash
	}
	params.FourkBool = params.Fourk == 1
	c.JSON(svr.PlayURLV2(c, mid, params, plat))
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
	header := c.Request.Header
	//从header获取
	params.Buvid = header.Get("Buvid")
	params.XTfIsp = header.Get("X-Tf-Isp")
	params.NetType, params.TfType = model.TrafficFree(params.XTfIsp)
	c.JSON(svr.PlayurlHls(c, mid, params))
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
	header := c.Request.Header
	//从header获取
	params.Buvid = header.Get("Buvid")
	params.XTfIsp = header.Get("X-Tf-Isp")
	params.NetType, params.TfType = model.TrafficFree(params.XTfIsp)
	//发生错误时，返回空内容
	rly, err := svr.HlsMaster(c, mid, params)
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

func hlsStream(c *bm.Context) {
	params := &model.ParamHls{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	//从header获取
	params.Buvid = header.Get("Buvid")
	params.XTfIsp = header.Get("X-Tf-Isp")
	params.NetType, params.TfType = model.TrafficFree(params.XTfIsp)
	//发生错误时，返回空内容
	rly, err := svr.M3U8Scheduler(c, mid, params)
	if err != nil || rly == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	hlsResponse(c, rly.M3u8Data)
}

func bubble(c *bm.Context) {
	params := new(model.BubbleParam)
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if (params.Aid <= 0 || params.Cid <= 0) && (params.SeasonId <= 0 || params.EpId <= 0) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(svr.Bubble(c, params, mid))
}

func bubbleSubmit(c *bm.Context) {
	req := new(struct {
		Code string `form:"code" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(svr.BubbleSubmit(c, mid, req.Code))
}

func projPageAct(c *bm.Context) {
	params := new(model.ProjPageParam)
	if err := c.Bind(params); err != nil {
		return
	}
	authParam := &model.AuthArcParam{
		PlayurlType: params.PlayurlType,
		Aid:         params.Aid,
		Cid:         params.Cid,
		SeasonId:    params.SeasonId,
		EpId:        params.EpId,
		RoomId:      0,
	}
	if err := authArcId(authParam); err != nil {
		c.JSON(nil, err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.Mid = midInter.(int64)
	}
	c.JSON(svr.ProjPageAct(c, params))
}

func projActAll(c *bm.Context) {
	params := new(model.ProjActAllParam)
	if err := c.Bind(params); err != nil {
		return
	}
	authParam := &model.AuthArcParam{
		PlayurlType: params.PlayurlType,
		Aid:         params.Aid,
		Cid:         params.Cid,
		SeasonId:    params.SeasonId,
		EpId:        params.EpId,
		RoomId:      params.RoomId,
	}
	if err := authArcId(authParam); err != nil {
		c.JSON(nil, err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.Mid = midInter.(int64)
	}
	c.JSON(svr.ProjActAll(c, params))
}

func authArcId(param *model.AuthArcParam) error {
	if param == nil {
		return ecode.RequestErr
	}
	switch param.PlayurlType {
	case _ugcType:
		if param.Aid <= 0 || param.Cid <= 0 {
			return ecode.RequestErr
		}
	case _pgcType, _pugvType:
		if param.Aid <= 0 || param.Cid <= 0 || param.SeasonId <= 0 || param.EpId <= 0 {
			return ecode.RequestErr
		}
	case _liveType:
		if param.RoomId <= 0 {
			return ecode.RequestErr
		}
	default:
		return ecode.RequestErr
	}
	return nil
}
