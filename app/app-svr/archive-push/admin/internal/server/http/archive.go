package http

import (
	"strings"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
)

func archivesPushed(ctx *bm.Context) {
	archivesPushedReq := &struct {
		Pn       int    `form:"pn" json:"pn" validate:"min=1" default:"1"`
		Ps       int    `form:"ps" json:"ps" validate:"min=1" default:"20"`
		BVIDs    string `form:"bvids" json:"bvids"`
		VendorID int64  `form:"vendorId" json:"vendorId"`
		PushType int32  `form:"pushType" json:"pushType"`
	}{}
	if err := ctx.BindWith(archivesPushedReq, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if archivesPushedReq.Pn == 0 {
		archivesPushedReq.Pn = 1
	}
	if archivesPushedReq.Ps == 0 {
		archivesPushedReq.Ps = 20
	}
	bvids := make([]string, 0)
	if len(archivesPushedReq.BVIDs) > 0 {
		_bvidSplit := strings.Split(archivesPushedReq.BVIDs, ",")
		for _, bvid := range _bvidSplit {
			bvids = append(bvids, bvid)
		}
	}

	archives, total, err := svc.GetPushedArchives(archivesPushedReq.VendorID, bvids, archivesPushedReq.PushType, archivesPushedReq.Pn, archivesPushedReq.Ps)
	res := struct {
		Items []*model.ArchivePushDetailByBVID `json:"items"`
		Pager *model.Page                      `json:"pager"`
	}{
		Pager: &model.Page{
			Num:   archivesPushedReq.Pn,
			Size:  archivesPushedReq.Ps,
			Total: int64(total),
		},
	}
	res.Items = archives
	ctx.JSON(res, err)
}

func archiveWithdraw(ctx *bm.Context) {
	var err error
	username, uid := util.UserInfo(ctx)
	archiveWithdrawReq := &struct {
		VendorID int64  `form:"vendorId" json:"vendorId" validate:"min=1"`
		BVID     string `form:"bvid" json:"bvid" validate:"required"`
		Reason   string `form:"reason" json:"reason" validate:"required"`
	}{}
	if err := ctx.BindWith(archiveWithdrawReq, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	err = svc.WithdrawArchive(archiveWithdrawReq.BVID, archiveWithdrawReq.Reason, archiveWithdrawReq.VendorID, true, username, uid)
	ctx.JSON(nil, err)
}

func apiArchivesStatusSync(ctx *bm.Context) {
	req := &model.SyncArchiveStatusReq{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, svc.SyncArchiveStatus(*req))
}
