package http

import (
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/space/admin/model"
	"go-gateway/app/web-svr/space/admin/util"
)

const (
	Online   = 1
	Offline  = 0
	LastTime = 2147454847
)

func userTabAdd(c *bm.Context) {
	var (
		err error
	)
	var req = &model.UserTabReq{}
	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	// username
	username, uid := util.UserInfo(c)
	if username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	// 验证起始时间
	if req.Etime > 0 {
		if req.Etime.Time().Unix() < time.Now().Unix() {
			c.JSONMap(map[string]interface{}{"message": "结束时间晚于当前时间"}, ecode.RequestErr)
			c.Abort()
			return
		}
	} else {
		req.Etime = LastTime // 结束时间未配置, 则用不过期
	}
	if req.Etime <= req.Stime {
		c.JSONMap(map[string]interface{}{"message": "结束时间小于开始时间"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.Stime.Time().Unix() <= time.Now().Unix() {
		req.Online = 1
	} else {
		req.Online = 0
	}

	if req.TabCont != 0 && req.H5Link != "" {
		c.JSONMap(map[string]interface{}{"message": "活动NA页面和H5链接不可同时配置"}, ecode.RequestErr)
		c.Abort()
		return
	}

	if len(req.Limits) == 0 {
		c.JSONMap(map[string]interface{}{"message": "未设定任何下发版本限制信息"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if _, err = spcSvc.MidInfo(c, req.Mid); err != nil {
		c.JSONMap(map[string]interface{}{"message": "无效mid"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = spcSvc.AddSpaceUserTab(req); err != nil {
		res := map[string]interface{}{}
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
	// action log
	if err = util.AddLogs(model.LogExamine, username, uid, req.ID, "Add", req); err != nil {
		log.Error("userTabAdd AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func commercialToUser(req *model.CommercialTabReq) (ret *model.UserTabReq) {
	ret = &model.UserTabReq{}
	ret.ID = req.ID
	ret.Online = req.Online
	ret.Etime = req.Etime
	ret.Stime = req.Stime
	ret.Deleted = req.Deleted
	ret.Mid = req.Mid
	ret.TabType = req.TabType
	ret.TabOrder = req.TabOrder
	ret.TabCont = req.TabCont
	ret.TabName = req.TabName
	ret.IsSync = 1
	ret.IsDefault = req.IsDefault
	return
}

func checkParam(param *model.CommercialTabReq) (err error) {
	if param.TabType < 1 {
		err = ecode.Error(-400, "Tab类型错误")
		return
	}
	if param.TabCont < 1 {
		err = ecode.Error(-400, "Tab内容错误")
		return
	}
	if param.Stime < 1 {
		err = ecode.Error(-400, "开始时间不合法")
		return
	}
	if param.Etime <= param.Stime {
		err = ecode.Error(-400, "结束时间小于开始时间")
	} else if param.Etime.Time().Unix() < time.Now().Unix() {
		err = ecode.Error(-400, "结束时间晚于当前时间")
	}
	return
}

func commercialTabAdd(c *bm.Context) {
	var (
		err error
	)
	var req = &model.CommercialTabReq{}
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	// username
	//username, uid := util.UserInfo(c)
	if req.Username == "" {
		c.JSONMap(map[string]interface{}{"message": "required username"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.Etime == 0 {
		req.Etime = LastTime // 结束时间未配置, 则用不过期
	}
	// 校验必填参数
	if err = checkParam(req); err != nil {
		c.Abort()
		return
	}
	if req.Stime.Time().Unix() <= time.Now().Unix() {
		req.Online = 1
	} else {
		req.Online = 0
	}
	if _, err = spcSvc.MidInfo(c, req.Mid); err != nil {
		c.JSONMap(map[string]interface{}{"message": "无效mid"}, ecode.RequestErr)
		c.Abort()
		err = ecode.RequestErr
		return
	}
	ret := commercialToUser(req)
	if err = spcSvc.AddSpaceUserTab(ret); err != nil {
		res := map[string]interface{}{}
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res := &struct {
		ID int64 `json:"id" form:"id"`
	}{}
	res.ID = ret.ID
	c.JSON(res, nil)
	// action log
	if err = util.AddLogs(model.LogExamine, req.Username, 0, ret.ID, "Add", ret); err != nil {
		log.Error("userTabAdd AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func userTabModify(c *bm.Context) {
	var (
		err error
	)
	var req = &model.UserTabReq{}
	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, uid := util.UserInfo(c)
	if username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		c.JSONMap(map[string]interface{}{"message": "id无效"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.TabCont != 0 && req.H5Link != "" {
		c.JSONMap(map[string]interface{}{"message": "活动NA页面和H5链接不可同时配置"}, ecode.RequestErr)
		c.Abort()
		return
	}
	// 验证起始时间
	if req.Etime > 0 {
		if req.Etime.Time().Unix() < time.Now().Unix() {
			c.JSONMap(map[string]interface{}{"message": "结束时间晚于当前时间"}, ecode.RequestErr)
			c.Abort()
			err = ecode.RequestErr
			return
		}
	} else {
		req.Etime = LastTime // 结束时间未配置, 则用不过期
	}
	if req.Etime <= req.Stime {
		c.JSONMap(map[string]interface{}{"message": "结束时间小于开始时间"}, ecode.RequestErr)
		c.Abort()
		err = ecode.RequestErr
		return
	}
	// 新增配置状态, 早于当前时间则立即生效
	if req.Stime.Time().Unix() <= time.Now().Unix() {
		req.Online = 1
	} else {
		req.Online = 0
	}

	if len(req.Limits) == 0 {
		c.JSONMap(map[string]interface{}{"message": "未设定任何下发版本限制信息"}, ecode.RequestErr)
		c.Abort()
		return
	}

	if _, err = spcSvc.MidInfo(c, req.Mid); err != nil {
		c.JSONMap(map[string]interface{}{"message": "无效mid"}, ecode.RequestErr)
		c.Abort()
		err = ecode.RequestErr
		return
	}
	if err = spcSvc.ModifySpaceUserTab(req); err != nil {
		res := map[string]interface{}{}
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
	// action log
	if err = util.AddLogs(model.LogExamine, username, uid, req.ID, "Modify", req); err != nil {
		log.Error("modifyUserTab AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func commercialTabModify(c *bm.Context) {
	var (
		err error
	)
	var req = &model.CommercialTabReq{}
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	if req.Username == "" {
		c.JSONMap(map[string]interface{}{"message": "required username"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		c.JSONMap(map[string]interface{}{"message": "id无效"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.Etime == 0 {
		req.Etime = LastTime // 结束时间未配置, 则用不过期
	}
	// 校验必填参数
	if err = checkParam(req); err != nil {
		c.Abort()
		return
	}
	// 新增配置状态, 早于当前时间则立即生效
	if req.Stime.Time().Unix() <= time.Now().Unix() {
		req.Online = 1
	} else {
		req.Online = 0
	}
	if _, err = spcSvc.MidInfo(c, req.Mid); err != nil {
		c.JSONMap(map[string]interface{}{"message": "无效mid"}, ecode.RequestErr)
		c.Abort()
		err = ecode.RequestErr
		return
	}
	ret := commercialToUser(req)
	if err = spcSvc.ModifySpaceUserTab(ret); err != nil {
		res := map[string]interface{}{}
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
	// action log
	if err = util.AddLogs(model.LogExamine, req.Username, 0, ret.ID, "Modify", ret); err != nil {
		log.Error("modifyUserTab AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func userTabOnline(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID     int64 `json:"id" form:"id"`
		Online int   `json:"online" form:"online"`
	}{}
	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		res["message"] = "传参错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	username, uid := util.UserInfo(c)
	if username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		res["message"] = "id无效"
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = spcSvc.OnlineUserTab(req.ID, req.Online); err != nil {
		res["message"] = "上下线错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	c.JSON(nil, nil)
	// action log
	var action string
	if req.Online == Online {
		action = "online"
	} else {
		action = "offline"
	}
	if err = util.AddLogs(model.LogExamine, username, uid, req.ID, action, req); err != nil {
		log.Error("onlineUserTab AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func commercialTabOnline(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID       int64  `json:"id"`
		Online   int    `json:"online"`
		Username string `json:"username"`
	}{}
	if err = c.BindWith(req, binding.JSON); err != nil {
		res["message"] = "传参错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.Username == "" {
		c.JSONMap(map[string]interface{}{"message": "required username"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		res["message"] = "id无效"
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = spcSvc.OnlineUserTab(req.ID, req.Online); err != nil {
		res["message"] = "上下线错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	c.JSON(nil, nil)
	// action log
	var action string
	if req.Online == Online {
		action = "online"
	} else {
		action = "offline"
	}
	if err = util.AddLogs(model.LogExamine, req.Username, 0, req.ID, action, req); err != nil {
		log.Error("onlineUserTab AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func userTabList(c *bm.Context) {
	var (
		err  error
		list *model.UserTabList
	)
	res := map[string]interface{}{}
	var req = &model.UserTabListReq{}
	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		c.JSONMap(map[string]interface{}{"message": "传参错误"}, ecode.RequestErr)
		c.Abort()
		return
	}
	// 校验page大小
	if req.Pn < 1 {
		res["message"] = "页数最少为1"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	//nolint:gomnd
	if req.Ps > 200 {
		res["message"] = "页面数量太多"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if list, err = spcSvc.UserTabList(req); err != nil {
		res["message"] = "列表获取失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(list, nil)
}

func userTabDelete(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID   int64 `json:"id"`
		Type int   `json:"type" default:"1"`
	}{}
	if err = c.BindWith(req, binding.JSON); err != nil {
		res["message"] = "传参错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	username, uid := util.UserInfo(c)
	if username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		res["message"] = "id无效"
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = spcSvc.DeleteSpaceUserTab(req.ID, req.Type); err != nil {
		res["message"] = "删除错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	c.JSON(nil, nil)
	// action log
	if err = util.AddLogs(model.LogExamine, username, uid, req.ID, "delete", req); err != nil {
		log.Error("deleteUserTab AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func commercialTabDelete(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID       int64  `json:"id" form:"id"`
		Username string `json:"username" form:"username"`
		Type     int    `json:"type" form:"type" default:"1"`
	}{}
	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		res["message"] = "传参错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.Username == "" {
		c.JSONMap(map[string]interface{}{"message": "required username"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		res["message"] = "id无效"
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = spcSvc.DeleteSpaceUserTab(req.ID, req.Type); err != nil {
		res["message"] = "删除错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	c.JSON(nil, nil)
	// action log
	if err = util.AddLogs(model.LogExamine, req.Username, 0, req.ID, "delete", req); err != nil {
		log.Error("deleteUserTab AddLog arg(%v) error(%v)", req, err)
		return
	}
}

func userMidInfo(c *bm.Context) {
	var (
		err  error
		info *model.MidInfoReply
	)
	res := map[string]interface{}{}
	req := &struct {
		Mid int64 `json:"mid" form:"mid"`
	}{}
	if err = c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if info, err = spcSvc.MidInfo(c, req.Mid); err != nil {
		res["message"] = "mid校验失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(info, nil)
}
