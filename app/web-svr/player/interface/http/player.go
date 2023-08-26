package http

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/player/interface/model"
)

const (
	_platformH5    = "html5"
	_platformH5New = "html5_new"
)

var (
	emptyByte    = []byte{}
	_headerBuvid = "Buvid"
	_buvid       = "buvid3"
	_innerSign   = "innersign"
)

// nolint:gomnd
func player(c *bm.Context) {
	var (
		aid, mid  int64
		sid       model.Sid
		sidCookie *http.Cookie
		cidStr    string
		err       error
		ip        = metadata.String(c, metadata.RemoteIP)
	)
	v := new(model.PlayerArg)
	if err = c.Bind(v); err != nil {
		return
	}
	request := c.Request
	refer := request.Referer()
	// response header
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("X-Remote-IP", ip)
	if strings.Index(v.ID, "cid:") == 0 {
		cidStr = v.ID[4:]
		if v.Cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	if v.Cid <= 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// get mid
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	sidCookie, err = request.Cookie("sid")
	if err != nil || sidCookie.Value == "" {
		http.SetCookie(c.Writer, &http.Cookie{Name: "sid", Value: string(sid.Create()), Path: "/", Domain: ".bilibili.com"})
	}
	if sidCookie != nil {
		if sidCookie.Value != "" {
			sidStr := sidCookie.Value
			if !model.Sid(sidStr).Valid() {
				http.SetCookie(c.Writer, &http.Cookie{Name: "sid", Value: string(sid.Create()), Path: "/", Domain: ".bilibili.com"})
			}
		}
	}
	var (
		ips      = strings.Split(c.Request.Header.Get("X-Forwarded-For"), ",")
		cdnIP    string
		playByte []byte
	)
	if len(ips) >= 2 {
		cdnIP = ips[1]
	}
	buvid := request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	innerSign := ""
	innerCookie, _ := request.Cookie(_innerSign)
	if innerCookie != nil {
		innerSign = innerCookie.Value
	}
	now := time.Now()
	playSvr.ShowInfoc(c, ip, now, buvid, aid, mid)
	if playByte, err = playSvr.Player(c, mid, aid, v, cdnIP, refer, innerSign, now); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Bytes(http.StatusOK, "text/plain; charset=utf-8", playByte)
}

// nolint:gomnd
func playerV2(ctx *bm.Context) {
	var (
		mid int64
		sid model.Sid
		err error
	)
	arg := new(model.PlayerV2Arg)
	if err = ctx.Bind(arg); err != nil {
		return
	}
	if arg.Aid, err = bvArgCheck(arg.Aid, arg.Bvid); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	sidCookie, err := ctx.Request.Cookie("sid")
	if err != nil || sidCookie.Value == "" {
		http.SetCookie(ctx.Writer, &http.Cookie{Name: "sid", Value: string(sid.Create()), Path: "/", Domain: ".bilibili.com"})
	}
	if sidCookie != nil {
		if sidCookie.Value != "" {
			sidStr := sidCookie.Value
			if !model.Sid(sidStr).Valid() {
				http.SetCookie(ctx.Writer, &http.Cookie{Name: "sid", Value: string(sid.Create()), Path: "/", Domain: ".bilibili.com"})
			}
		}
	}
	buvid := ctx.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	arg.Buvid = buvid
	arg.Refer = ctx.Request.Referer()
	innerCookie, _ := ctx.Request.Cookie(_innerSign)
	if innerCookie != nil {
		arg.InnerSign = innerCookie.Value
	}
	ips := strings.Split(ctx.Request.Header.Get("X-Forwarded-For"), ",")
	if len(ips) >= 2 {
		arg.CdnIP = ips[1]
	}
	// get mid
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	now := time.Now()
	arg.Now = now
	playSvr.ShowInfoc(ctx, metadata.String(ctx, metadata.RemoteIP), now, buvid, arg.Aid, mid)
	data, err := playSvr.PlayerV2(ctx, arg, mid)
	if err != nil {
		if playSvr.SLBRetry(err) {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(data, nil)
}

func carousel(c *bm.Context) {
	// response header
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	type msg struct {
		Items []*model.Item `xml:"item"`
	}
	var (
		items []*model.Item
		err   error
	)
	if items, err = playSvr.Carousel(c); err != nil {
		c.Bytes(http.StatusOK, "text/xml; charset=utf-8", emptyByte)
		return
	}
	result := msg{Items: items}
	output, err := xml.MarshalIndent(result, "  ", "    ")
	if err != nil {
		output = emptyByte
	}
	c.Bytes(http.StatusOK, "text/xml; charset=utf-8", output)
}

func policy(c *bm.Context) {
	var (
		id  int64
		mid int64
		err error
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params := c.Request.Form
	idStr := params.Get("id")
	if id, err = strconv.ParseInt(idStr, 10, 64); err != nil || id < 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(playSvr.Policy(c, id, mid))
}

func view(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(playSvr.View(c, aid))
}

func matPage(c *bm.Context) {
	c.JSON(playSvr.Matsuri(c, time.Now()), nil)
}

func playerCardClick(ctx *bm.Context) {
	arg := new(model.PlayerCardClickArg)
	if err := ctx.Bind(arg); err != nil {
		return
	}
	arg.Buvid = ctx.Request.Header.Get(_headerBuvid)
	if arg.Buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			arg.Buvid = cookie.Value
		}
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, playSvr.PlayerCardClick(ctx, arg, mid))
}

func onlineTotal(ctx *bm.Context) {
	v := new(struct {
		Aid      int64  `form:"aid"`
		Bvid     string `form:"bvid"`
		Cid      int64  `form:"cid" validate:"min=1"`
		Business int32  `form:"business" default:"1" validate:"min=1,max=2"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	aid, err := bvArgCheck(v.Aid, v.Bvid)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	buvid := ctx.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}
	ctx.JSON(playSvr.OnlineTotal(ctx, mid, buvid, aid, v.Cid, v.Business))
}
