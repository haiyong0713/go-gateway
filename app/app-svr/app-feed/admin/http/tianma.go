package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"

	model "go-gateway/app/app-svr/app-feed/admin/model/tianma"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/app/app-svr/app-feed/ecode"
)

func IsHdfsPathAccessible(c *bm.Context) {
	hdfsPath := c.Request.Form.Get("hdfs_path")
	flag, _ := tianmaSvc.IsHdfsPathAccessible(hdfsPath)
	resp := struct {
		IsAccessible bool `json:"is_accessible"`
	}{
		IsAccessible: flag,
	}
	c.JSON(resp, nil)
	//nolint:gosimple
	return
}

// 判断http链接是否可访问
func IsHttpPathAccessible(c *bm.Context) {
	httpPath := c.Request.Form.Get("http_path")
	flag, err := tianmaSvc.IsHttpPathAccessible(httpPath)
	if err != nil {
		c.JSONMap(nil, err)
		c.Abort()
		return
	}
	resp := struct {
		IsAccessible bool `json:"is_accessible"`
	}{
		IsAccessible: flag,
	}

	c.JSON(resp, nil)
}

// 获取用于上传的预签名 url
func bossSignedUploadUrl(c *bm.Context) {
	var (
		err error
	)

	_, username := util.UserInfo(c)

	url, err := tianmaSvc.BossSignedUploadUrl(username)

	c.JSON(url, err)
	//nolint:gosimple
	return
}

// 获取用于下载的预签名 url，有有效期限制
type bossSignedDownloadUrlReq struct {
	Key string `form:"key" json:"key" validate:"required"`
}

func bossSignedDownloadUrl(c *bm.Context) {
	var (
		err     error
		request = &bossSignedDownloadUrlReq{}
	)

	if err = c.Bind(request); err != nil {
		return
	}

	_, username := util.UserInfo(c)

	url, err := tianmaSvc.BossSignedDownloadUrl(request.Key, username)

	if err != nil {
		log.Error("tianmaSvc.BossSignedDownloadUrl error(%v)", err)
	}

	c.JSON(url, err)
	//nolint:gosimple
	return
}

// 更改某一条推荐的文件信息
type updateMidFileInfoReq struct {
	Id         int64  `json:"id" form:"id" validate:"required"`
	FileStatus int    `json:"file_status" form:"file_status"`
	FilePath   string `json:"file_path" form:"file_path"`
}

func updateMidFileInfo(c *bm.Context) {
	var (
		err     error
		request = &updateMidFileInfoReq{}
	)

	if err = c.Bind(request); err != nil {
		return
	}

	_, username := util.UserInfo(c)

	err = tianmaSvc.UpdateMidFileInfo(request.Id, &model.PosRecItem{
		FileStatus: request.FileStatus,
		FilePath:   request.FilePath,
	}, username)

	if err != nil {
		log.Error("tianmaSvc.UpdateMidFileInfo error(%v)", err)
	}

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 新建业务弹窗配置
func popupConfigAdd(c *bm.Context) {
	var (
		err error
		req = &struct {
			ImageURL    string     `json:"img_url" form:"img_url" validate:"required"`
			Description string     `json:"description" form:"description" validate:"required,min=5,max=30"`
			ReType      int        `json:"redirect_type" form:"redirect_type" validate:"required"`
			ReTarget    string     `json:"redirect_target" form:"redirect_target"`
			Builds      string     `json:"builds" form:"builds"`
			CrowdType   int        `json:"crowd_type" form:"crowd_type" validate:"required"`
			CrowdBase   int        `json:"crowd_base" form:"crowd_base"`
			CrowdValue  string     `json:"crowd_value" form:"crowd_value"`
			STime       xtime.Time `json:"stime" form:"stime" validate:"required,gt=0"`
			ETime       xtime.Time `json:"etime" form:"etime" validate:"required,gt=0"`
		}{}
		res struct {
			ID int64 `json:"id"`
		}
	)

	uid, username := util.UserInfo(c)

	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	if req.ReType != model.PopupReTypeNone && req.ReTarget == "" {
		err = ecode.PopupConfigParameterError
	}
	if req.CrowdType != model.PopupCrowdTypeNone && req.CrowdValue == "" {
		err = ecode.PopupConfigParameterError
	}
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}

	res.ID, err = tianmaSvc.AddPopupConfig(&model.PopupConfig{
		ImageURL:    req.ImageURL,
		Description: req.Description,
		ReType:      req.ReType,
		ReTarget:    req.ReTarget,
		Builds:      req.Builds,
		CrowdType:   req.CrowdType,
		CrowdBase:   req.CrowdBase,
		CrowdValue:  req.CrowdValue,
		STime:       req.STime,
		ETime:       req.ETime,
	}, username, uid)

	c.JSON(res, err)
	//nolint:gosimple
	return
}

// 编辑配置
func popupConfigEdit(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID          int64      `json:"id" form:"id" validate:"required,gt=0"`
			ImageURL    string     `json:"img_url" form:"img_url"`
			Description string     `json:"description" form:"description"`
			ReType      int        `json:"redirect_type" form:"redirect_type"`
			ReTarget    string     `json:"redirect_target" form:"redirect_target"`
			Builds      string     `json:"builds" form:"builds"`
			CrowdType   int        `json:"crowd_type" form:"crowd_type"`
			CrowdBase   int        `json:"crowd_base" form:"crowd_base"`
			CrowdValue  string     `json:"crowd_value" form:"crowd_value"`
			STime       xtime.Time `json:"stime" form:"stime"`
			ETime       xtime.Time `json:"etime" form:"etime"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	err = tianmaSvc.UpdatePopupConfig(&model.PopupConfig{
		ID:          req.ID,
		ImageURL:    req.ImageURL,
		Description: req.Description,
		ReType:      req.ReType,
		ReTarget:    req.ReTarget,
		Builds:      req.Builds,
		CrowdType:   req.CrowdType,
		CrowdBase:   req.CrowdBase,
		CrowdValue:  req.CrowdValue,
		STime:       req.STime,
		ETime:       req.ETime,
	}, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 获取配置列表
func popupConfigList(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID     int64  `json:"id" form:"id"`
			Status int    `json:"status" form:"status"`
			Ps     int    `json:"ps" form:"ps" default:"10"`
			Pn     int    `json:"pn" form:"pn" default:"1"`
			Order  string `json:"order" form:"order" default:"mtime DESC"`
		}{}
		listRes []*model.PopupConfig
		total   int64
	)

	if err = c.Bind(req); err != nil {
		return
	}

	listRes, total, err = tianmaSvc.GetPopupConfigList(&model.PopupConfig{
		ID:     req.ID,
		Status: req.Status,
	}, req.Ps, req.Pn, req.Order)

	res := &model.PopupConfigListWithPager{
		Pager: &model.Pager{
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

// 修改审核状态，包括默认配置和自选配置
func popupConfigDelete(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID         int64 `json:"id" form:"id" validate:"required"`
			AuditState int   `json:"audit_state" form:"audit_state"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	err = tianmaSvc.DeletePopupConfig(req.ID, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}

// 修改审核状态，包括默认配置和自选配置
func popupConfigAudit(c *bm.Context) {
	var (
		err error
		req = &struct {
			ID         int64 `json:"id" form:"id" validate:"required"`
			AuditState int   `json:"audit_state" form:"audit_state" validate:"required"`
		}{}
	)

	uid, username := util.UserInfo(c)

	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	err = tianmaSvc.AuditPopupConfig(req.ID, req.AuditState, username, uid)

	c.JSON(nil, err)
	//nolint:gosimple
	return
}
