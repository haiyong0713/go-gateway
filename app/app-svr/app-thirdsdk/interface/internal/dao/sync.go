package dao

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_userBindSyncURL  = "/x/admin/archive-push/api/users/bindingSync"
	_arcStatusSyncURL = "/x/admin/archive-push/api/archives/statusSync"
)

func (d *dao) UserBindSync(ctx context.Context, vendorID string, param *model.UserBindParam) error {
	params := url.Values{}
	params.Set("vendorId", vendorID)
	params.Set("platform", param.Platform)
	params.Set("bOpenId", param.BOpenID)
	params.Set("oOpenId", param.OOpenID)
	params.Set("action", param.Action)
	params.Set("actionTime", param.ActionTime)
	params.Set("actionMsg", param.ActionMsg)
	var res struct {
		Code int `json:"code"`
	}
	if err := d.httpMgr.Post(ctx, d.userBindSync, "", params, &res); err != nil {
		return err
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrap(ecode.Int(res.Code), d.userBindSync+"?"+params.Encode())
	}
	return nil
}

func (d *dao) ArcStatusSync(ctx context.Context, vendorID string, param *model.ArcStatusParam) error {
	params := url.Values{}
	params.Set("vendorId", vendorID)
	params.Set("platform", param.Platform)
	params.Set("bvid", param.Bvid)
	params.Set("ovid", param.Ovid)
	params.Set("status", param.Status)
	params.Set("statusTime", param.StatusTime)
	params.Set("statusMsg", param.StatusMsg)
	var res struct {
		Code int `json:"code"`
	}
	if err := d.httpMgr.Post(ctx, d.arcStatusSync, "", params, &res); err != nil {
		return err
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrap(ecode.Int(res.Code), d.arcStatusSync+"?"+params.Encode())
	}
	return nil
}
