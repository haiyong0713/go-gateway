package http

import (
	"strconv"
	"strings"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
)

// authors 查询作者
func authors(ctx *bm.Context) {
	authorsReq := &struct {
		IDs                 string `form:"ids" json:"ids"`
		CUser               string `form:"cuser" json:"cuser"`
		MID                 int64  `form:"mid" json:"mid"`
		AuthorizationStatus int32  `form:"authorizationStatus" json:"authorizationStatus"`
		BindStatus          int32  `form:"bindStatus" json:"bindStatus"`
		VerificationStatus  int32  `form:"verificationStatus" json:"verificationStatus"`
		PushVendorID        int64  `form:"pushVendorId" json:"pushVendorId"`
		Pn                  int    `form:"pn" json:"pn" validate:"min=1" default:"1"`
		Ps                  int    `form:"ps" json:"ps" validate:"min=1" default:"20"`
	}{}
	if err := ctx.BindWith(authorsReq, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if authorsReq.Pn == 0 {
		authorsReq.Pn = 1
	}
	if authorsReq.Ps == 0 {
		authorsReq.Ps = 20
	}

	ids := make([]int64, 0)
	if len(authorsReq.IDs) > 0 {
		idStrs := strings.Split(authorsReq.IDs, ",")
		for _, idStr := range idStrs {
			if id, _err := strconv.ParseInt(idStr, 10, 64); _err == nil {
				ids = append(ids, id)
			}
		}
	}

	type Res struct {
		Items []*model.ArchivePushAuthorX `json:"items"`
		Pager *model.Page                 `json:"pager"`
	}
	var res *Res

	authorList, total, err := svc.GetAuthorsByPage(ids, authorsReq.MID, authorsReq.AuthorizationStatus, authorsReq.BindStatus, authorsReq.VerificationStatus, authorsReq.PushVendorID, authorsReq.CUser, authorsReq.Pn, authorsReq.Ps)
	if err == nil {
		res = &Res{
			Items: authorList,
			Pager: &model.Page{
				Num:   authorsReq.Pn,
				Size:  authorsReq.Ps,
				Total: total,
			},
		}
	}
	ctx.JSON(res, err)
}

// addAuthors 添加作者
func addAuthors(ctx *bm.Context) {
	username, uid := util.UserInfo(ctx)
	req := &struct {
		PushVendorID int64  `json:"pushVendorId" form:"pushVendorId" validate:"min=1"`
		FileURL      string `json:"fileUrl" form:"fileUrl" validate:"required"`
		ActivityID   int64  `json:"activityId" form:"activityId" validate:"min=1"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(svc.UploadAuthors(req.PushVendorID, req.FileURL, req.ActivityID, username, uid))
}

// removeAuthor 移除作者
func removeAuthor(ctx *bm.Context) {
	username, uid := util.UserInfo(ctx)
	req := &struct {
		PushVendorID int64  `json:"pushVendorId" form:"pushVendorId" validate:"min=1"`
		MID          int64  `json:"mid" form:"mid" validate:"min=1"`
		Reason       string `json:"reason" form:"reason" validate:"required"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, svc.RemoveAuthor(req.PushVendorID, req.MID, req.Reason, false, username, uid))
}

func authorPushBatches(ctx *bm.Context) {
	req := &struct {
		IDs   string `form:"ids" json:"ids"`
		CUser string `form:"cuser" json:"cuser"`
		Pn    int    `form:"pn" json:"pn" validate:"min=1" default:"1"`
		Ps    int    `form:"ps" json:"ps" validate:"min=1" default:"20"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if req.Pn == 0 {
		req.Pn = 1
	}
	if req.Ps == 0 {
		req.Ps = 20
	}

	ids := make([]int64, 0)
	if len(req.IDs) > 0 {
		idStrs := strings.Split(req.IDs, ",")
		for _, idStr := range idStrs {
			if id, _err := strconv.ParseInt(idStr, 10, 64); _err == nil {
				ids = append(ids, id)
			}
		}
	}
	pushes, total, err := svc.GetAuthorPushFullsByPage(ids, req.CUser, req.Pn, req.Ps)
	res := struct {
		Items []*model.ArchivePushAuthorPushFull `json:"items"`
		Pager *model.Page                        `json:"pager"`
	}{
		Items: pushes,
		Pager: &model.Page{
			Num:   req.Pn,
			Size:  req.Ps,
			Total: total,
		},
	}
	ctx.JSON(res, err)
}

func createAuthorPushBatches(ctx *bm.Context) {
	username, uid := util.UserInfo(ctx)
	req := &struct {
		PushVendorID int64  `json:"pushVendorId" form:"pushVendorId" validate:"min=1"`
		Tags         string `json:"tags" form:"tags"`
		DelayMinutes int32  `json:"delayMinutes" form:"delayMinutes"`
		Authorized   bool   `json:"authorized" form:"authorized"`
		Binded       bool   `json:"binded" form:"binded"`
		Verified     bool   `json:"verified" form:"verified"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	pushConditions := []*model.ArchivePushAuthorPushCondition{
		{
			Type:  model.ArchivePushAuthorPushConditionTypeAuthorized,
			Op:    model.ArchivePushAuthorPushConditionOpEquals,
			Value: req.Authorized,
		},
		{
			Type:  model.ArchivePushAuthorPushConditionTypeBinded,
			Op:    model.ArchivePushAuthorPushConditionOpEquals,
			Value: req.Binded,
		},
		{
			Type:  model.ArchivePushAuthorPushConditionTypeVerified,
			Op:    model.ArchivePushAuthorPushConditionOpEquals,
			Value: req.Verified,
		},
	}

	ctx.JSON(svc.AddAuthorPush(req.PushVendorID, req.Tags, req.DelayMinutes, pushConditions, false, username, uid))
}

func editAuthorPushBatches(ctx *bm.Context) {
	username, uid := util.UserInfo(ctx)
	req := &struct {
		ID           int64  `json:"id" form:"id" validate:"min=1"`
		Tags         string `json:"tags" form:"tags"`
		DelayMinutes int32  `json:"delayMinutes" form:"delayMinutes"`
		Authorized   bool   `json:"authorized" form:"authorized"`
		Binded       bool   `json:"binded" form:"binded"`
		Verified     bool   `json:"verified" form:"verified"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	pushConditions := []*model.ArchivePushAuthorPushCondition{
		{
			Type:  model.ArchivePushAuthorPushConditionTypeAuthorized,
			Op:    model.ArchivePushAuthorPushConditionOpEquals,
			Value: req.Authorized,
		},
		{
			Type:  model.ArchivePushAuthorPushConditionTypeBinded,
			Op:    model.ArchivePushAuthorPushConditionOpEquals,
			Value: req.Binded,
		},
		{
			Type:  model.ArchivePushAuthorPushConditionTypeVerified,
			Op:    model.ArchivePushAuthorPushConditionOpEquals,
			Value: req.Verified,
		},
	}

	ctx.JSON(nil, svc.EditAuthorPush(req.ID, req.Tags, req.DelayMinutes, pushConditions, username, uid))
}

func inactivateAuthorPushBatches(ctx *bm.Context) {
	username, uid := util.UserInfo(ctx)
	req := &struct {
		ID     int64  `json:"id" form:"id" validate:"min=1"`
		Reason string `json:"reason" form:"reason" validate:"required"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, svc.InactivateAuthorPush(req.ID, req.Reason, false, username, uid))
}

// api

// apiGetAuthors 网关查询作者绑定信息
func apiGetAuthors(ctx *bm.Context) {
	apiGetAuthorsReq := &struct {
		VendorID int64  `json:"vendorId" form:"vendorId" validate:"required"`
		MID      int64  `json:"mid" form:"mid"`
		OpenID   string `json:"openId" form:"openId"`
	}{}
	if err := ctx.BindWith(apiGetAuthorsReq, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	authorList, err := svc.GetAuthorsByUser(apiGetAuthorsReq.VendorID, apiGetAuthorsReq.MID, apiGetAuthorsReq.OpenID)
	ctx.JSON(authorList, err)
}

// apiAuthorsBindingSync 外部回传作者绑定信息
func apiAuthorsBindingSync(ctx *bm.Context) {
	req := &model.SyncAuthorBindingReq{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, svc.SyncAuthorBinding(*req))
}
