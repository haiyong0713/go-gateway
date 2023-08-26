package http

import (
	"encoding/json"
	"github.com/pkg/errors"
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"
	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/app/app-svr/app-feed/ecode"
	"strconv"
	"time"
)

// listFrontpages 获取
func listFrontpages(ctx *bm.Context) {
	req := &struct {
		ResourceID int64 `form:"resource_id" json:"resource_id"`
		Num        int64 `form:"pn" json:"pn" default:"1"`
		Size       int64 `form:"size" json:"size" default:"5"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if data, total, err := frontpageSvc.GetConfigs(ctx, req.ResourceID, req.Num, req.Size); err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	} else {
		res := &struct {
			Data  model.FrontpagesForFE `json:"data"`
			Pager model.Pager           `json:"pager"`
		}{
			Data:  data,
			Pager: model.Pager{Num: req.Num, Size: req.Size, Total: total},
		}
		ctx.JSON(res, nil)
	}
}

func listFrontpageMenus(ctx *bm.Context) {
	ctx.JSON(frontpageSvc.GetMenus())
}

// getFrontpageDetail 获取版头配置详情
func getFrontpageDetail(ctx *bm.Context) {
	var (
		resourceIDStr string
		resourceID    int64
		idStr         string
		id            int64
		exists        bool
		err           error
	)
	if idStr, exists = ctx.Params.Get("id"); !exists {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	id, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	if resourceIDStr, exists = ctx.Params.Get("resource"); !exists {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	resourceID, err = strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	if id == 0 {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	ctx.JSON(frontpageSvc.GetConfig(resourceID, id))
}

func addFrontpage(ctx *bm.Context) {
	var (
		err             error
		username        string
		configRuleBytes []byte
		configRuleStr   string
		stime           time.Time
		etime           time.Time
	)
	_, username = util.UserInfo(ctx)

	req := &struct {
		ResourceID       int64  `json:"resource_id" form:"resource_id" validate:"min=0"`
		Pic              string `json:"pic" form:"pic" validate:"required"`
		LitPic           string `json:"litpic" form:"litpic" validate:"required"`
		STime            string `json:"stime" form:"stime" validate:"required"`
		ETime            string `json:"etime" form:"etime" validate:"required"`
		IsCover          int    `json:"is_cover" form:"is_cover" validate:"min=0"`
		Style            int    `json:"style" form:"style"`
		Name             string `json:"name" form:"name"`
		URL              string `json:"url" form:"url"`
		LocPolicyGroupID int64  `json:"loc_policy_group_id" form:"loc_policy_group_id"`
		AType            int    `json:"atype" form:"atype"`
		IsSplitLayer     int    `json:"is_split_layer" form:"is_split_layer"`
		SplitLayer       string `json:"split_layer" form:"split_layer"`
	}{}
	if err = ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	// stime & etime
	if stime, err = time.ParseInLocation(model.DefaultTimeLayout, req.STime, time.Local); err != nil {
		ctx.JSON(nil, errors.Wrapf(err, "stime parse error"))
		ctx.Abort()
		return
	}
	if etime, err = time.ParseInLocation(model.DefaultTimeLayout, req.ETime, time.Local); err != nil {
		ctx.JSON(nil, errors.Wrapf(err, "etime parse error"))
		ctx.Abort()
		return
	}
	if stime.After(etime) {
		ctx.JSON(nil, xecode.Error(xecode.RequestErr, "stime不能在etime之后"))
		ctx.Abort()
		return
	}

	// validate dup
	var duplicated bool
	if duplicated, err = frontpageSvc.CheckConfigDuplicated(req.ResourceID, stime, etime, req.LocPolicyGroupID, 0); err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	} else if duplicated {
		ctx.JSON(nil, ecode.FrontPageConfigDuplicated)
		ctx.Abort()
		return
	}

	// rule
	configRule := model.ConfigRule{
		IsCover: req.IsCover,
		Style:   req.Style,
	}
	if configRuleBytes, err = json.Marshal(configRule); err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	configRuleStr = string(configRuleBytes)

	// build model
	toAddFrontpage := model.Config{
		ConfigName:       req.Name,
		ContractID:       model.DefaultContractID,
		ResourceID:       req.ResourceID,
		Pic:              req.Pic,
		LitPic:           req.LitPic,
		URL:              req.URL,
		STime:            xtime.Time(stime.Unix()),
		ETime:            xtime.Time(etime.Unix()),
		Rule:             configRuleStr,
		LocPolicyGroupID: req.LocPolicyGroupID,
		Atype:            int8(req.AType),
		IsSplitLayer:     req.IsSplitLayer,
		SplitLayer:       req.SplitLayer,
	}

	ctx.JSON(frontpageSvc.AddConfig(toAddFrontpage, username))
}

func editFrontpage(ctx *bm.Context) {
	var (
		err             error
		username        string
		configRuleBytes []byte
		configRuleStr   string
		stime           time.Time
		etime           time.Time
	)
	_, username = util.UserInfo(ctx)

	req := &struct {
		ResourceID       int64  `json:"resource_id" form:"resource_id" validate:"min=0"`
		ID               int64  `json:"id" form:"id" validate:"min=1"`
		Pic              string `json:"pic" form:"pic"`
		LitPic           string `json:"litpic" form:"litpic"`
		STime            string `json:"stime" form:"stime"`
		ETime            string `json:"etime" form:"etime"`
		IsCover          int    `json:"is_cover" form:"is_cover" validate:"min=0"`
		Style            int    `json:"style" form:"style"`
		Name             string `json:"name" form:"name"`
		URL              string `json:"url" form:"url"`
		LocPolicyGroupID int64  `json:"loc_policy_group_id" form:"loc_policy_group_id"`
		AType            int    `json:"atype" form:"atype"`
		IsSplitLayer     int    `json:"is_split_layer" form:"is_split_layer"`
		SplitLayer       string `json:"split_layer" form:"split_layer"`
	}{}
	if err = ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	updateMap := make(map[string]interface{})
	if req.Pic != "" {
		updateMap["pic"] = req.Pic
	}
	if req.LitPic != "" {
		updateMap["litpic"] = req.LitPic
	}
	updateMap["url"] = req.URL
	updateMap["config_name"] = req.Name
	updateMap["is_cover"] = req.IsCover
	updateMap["style"] = req.Style
	updateMap["loc_policy_group_id"] = req.LocPolicyGroupID
	updateMap["atype"] = req.AType
	updateMap["is_split_layer"] = req.IsSplitLayer
	updateMap["split_layer"] = req.SplitLayer
	// stime & etime
	if req.STime != "" {
		if stime, err = time.ParseInLocation(model.DefaultTimeLayout, req.STime, time.Local); err != nil {
			ctx.JSON(nil, err)
			ctx.Abort()
			return
		}
		updateMap["stime"] = stime
	}
	if req.ETime != "" {
		if etime, err = time.ParseInLocation(model.DefaultTimeLayout, req.ETime, time.Local); err != nil {
			ctx.JSON(nil, err)
			ctx.Abort()
			return
		}
		updateMap["etime"] = etime
	}

	// rule
	configRule := model.ConfigRule{
		IsCover: req.IsCover,
		Style:   req.Style,
	}
	if configRuleBytes, err = json.Marshal(configRule); err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	configRuleStr = string(configRuleBytes)
	updateMap["rule"] = configRuleStr

	ctx.JSON(nil, frontpageSvc.EditConfig(req.ID, req.ResourceID, updateMap, username))
}

func actionFrontpage(ctx *bm.Context) {
	var (
		err      error
		username string
	)
	_, username = util.UserInfo(ctx)

	req := &struct {
		ResourceID int64  `json:"resource_id" form:"resource_id" validate:"min=0"`
		ID         int64  `json:"id" form:"id" validate:"min=1"`
		Action     string `json:"action" form:"action" validate:"required"`
	}{}
	if err = ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, frontpageSvc.ActionConfig(req.ID, req.Action, username))
}

func listFrontpageHistory(ctx *bm.Context) {
	req := &struct {
		ResourceID int64 `form:"resource_id" json:"resource_id"`
		Num        int64 `form:"num" json:"num" default:"1"`
		Size       int64 `form:"size" json:"size" default:"5"`
	}{}
	if err := ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	data, total, err := frontpageSvc.GetConfigHistories(req.ResourceID, req.Num, req.Size)
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	res := &struct {
		Items []*model.Config `json:"items"`
		Pager model.Pager     `json:"pager"`
	}{
		Items: data,
		Pager: model.Pager{Num: req.Num, Size: req.Size, Total: total},
	}
	ctx.JSON(res, nil)
}

func listFrontpageLocationPolicies(ctx *bm.Context) {
	ctx.JSON(frontpageSvc.GetAllPolicyGroups(ctx))
}
