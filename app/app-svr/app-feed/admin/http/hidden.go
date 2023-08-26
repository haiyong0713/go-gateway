package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/hidden"
)

func hiddenList(c *bm.Context) {
	arg := &hidden.ListParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(hiddenSvc.HiddenList(c, arg))
}

func correctSaveParam(arg *hidden.HiddenSaveParam) bool {
	//频道入口和分区入口不同时为空
	if arg.SID == 0 && arg.RID == 0 && arg.CID == 0 && arg.ModuleID == 0 && arg.HideDynamic == 0 {
		return false
	}
	//stime should < etime
	if arg.Stime >= arg.Etime {
		return false
	}
	if arg.HiddenCondition != "include" && arg.HiddenCondition != "exclude" {
		return false
	}
	return true
}

func hiddenSave(c *bm.Context) {
	arg := &hidden.HiddenSaveParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	if !correctSaveParam(arg) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var (
		name string
		uid  int64
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	c.JSON(nil, hiddenSvc.HiddenSave(c, arg, name, uid))
}

func hiddenOpt(c *bm.Context) {
	arg := &struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var (
		name string
		uid  int64
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	c.JSON(nil, hiddenSvc.HiddenOpt(c, arg.ID, uid, arg.State, name))

}

func hiddenDetail(c *bm.Context) {
	arg := &struct {
		ID int64 `form:"id" validate:"min=1"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(hiddenSvc.HiddenDetail(c, arg.ID))
}

func entranceSearch(c *bm.Context) {
	arg := &struct {
		OID  int64 `form:"oid" validate:"min=1"`
		Type int   `form:"type"` //0 首页入口 1 频道入口 2 【我的】页入口 3 一级模块
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(hiddenSvc.EntranceSearch(c, arg.OID, arg.Type))
}
