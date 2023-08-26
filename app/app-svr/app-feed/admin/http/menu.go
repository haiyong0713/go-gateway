package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/menu"
)

func menuTabList(c *bm.Context) {
	arg := &menu.ListParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(menuSvr.MenuTabList(c, arg))
}

func menuTabSave(c *bm.Context) {
	arg := &menu.TabSaveParam{}
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
	c.JSON(menuSvr.MenuTabSave(c, arg, name, uid))
}

func menuTabOperate(c *bm.Context) {
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
	c.JSON(nil, menuSvr.MenuTabOperate(c, arg.ID, uid, arg.State, name))

}

func menuSearch(c *bm.Context) {
	arg := &struct {
		TabID int64 `form:"tab_id" validate:"min=1"`
		Type  int   `form:"type"` //0 运营导航模块 1 固定导航
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(menuSvr.MenuSearch(c, arg.TabID, arg.Type))
}
