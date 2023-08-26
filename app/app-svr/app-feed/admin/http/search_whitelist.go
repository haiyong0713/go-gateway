package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/search_whitelist"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// 添加白名单内稿件的时候，传入id列表，返回对应的稿件预览列表
func searchWhiteListArchivePreview(c *bm.Context) {
	var (
		req struct {
			//nolint:staticcheck
			AvidList []string `json:"avid_list,split" form:"avid_list,split"`
		}
		res struct {
			Items []*model.WhiteListArchiveItem `json:"items"`
		}
		avidList []int64
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	for _, bvidStr := range req.AvidList {
		if avid, err := bvav.ToAvInt(bvidStr); err == nil {
			avidList = append(avidList, avid)
		}
	}
	res.Items, err = searchWhitelist.SearchWhiteListArchivePreview(c, util.Int64ArrayDedup(avidList))
	c.JSON(res, err)
}

// 白名单操作：一审/二审通过、驳回、下线、删除
func searchWhiteListOption(c *bm.Context) {
	var (
		req struct {
			ID        int64  `json:"id" form:"id"`
			Operation string `json:"operation" form:"operation"`
		}
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	_, username := util.UserInfo(c)
	err = searchWhitelist.SearchWhiteListOption(c, req.ID, req.Operation, username)
	c.JSON(nil, err)
}

// 编辑
func searchWhiteListEdit(c *bm.Context) {
	var (
		req struct {
			ID int64 `json:"id" form:"id" validate:"required"`
			//nolint:staticcheck
			SearchWord []string   `json:"search_word,split" form:"search_word,split" validate:"required"`
			STime      xtime.Time `json:"stime" form:"stime"`
			ETime      xtime.Time `json:"etime" form:"etime"`
			//nolint:staticcheck
			AvidList []string `json:"avid_list,split" form:"avid_list,split" validate:"max=100"`
		}
		avidList []int64
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	_, username := util.UserInfo(c)
	for _, bvidStr := range req.AvidList {
		if avid, err := bvav.ToAvInt(bvidStr); err == nil {
			avidList = append(avidList, avid)
		}
	}
	err = searchWhitelist.SearchWhiteListEdit(c, req.ID, util.StringArrayDedup(req.SearchWord), util.Int64ArrayDedup(avidList), req.STime, req.ETime, username, true)
	c.JSON(nil, err)
}

// 添加
func searchWhiteListAdd(c *bm.Context) {
	var (
		req struct {
			//nolint:staticcheck
			SearchWord []string   `json:"search_word,split" form:"search_word,split" validate:"required"`
			STime      xtime.Time `json:"stime" form:"stime"`
			ETime      xtime.Time `json:"etime" form:"etime"`
			//nolint:staticcheck
			AvidList []string `json:"avid_list,split" form:"avid_list,split" validate:"max=100"`
		}
		avidList []int64
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	_, username := util.UserInfo(c)
	for _, bvidStr := range req.AvidList {
		if avid, err := bvav.ToAvInt(bvidStr); err == nil {
			avidList = append(avidList, avid)
		}
	}
	err = searchWhitelist.SearchWhiteListAdd(c, util.StringArrayDedup(req.SearchWord), util.Int64ArrayDedup(avidList), req.STime, req.ETime, username, true, 0)
	c.JSON(nil, err)
}

// 根据白名单配置id获取稿件列表
func searchWhiteListArchiveList(c *bm.Context) {
	var (
		req struct {
			ID int64 `json:"id" form:"id"`
		}
		res struct {
			Items []*model.WhiteListArchiveItem `json:"items"`
		}
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	res.Items, err = searchWhitelist.SearchWhiteListArchiveList(c, req.ID)
	c.JSON(res, err)
}

// 获取白名单配置列表
func searchWhiteList(c *bm.Context) {
	var (
		req struct {
			SearchWord string     `json:"search_word" form:"search_word"`
			CUser      string     `json:"c_user" form:"c_user"`
			STime      xtime.Time `json:"stime" form:"stime"`
			ETime      xtime.Time `json:"etime" form:"etime"`
			Status     int        `json:"status" form:"status"`
			Page       int        `json:"page" form:"page" default:"1" validate:"min=1"`
			Size       int        `json:"size" form:"size" default:"20" validate:"min=1"`
		}
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	var (
		res = struct {
			Items []*model.WhiteListItemWithQueryAndArchive `json:"items"`
			Page  common.Page                               `json:"page"`
		}{
			Page: common.Page{
				Num:  req.Page,
				Size: req.Size,
			},
		}
	)
	res.Page.Total, res.Items, err = searchWhitelist.SearchWhiteList(c, req.Status, req.STime, req.ETime, req.CUser, req.SearchWord, true, req.Size, req.Page)
	c.JSON(res, err)
}

// 给网关和AI使用的开放接口，获取配置列表
func openSearchWhiteList(c *bm.Context) {
	var (
		req struct {
			STime xtime.Time `json:"start_ts" form:"start_ts"`
			ETime xtime.Time `json:"end_ts" form:"end_ts"`
			Page  int        `json:"page" form:"page" default:"1" validate:"min=1"`
			Size  int        `json:"size" form:"size" default:"20" validate:"min=1"`
		}
	)
	err := c.BindWith(&req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type")))
	if err != nil {
		return
	}
	var (
		res = struct {
			Items []*model.WhiteListItemWithQueryAndArchive `json:"items"`
			Page  common.Page                               `json:"page"`
		}{
			Page: common.Page{
				Num:  req.Page,
				Size: req.Size,
			},
		}
	)
	res.Page.Total, res.Items, err = searchWhitelist.SearchWhiteList(c, 0, req.STime, req.ETime, "", "", false, req.Size, req.Page)
	c.JSON(res, err)
}
