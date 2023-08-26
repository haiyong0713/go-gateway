package http

import (
	"unicode/utf8"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/space/admin/model"
	"go-gateway/app/web-svr/space/admin/util"
	"go-gateway/app/web-svr/space/ecode"
)

func topPhotoArcs(ctx *bm.Context) {
	v := new(struct {
		Mids []int64 `form:"mids,split" validate:"required,max=20,dive,gt=0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(spcSvc.TopPhotoArcs(ctx, v.Mids))
}

// 获取待审核列表
func getPhotoList(ctx *bm.Context) {
	req := new(struct {
		UploadTimeStart string `json:"upload_time_start" form:"upload_time_start"`
		UploadTimeEnd   string `json:"upload_time_end" form:"upload_time_end"`
		PlatFrom        int    `json:"platfrom" form:"platfrom"`
		MIDs            string `json:"mids" form:"mids"`
		Pn              int    `json:"pn" form:"pn"`
		Ps              int    `json:"ps" form:"ps"`
	})

	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	mids := util.ParamsFilterTo64(req.MIDs)

	if req.Pn == 0 {
		req.Pn = 1
	}
	if req.Ps == 0 {
		req.Ps = 10
	}

	params := &model.MemberUploadTopPhotoSearchParams{
		UploadTimeStart: req.UploadTimeStart,
		UploadTimeEnd:   req.UploadTimeEnd,
		PlatFrom:        req.PlatFrom,
		MIDs:            mids,
	}

	ctx.JSON(spcSvc.GetTopPhotoList(ctx, params, req.Pn, req.Ps))

}

// 通过审核
func passPhoto(ctx *bm.Context) {
	uname, uid := util.UserInfo(ctx)
	req := new(struct {
		IDs string `json:"ids" form:"ids"`
	})

	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ids := util.ParamsFilterTo64(req.IDs)
	ctx.JSON(nil, spcSvc.PassPhoto(ctx, ids, uname, uid))

}

// 驳回
func backPhoto(ctx *bm.Context) {
	uname, uid := util.UserInfo(ctx)
	req := new(struct {
		ID            int64  `json:"id" form:"id"`
		Reason        int    `json:"reason" form:"reason"`
		ReasonDefault string `json:"reason_default" form:"reason_default"`
		AccountBlock  int    `json:"account_block" form:"account_block"`
		ReasonType    int    `json:"reason_type" form:"reason_type"`
		BlockRemark   string `json:"block_remark" form:"block_remark"`
		BlockTime     int    `json:"block_time" form:"block_time"`
		BlockNotify   int    `json:"block_notify" form:"block_notify"`
		Moral         int    `json:"moral" form:"moral"`
	})

	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	if req.ID <= 0 {
		ctx.JSON(nil, ecode.TopPhotoRequestError)
		ctx.Abort()
		return
	}

	if req.Reason == 0 {
		ctx.JSON(nil, ecode.TopPhotoNoReason)
		ctx.Abort()
		return
	} else if req.Reason == -1 {
		if req.ReasonDefault == "" || utf8.RuneCount([]byte(req.ReasonDefault)) > 200 {
			ctx.JSON(nil, ecode.TopPhotoNoReason)
			ctx.Abort()
			return
		}
	}

	backPhotoParam := &model.BackPhotoParam{
		ID:            req.ID,
		Reason:        req.Reason,
		ReasonDefault: req.ReasonDefault,
	}

	accountBlockParam := &model.AccountBlockParam{
		AccountBlock: req.AccountBlock,
		ReasonType:   req.ReasonType,
		BlockRemark:  req.BlockRemark,
		BlockTime:    req.BlockTime,
		BlockNotify:  req.BlockNotify,
		Moral:        req.Moral,
	}

	ctx.JSON(nil, spcSvc.BackPhoto(backPhotoParam, accountBlockParam, uname, uid))
}

func rePass(ctx *bm.Context) {
	uname, uid := util.UserInfo(ctx)
	req := new(struct {
		ID int64 `json:"id" form:"id"`
	})

	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, spcSvc.RePass(req.ID, uname, uid))
}

// VipAuditLog 日志列表
func vipAuditLogList(ctx *bm.Context) {
	req := new(struct {
		UploadTimeStart string `json:"upload_time_start" form:"upload_time_start"`
		UploadTimeEnd   string `json:"upload_time_end" form:"upload_time_end"`
		AuditTimeStart  string `json:"audit_time_start" form:"audit_time_start"`
		AuditTimeEnd    string `json:"audit_time_end" form:"audit_time_end"`
		Status          string `json:"status" form:"status"`
		Platfrom        string `json:"platfrom" form:"platfrom"`
		MIDs            string `json:"mids" form:"mids"`
		Operator        string `json:"operator" form:"operator"`
		Pn              int    `json:"pn" form:"pn"`
		Ps              int    `json:"ps" form:"ps"`
	})

	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	mids := util.ParamsFilterTo64(req.MIDs)
	status := util.ParamsFilter(req.Status)
	platfrom := util.ParamsFilter(req.Platfrom)

	params := &model.VipAuditLogSearch{
		UploadTimeStart: req.UploadTimeStart,
		UploadTimeEnd:   req.UploadTimeEnd,
		AuditTimeStart:  req.AuditTimeStart,
		AuditTimeEnd:    req.AuditTimeEnd,
		Status:          status,
		Platfrom:        platfrom,
		MIDs:            mids,
		Operator:        req.Operator,
	}

	ctx.JSON(spcSvc.AuditLogList(ctx, params, req.Pn, req.Ps))
}

// actionLogList 获取行为日志
func actionLogList(ctx *bm.Context) {
	req := new(struct {
		IDs string `json:"ids" form:"ids"`
		Pn  int    `json:"pn" form:"pn"`
		Ps  int    `json:"ps" form:"ps"`
	})

	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Contetn-Type"))); err != nil {
		return
	}

	ctx.JSON(spcSvc.GetActionLogList(req.IDs, req.Pn, req.Ps))

}
