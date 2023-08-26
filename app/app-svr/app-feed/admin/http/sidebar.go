package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

func moduleList(c *bm.Context) {
	var moduleParam = struct {
		Plat int32 `form:"plat"`
	}{}
	if err := c.Bind(&moduleParam); err != nil {
		return
	}
	c.JSON(sidebarSvc.ModuleList(c, moduleParam.Plat))
}

func moduleSave(c *bm.Context) {
	arg := &show.SaveModuleParam{}
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
	if arg.MType != 1 {
		//当前仅开放添加:
		//我的页模块
		//首页模块的发布弹窗和分类入口
		//首页模块的港澳台垂类tab
		if !(arg.MType == 2 && (arg.Style == show.ModuleStyle_Launcher || arg.Style == show.ModuleStyle_Classification || arg.Style == show.ModuleStyle_Recommend_tab || arg.Style == show.ModuleStyle_Publish_bubble)) {
			c.JSON(nil, ecode.RequestErr)
			c.Abort()
			return
		}
	}
	if arg.MType == 1 && arg.Style == 3 { //运营位模块只接收 指定的运营样式和下发条件
		if (arg.OpStyleType == 0 && arg.OpLoadCondition != "10,11") || (arg.OpStyleType == 1 && arg.OpLoadCondition != "00,01") {
			c.JSON(nil, ecode.RequestErr)
			c.Abort()
			return
		}
	}
	c.JSON(nil, sidebarSvc.ModuleSave(c, arg, name, uid))
}

func moduleDetail(c *bm.Context) {
	arg := &struct {
		ID int64 `form:"id" validate:"min=1"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(sidebarSvc.ModuleDetail(c, arg.ID))
}

func moduleOpt(c *bm.Context) {
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
	c.JSON(nil, sidebarSvc.ModuleOpt(c, arg.ID, uid, arg.State, name))
}

// 二级模块列表
func moduleItemList(c *bm.Context) {
	req := new(show.ModuleItemListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(sidebarSvc.ModuleItemList(c, req))
}

// 二级模块保存
func moduleItemSave(c *bm.Context) {
	req := new(show.SidebarEntity)
	if err := c.Bind(req); err != nil {
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
	c.JSON(nil, sidebarSvc.ModuleItemSave(c, req, name, uid))
}

// 二级模块详情
func moduleItemDetail(c *bm.Context) {
	req := &struct {
		ID int64 `form:"id" validate:"min=1"`
	}{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(sidebarSvc.ModuleItemDetail(c, req.ID))
}

// 二级模块状态变更
func moduleItemOpt(c *bm.Context) {
	req := &struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int32 `form:"state"`
	}{}
	if err := c.Bind(req); err != nil {
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
	c.JSON(nil, sidebarSvc.ModuleItemOpt(c, req.ID, req.State, name, uid))
}
