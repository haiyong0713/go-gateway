package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	mdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"regexp"
	"strings"
	"time"
)

var mobiAppRe = regexp.MustCompile(`mobi_app/([a-z]|[A-Z])*`)

func risk(ctx *bm.Context, mid int64, action string) *mdl.Base {
	rs := new(mdl.Base)
	request := ctx.Request
	ctx.Bind(rs)
	rs.UserAgent = request.UserAgent()
	rs.Referer = request.Referer()
	rs.IP = metadata.String(ctx, metadata.RemoteIP)
	if rs.Buvid == "" {
		if res, err := request.Cookie("Buvid"); err == nil {
			rs.Buvid = res.Value
		} else {
			if res, err := request.Cookie("buvid3"); err == nil {
				rs.Buvid = res.Value
			}
		}
	}
	rs.Ctime = time.Now().Format("2006-01-02 15:04:05")
	rs.Platform = ctx.Request.Form.Get("platform")
	if rs.Platform == "" {
		rs.Platform = mdl.PlatformWeb
	}
	rs.MID = mid
	rs.Action = action
	rs.API = ctx.Request.URL.Path
	rs.Origin = request.Header.Get("Origin")
	rs.EsTime = time.Now().Unix()
	rs.Build = ctx.Request.Form.Get("build")
	return rs
}

func riskParseUA(ctx *bm.Context, mid int64, action string) *mdl.Base {
	rs := new(mdl.Base)
	request := ctx.Request
	ctx.Bind(rs)
	ua := request.UserAgent()
	rs.UserAgent = ua
	rs.Referer = request.Referer()
	rs.IP = metadata.String(ctx, metadata.RemoteIP)
	if rs.Buvid == "" {
		if res, err := request.Cookie("Buvid"); err == nil {
			rs.Buvid = res.Value
		} else {
			if res, err := request.Cookie("buvid3"); err == nil {
				rs.Buvid = res.Value
			}
		}
	}
	rs.Ctime = time.Now().Format("2006-01-02 15:04:05")
	rs.Platform = ctx.Request.Form.Get("platform")
	if rs.Platform == "" {
		appStr := mobiAppRe.FindString(ua)
		if appStr != "" {
			tmp := strings.Split(appStr, "/")
			if len(tmp) >= 2 {
				rs.Platform = tmp[1]
			}
		}

		if rs.Platform == "" {
			rs.Platform = mdl.PlatformWeb
		}
	}
	rs.MID = mid
	rs.Action = action
	rs.API = ctx.Request.URL.Path
	rs.Origin = request.Header.Get("Origin")
	rs.EsTime = time.Now().Unix()
	rs.Build = ctx.Request.Form.Get("build")
	return rs
}
