package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/menu"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func skinList(c *bm.Context) {
	arg := &menu.SkinListParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(menuSvr.MenuSkinList(c, arg))
}

func skinSave(c *bm.Context) {
	arg := &menu.SkinSaveParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	uid, username := util.UserInfo(c)
	c.JSON(menuSvr.MenuSkinSave(c, arg, username, uid))
}

func skinEdit(c *bm.Context) {
	arg := &struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	uid, username := util.UserInfo(c)
	c.JSON(nil, menuSvr.MenuSkinOperate(c, arg.ID, uid, arg.State, username))
}

func skinSearch(c *bm.Context) {
	arg := &struct {
		SkinID int64 `form:"skin_id" validate:"min=1"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(menuSvr.SkinSearch(c, arg.SkinID))
}
