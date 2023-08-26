package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	rsmdl "go-gateway/app/web-svr/web-show/interface/model/resource"
)

const (
	_headerBuvid = "Buvid"
	_buvid       = "buvid3"
)

func resources(c *bm.Context) {
	arg := new(rsmdl.ArgRess)
	if err := c.Bind(arg); err != nil {
		return
	}
	arg.Mid, arg.Sid, arg.Buvid = device(c)
	arg.UserAgent = c.Request.UserAgent()
	data, adsControl, count, err := resSvc.Resources(c, arg)
	if err != nil {
		log.Error("resSvc.Resource error(%v)", err)
		if !ecode.EqualError(ecode.NothingFound, err) {
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	bs, _ := json.Marshal(data)
	log.Info("h.resources arg:(%+v); data:(%+v)", arg, string(bs))
	c.JSONMap(map[string]interface{}{
		"count":       count,
		"ads_control": adsControl,
		"data":        data,
	}, nil)
}

func resource(c *bm.Context) {
	arg := new(rsmdl.ArgRes)
	if err := c.Bind(arg); err != nil {
		return
	}
	arg.Mid, arg.Sid, arg.Buvid = device(c)
	arg.UserAgent = c.Request.UserAgent()
	data, count, err := resSvc.Resource(c, arg)
	if err != nil {
		log.Error("resSvc.Resource error(%v)", err)
		if !ecode.EqualError(ecode.NothingFound, err) {
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSONMap(map[string]interface{}{
		"count": count,
		"data":  data,
	}, nil)
}

func relation(c *bm.Context) {
	arg := new(rsmdl.ArgAid)
	if err := c.Bind(arg); err != nil {
		return
	}
	arg.Mid, arg.Sid, arg.Buvid = device(c)
	arg.UserAgent = c.Request.UserAgent()
	c.JSON(resSvc.Relation(c, arg))
}

func advideo(c *bm.Context) {
	arg := new(rsmdl.ArgAid)
	if err := c.Bind(arg); err != nil {
		return
	}
	midTemp, ok := c.Get("mid")
	if !ok {
		log.Info("mid not exist")
		arg.Mid = 0
	} else {
		arg.Mid = midTemp.(int64)
	}
	c.JSON(resSvc.VideoAd(c, arg), nil)
}

func urlMonitor(c *bm.Context) {
	params := c.Request.Form
	pfStr := params.Get("pf")
	pf, _ := strconv.Atoi(pfStr)
	c.JSON(resSvc.URLMonitor(c, pf), nil)
}

func device(c *bm.Context) (mid int64, sid, buvid string) {
	midTemp, ok := c.Get("mid")
	buvid = c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	if !ok {
		if sidCookie, err := c.Request.Cookie("sid"); err == nil {
			sid = sidCookie.Value
		}
	} else {
		mid = midTemp.(int64)
	}
	return
}

func frontPage(ctx *bm.Context) {
	var (
		params = ctx.Request.Form
		resid  int64
		err    error
	)
	if resid, err = strconv.ParseInt(params.Get("resid"), 10, 64); err != nil || resid == 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	data, err := resSvc.FrontPage(ctx, resid)
	if err != nil {
		if resSvc.SLBRetry(err) {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(data, nil)
}

func pageHeader(ctx *bm.Context) {
	v := &struct {
		ResourceID int64 `form:"resource_id" validate:"min=1"`
	}{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	data, err := resSvc.PageHeader(ctx, v.ResourceID, ip)
	if err != nil {
		if resSvc.SLBRetry(err) {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(data, nil)
}
