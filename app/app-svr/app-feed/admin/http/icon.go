package http

import (
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/icon"
)

func iconList(c *bm.Context) {
	arg := &icon.ListParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(iconSvc.IconList(c, arg))
}

func iconSave(c *bm.Context) {
	arg := &icon.IconSaveParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var (
		name string
		uid  int64
		m    []*icon.Module
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	if err := json.Unmarshal([]byte(arg.Module), &m); err != nil {
		log.Error("iconSave Module Unmarshal module(%s) err(%+v)", arg.Module, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if len(m) == 0 {
		log.Error("iconSave module is empty %s", arg.Module)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	for _, v := range m {
		if v == nil || v.Oid == 0 {
			log.Error("iconSave module is err %s", arg.Module)
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	//stime should < etime
	if arg.Stime >= arg.Etime {
		log.Error("iconSave stime>=etime %d,%d", arg.Stime, arg.Etime)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if arg.EffectGroup == icon.EffectGroupMid && arg.EffectURL == "" {
		//nolint:govet
		log.Error("iconSave EffectURL is empty", arg.EffectURL)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, iconSvc.IconSave(c, arg, name, uid))
}

func iconOpt(c *bm.Context) {
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
	if arg.State != icon.StateDel { //当前只允许删除
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, iconSvc.IconOpt(c, arg.ID, uid, arg.State, name))
}

func iconDetail(c *bm.Context) {
	arg := &struct {
		ID int64 `form:"id" validate:"min=1"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(iconSvc.IconDetail(c, arg.ID))
}

func iconModule(c *bm.Context) {
	arg := &struct {
		Plat int32 `form:"plat"`
	}{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(iconSvc.IconModule(c, arg.Plat))
}
