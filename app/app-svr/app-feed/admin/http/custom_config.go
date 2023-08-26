package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-feed/admin/model/resource"
	util "go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/pkg/idsafe/bvid"
)

//nolint:deadcode,unused
func ccLog(ctx *bm.Context) {
	req := &model.CCLogReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(resourceSvc.CCLog(ctx, req))
	//nolint:gosimple
	return
}

func configList(ctx *bm.Context) {
	req := &model.CCListReq{}
	res := map[string]interface{}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.TP = model.CustonConfigTPArchive
	if req.OriginType == 1 && req.Oid == "" {
		res["message"] = "请输入要查找的稿件id，审核同步结果仅能单条查询"
		ctx.JSONMap(res, ecode.RequestErr)
		return
	}
	if req.Oid != "" {
		oid, _ := strconv.ParseInt(req.Oid, 0, 64)
		if oid == 0 {
			oid, _ = bvid.BvToAv(req.Oid)
			if oid == 0 {
				res["message"] = "搜索条件有误"
				ctx.JSONMap(res, ecode.RequestErr)
				return
			}
		}
		req.OidNum = oid
	}
	ctx.JSON(resourceSvc.ConfigList(ctx, req))
	//nolint:gosimple
	return
}

// configAdd is
func configAdd(ctx *bm.Context) {
	req := &model.CCAddReq{}
	res := map[string]interface{}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.TP = model.CustonConfigTPArchive
	req.OperatorID, req.Operator = util.UserInfo(ctx)
	oid, _ := strconv.ParseInt(req.Oid, 0, 64)
	if oid == 0 {
		oid, _ = bvid.BvToAv(req.Oid)
		if oid == 0 {
			res["message"] = "bvid转avid失败"
			ctx.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	req.OidNum = oid
	req.OriginType = 0
	rows, err := resourceSvc.CCAdd(ctx, req)
	if err != nil {
		res["message"] = "新增失败:" + err.Error()
		ctx.JSONMap(res, ecode.RequestErr)
		return
	} else if rows == 0 {
		res["message"] = "新增失败，该稿件已存在，请重新编辑"
		ctx.JSONMap(res, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, nil)
	//nolint:gosimple
	return
}

func configEdit(ctx *bm.Context) {
	req := &model.CCUpdateReq{}
	res := map[string]interface{}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.TP = model.CustonConfigTPArchive
	req.OperatorID, req.Operator = util.UserInfo(ctx)
	oid, _ := strconv.ParseInt(req.Oid, 0, 64)
	if oid == 0 {
		oid, _ = bvid.BvToAv(req.Oid)
		if oid == 0 {
			res["message"] = "bvid转avid失败"
			ctx.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	req.OidNum = oid
	req.OriginType = 0
	err := resourceSvc.CCUpdate(ctx, req)
	if err != nil {
		res["message"] = "编辑失败:" + err.Error()
		ctx.JSONMap(res, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, nil)
	//nolint:gosimple
	return
}

func configOpt(ctx *bm.Context) {
	req := &model.CCOptReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if req.State != model.CustomConfigStateEnable {
		req.State = model.CustomConfigStateDisable
	}
	req.OperatorID, req.Operator = util.UserInfo(ctx)
	ctx.JSON(nil, resourceSvc.CCOpt(ctx, req))
	//nolint:gosimple
	return
}

func getConfig(ctx *bm.Context) {
	req := &model.GetConfigReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(resourceSvc.GetConfig(ctx, req))
	//nolint:gosimple
	return
}

func configLog(ctx *bm.Context) {
	req := &model.CCLogReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(resourceSvc.CCLog(ctx, req))
	//nolint:gosimple
	return
}
