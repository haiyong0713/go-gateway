package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-feed/admin/model/anti_crawler"
)

func antiCrawlerUserLog(ctx *bm.Context) {
	v := &model.UserLogParam{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.Buvid == "" && v.Mid == 0 {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "buvid和mid不能都为空"))
		return
	}
	switch len(v.TimeRange) {
	case 0:
	case 1:
		v.Stime = v.TimeRange[0]
	default:
		v.Stime = v.TimeRange[0]
		v.Etime = v.TimeRange[1]

	}
	if v.Etime < v.Stime {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "结束时间不能大于开始时间"))
	}
	data, err := antiSvr.UserLog(ctx, v)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["items"] = data
	res["hasNext"] = len(data) == v.PerPage
	ctx.JSON(res, nil)
}

func antiCrawlerBusinessConfigList(ctx *bm.Context) {
	req := new(model.BusinessConfigListReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	data, err := antiSvr.BusinessConfigList(ctx, req)
	if err != nil {
		ctx.JSONMap(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["items"] = data
	ctx.JSON(res, nil)
}

func antiCrawlerBusinessConfigUpdate(ctx *bm.Context) {
	req := new(model.BusinessConfigUpdateReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	if req.Forever == 0 && req.Datetime == 0 {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "datetime参数值非法"))
		return
	}
	ctx.JSON(nil, antiSvr.BusinessConfigUpdate(ctx, req))
}

func antiCrawlerBusinessConfigDelete(ctx *bm.Context) {
	req := new(model.BusinessConfigDeleteReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(nil, antiSvr.BusinessConfigDelete(ctx, req))
}
