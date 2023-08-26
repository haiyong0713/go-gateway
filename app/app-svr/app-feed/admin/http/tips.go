package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/tips"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func openSearchTips(c *bm.Context) {
	var (
		err error
		req = &struct {
			IDs     []int64    `form:"ids,split"`
			StartTs xtime.Time `form:"start_ts"`
			EndTs   xtime.Time `form:"end_ts"`
			Page    int        `form:"page"`
			Size    int        `form:"size"`
		}{}
		res = &struct {
			Items []model.SearchTipRes `json:"items"`
			Page  common.Page          `json:"page"`
		}{}
		total int
	)
	if err = c.Bind(req); err != nil {
		return
	}

	res.Items, total, err = tipsSvc.OpenSearchTips(c, req.IDs, req.StartTs, req.EndTs, "", -1, 0, 0)

	res.Page = common.Page{
		Num:   1,
		Size:  total,
		Total: total,
	}

	c.JSON(res, err)
}

func searchTipList(c *bm.Context) {
	var (
		err error
		req = &struct {
			SearchWord string `form:"search_word"`
			Status     int    `form:"status"`
			Page       int    `form:"page" default:"1"`
			Size       int    `form:"size" default:"20"`
		}{}
		res = &struct {
			Items []model.SearchTipRes `json:"items"`
			Page  common.Page          `json:"page"`
		}{}
		total int
	)
	if err = c.Bind(req); err != nil {
		return
	}

	res.Items, total, err = tipsSvc.OpenSearchTips(c, nil, 0, 0, req.SearchWord, req.Status, req.Size, req.Page)

	res.Page = common.Page{
		Size:  req.Size,
		Num:   req.Page,
		Total: total,
	}

	c.JSON(res, err)
}

func searchTipAdd(c *bm.Context) {
	var (
		err error
		req = &model.SearchTipRes{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	uid, username := util.UserInfo(c)
	if username == "" {
		//nolint:ineffassign
		err = ecode.Error(ecode.RequestErr, "未登录")
		return
	}
	err = tipsSvc.SearchTipAdd(c, req, username, uid)

	c.JSON(nil, err)
}

func searchTipUpdate(c *bm.Context) {
	var (
		err error
		req = &model.SearchTipRes{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	uid, username := util.UserInfo(c)
	if username == "" {
		//nolint:ineffassign
		err = ecode.Error(ecode.RequestErr, "未登录")
		return
	}
	err = tipsSvc.SearchTipUpdate(c, req, username, uid)

	c.JSON(nil, err)
}

func searchTipOperate(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID     int64 `form:"id" json:"id" validate:"required"`
			Status int   `form:"status" json:"status"`
		}{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	uid, username := util.UserInfo(c)
	if username == "" {
		//nolint:ineffassign
		err = ecode.Error(ecode.RequestErr, "未登录")
		return
	}
	err = tipsSvc.SearchTipOperate(c, req.ID, req.Status, username, uid)

	c.JSON(nil, err)
}

func searchTipOffline(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID int64 `form:"id" json:"id" validate:"required"`
		}{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	uid, username := util.UserInfo(c)
	if username == "" {
		//nolint:ineffassign
		err = ecode.Error(ecode.RequestErr, "未登录")
		return
	}
	err = tipsSvc.SearchTipOperate(c, req.ID, 2, username, uid)

	c.JSON(nil, err)
}
