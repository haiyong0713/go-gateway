package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/api"
	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

/**
网关
*/
// 给网关调用，获取正在生效中的闪屏策略
func openSplashList(c *bm.Context) {
	var (
		err error
		res *splashModel.GatewayConfig
	)

	res, err = splashSvc.GetSplashConfigOnline()

	c.JSON(res, err)
	//nolint:gosimple
	return
}

/**
物料编辑
*/
// 新增物料
func splashImageAdd(c *bm.Context) {
	var (
		err error
		req = &struct {
			ImageName      string `json:"img_name" form:"img_name" validate:"required"`
			ImageUrl       string `json:"img_url" form:"img_url"`
			Mode           int    `json:"mode" form:"mode" validate:"required"`
			ImageUrlNormal string `json:"img_url_normal" form:"img_url_normal"`
			ImageUrlFull   string `json:"img_url_full" form:"img_url_full"`
			ImageUrlPad    string `json:"img_url_pad" form:"img_url_pad"`
			LogoShow       int    `json:"logo_show" form:"logo_show" validate:"required"`
			LogoMode       int    `json:"logo_mode" form:"logo_mode" validate:"required"`
			LogoImageUrl   string `json:"logo_img_url" form:"logo_img_url"`
		}{}
		res struct {
			ID int64 `json:"id"`
		}
	)

	uid, username := util.UserInfo(c)

	if err = c.Bind(req); err != nil {
		return
	}

	param := &splashModel.SplashScreenImage{
		ImageName:      req.ImageName,
		ImageUrl:       req.ImageUrl,
		Mode:           req.Mode,
		ImageUrlNormal: req.ImageUrlNormal,
		ImageUrlFull:   req.ImageUrlFull,
		ImageUrlPad:    req.ImageUrlPad,
		LogoMode:       req.LogoMode,
		LogoHideFlag:   splashModel.LogoShow,
		LogoImageUrl:   req.LogoImageUrl,
	}
	//nolint:gomnd
	if req.LogoShow == 2 {
		param.LogoHideFlag = splashModel.LogoHide
	}

	res.ID, err = splashSvc.AddSplashImage(param, username, uid)

	c.JSON(res, err)
	//nolint:gosimple
	return
}

// 修改物料
func splashImageEdit(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID             int64  `json:"id" form:"id" validate:"required"`
			ImageName      string `json:"img_name" form:"img_name"`
			ImageUrl       string `json:"img_url" form:"img_url"`
			Mode           int    `json:"mode" form:"mode"`
			ImageUrlNormal string `json:"img_url_normal" form:"img_url_normal"`
			ImageUrlFull   string `json:"img_url_full" form:"img_url_full"`
			ImageUrlPad    string `json:"img_url_pad" form:"img_url_pad"`
			LogoShow       int    `json:"logo_show" form:"logo_show"`
			LogoMode       int    `json:"logo_mode" form:"logo_mode"`
			LogoImageUrl   string `json:"logo_img_url" form:"logo_img_url"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.Bind(req); err != nil {
		return
	}

	param := &splashModel.SplashScreenImage{
		ID:             req.ID,
		ImageName:      req.ImageName,
		ImageUrl:       req.ImageUrl,
		Mode:           req.Mode,
		ImageUrlNormal: req.ImageUrlNormal,
		ImageUrlFull:   req.ImageUrlFull,
		ImageUrlPad:    req.ImageUrlPad,
		LogoMode:       req.LogoMode,
		LogoShowFlag:   req.LogoShow,
		LogoImageUrl:   req.LogoImageUrl,
	}
	//nolint:gomnd
	if req.LogoShow == 2 {
		param.LogoHideFlag = splashModel.LogoHide
	}

	err = splashSvc.UpdateSplashImage(param, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 删除物料
func splashImageDelete(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID int64 `json:"id" form:"id" validate:"required"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.Bind(req); err != nil {
		return
	}

	err = splashSvc.DeleteSplashImage(req.ID, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 物料列表
func splashImageList(c *bm.Context) {
	var (
		err error
		res []*splashModel.SplashScreenImage
	)

	res, err = splashSvc.GetSplashImageList()

	c.JSON(res, err)
	//nolint:gosimple
	return
}

/**
配置
*/
// 修改默认配置审核状态
func splashConfigAudit(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID         int64 `json:"id" form:"id" validate:"required"`
			AuditState int   `json:"audit_state" form:"audit_state"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.Bind(req); err != nil {
		return
	}

	err = splashSvc.UpdateAuditState(req.ID, req.AuditState, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 新建配置
func splashConfigAdd(c *bm.Context) {
	var (
		err error
		req = &struct {
			IsImmediately  int        `json:"immediately" form:"immediately" default:"0"`
			STime          xtime.Time `json:"stime" form:"stime"`
			ETime          xtime.Time `json:"etime" form:"etime"`
			ShowMode       int        `json:"show_mode" form:"show_mode" validate:"required"`
			ForceShowTimes int32      `json:"force_show_times" form:"force_show_times" default:"0"`
			ConfigJson     string     `json:"config_json" form:"config_json" validate:"required"`
		}{}
		res struct {
			ID int64 `json:"id"`
		}
	)

	uid, username := util.UserInfo(c)

	if err = c.Bind(req); err != nil {
		return
	}

	res.ID, err = splashSvc.AddSplashConfig(&splashModel.SplashScreenConfig{
		STime:          req.STime,
		ETime:          req.ETime,
		ShowMode:       req.ShowMode,
		ConfigJson:     req.ConfigJson,
		IsImmediately:  req.IsImmediately,
		ForceShowTimes: req.ForceShowTimes,
	}, username, uid)

	c.JSON(res, err)
	//nolint:gosimple
	return
}

// 编辑配置
func splashConfigEdit(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID             int64      `json:"id" form:"id" validate:"required"`
			IsImmediately  int        `json:"immediately" form:"immediately" default:"0"`
			STime          xtime.Time `json:"stime" form:"stime"`
			ETime          xtime.Time `json:"etime" form:"etime"`
			ShowMode       int        `json:"show_mode" form:"show_mode" validate:"required"`
			ForceShowTimes int32      `json:"force_show_times" form:"force_show_times" default:"0"`
			ConfigJson     string     `json:"config_json" form:"config_json" validate:"required"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.Bind(req); err != nil {
		return
	}

	param := &splashModel.SplashScreenConfig{
		ID:             req.ID,
		STime:          req.STime,
		ETime:          req.ETime,
		ShowMode:       req.ShowMode,
		ConfigJson:     req.ConfigJson,
		IsImmediately:  req.IsImmediately,
		ForceShowTimes: req.ForceShowTimes,
	}

	err = splashSvc.UpdateSplashConfig(param, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 获取配置列表
func splashConfigList(c *bm.Context) {
	var (
		err error
		req = &struct {
			ShowMode int   `json:"show_mode" form:"show_mode" validate:"required"`
			Ps       int32 `json:"ps" form:"ps" default:"20"`
			Pn       int32 `json:"pn" form:"pn" default:"1"`
		}{}
		listRes []*splashModel.SplashScreenConfig
		total   int32
	)

	if err = c.Bind(req); err != nil {
		return
	}

	listRes, total, err = splashSvc.GetSplashConfigList(req.ShowMode, req.Ps, req.Pn)

	res := &splashModel.SplashConfigListWithPager{
		Pager: &splashModel.Pager{
			Pn: req.Pn,
			Ps: req.Ps,
		},
	}

	res.List = listRes
	res.Pager.Total = total

	c.JSON(res, err)
	//nolint:gosimple
	return
}

// splashConfigSelectAudit 修改自选配置审核状态
// Method: POST
// Path: /x/admin/feed/splash/config/select/audit
func splashConfigSelectAudit(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID         int64  `json:"id" form:"id" validate:"required"`
			AuditState string `json:"audit_state" form:"audit_state"`
		}{}
		auditState api.SplashScreenConfigAuditStatus_Enum
	)

	uid, username := util.UserInfo(c)

	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	if state, exists := api.SplashScreenConfigAuditStatus_Enum_value[req.AuditState]; exists {
		auditState = api.SplashScreenConfigAuditStatus_Enum(state)
	} else {
		c.JSON(nil, xecode.Error(xecode.RequestErr, "审核状态不正确"))
		c.Abort()
		return
	}

	err = splashSvc.UpdateSelectConfigAuditState(req.ID, auditState, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// splashConfigSelectSave 批量保存自选闪屏配置
// Method: POST
// Path: /x/admin/feed/splash/config/select/save
func splashConfigSelectSave(ctx *bm.Context) {
	var (
		err error
		req = &struct {
			Items []*splashModel.SelectConfigForSaving `json:"items"`
		}{}
	)

	uid, username := util.UserInfo(ctx)

	if err = ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(splashSvc.SaveSelectConfigs(ctx, req.Items, username, uid))
	//nolint:gosimple
	return
}

// splashConfigSelectList 获取自选闪屏配置列表
// Method: GET
// Path: /x/admin/feed/splash/config/select/list
func splashConfigSelectList(c *bm.Context) {
	var (
		err error
		req = &struct {
			Ps         int32  `json:"ps" form:"ps" default:"20"`
			Pn         int32  `json:"pn" form:"pn" default:"1"`
			State      string `json:"state" form:"state"`
			Sorting    string `json:"sorting" form:"sorting"`
			CategoryID int64  `json:"category_id" form:"category_id"`
			ImageID    int64  `json:"image_id" form:"image_id"`
		}{}
		configState api.SplashScreenConfigState_Enum
		listRes     []*splashModel.SelectConfig
		total       int32
	)

	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	if req.State != "" {
		if _configState, exists := api.SplashScreenConfigState_Enum_value[req.State]; exists {
			configState = api.SplashScreenConfigState_Enum(_configState)
		}
	} else {
		configState = -1
	}

	listRes, total, err = splashSvc.GetSelectConfigs(req.ImageID, req.CategoryID, configState, req.Sorting, req.Pn, req.Ps)

	res := &splashModel.SelectConfigsWithPager{
		Items: listRes,
		Pager: &splashModel.Pager{
			Pn:    req.Pn,
			Ps:    req.Ps,
			Total: total,
		},
	}

	c.JSON(res, err)
	//nolint:gosimple
	return
}

// splashConfigSelectSortBoundary 置顶/置底配置
// Method: POST
// Path: /x/admin/feed/splash/config/select/sortBoundary
func splashConfigSelectSortBoundary(ctx *bm.Context) {
	var (
		err error
		req = &struct {
			ID       int64  `json:"id" form:"id" validate:"min=1"`
			SortType string `json:"sort_type" form:"sort_type" validate:"required"`
		}{}
	)

	uid, username := util.UserInfo(ctx)

	if err = ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, splashSvc.UpdateSelectConfigSort(req.ID, req.SortType, username, uid))
	//nolint:gosimple
	return
}

// splashConfigSelectDelete 删除配置
// Method: POST
// Path: /x/admin/feed/splash/config/select/delete
func splashConfigSelectDelete(ctx *bm.Context) {
	var (
		err error
		req = &struct {
			IDs []int64 `json:"ids" form:"ids"`
		}{}
	)

	uid, username := util.UserInfo(ctx)

	if err = ctx.BindWith(req, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	ctx.JSON(nil, splashSvc.DeleteSelectConfigs(req.IDs, username, uid))
	//nolint:gosimple
	return
}

// splashCategorySave 保存所有自选闪屏分类列表
// Method: POST
// Path: /x/admin/feed/splash/category/save
func splashCategorySave(c *bm.Context) {
	var (
		req = &struct {
			Items []*splashModel.CategoryForSaving `json:"items"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	c.JSON(splashSvc.SaveAllCategories(req.Items, username, uid))
	//nolint:gosimple
	return
}

// splashCategoryListAll 获取自选闪屏分类列表
// Method: GET
// Path: /x/admin/feed/splash/category/all
func splashCategoryListAll(c *bm.Context) {
	var (
		req = &struct {
			Ps    int    `json:"ps" form:"ps" default:"20"`
			Pn    int    `json:"pn" form:"pn" default:"1"`
			State string `json:"state" form:"state"`
		}{}
		state api.SplashScreenConfigState_Enum
	)

	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	if req.State == "" {
		state = -1
	} else {
		if _state, exists := api.SplashScreenConfigState_Enum_value[req.State]; exists {
			state = api.SplashScreenConfigState_Enum(_state)
		} else {
			c.JSON(nil, xecode.Error(xecode.RequestErr, "状态不正确"))
			c.Abort()
			return
		}
	}

	c.JSON(splashSvc.GetAllCategories(true, state))
	//nolint:gosimple
	return
}
