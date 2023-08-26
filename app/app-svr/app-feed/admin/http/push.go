package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/push"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func pushList(c *bm.Context) {
	c.JSON(pushSvc.PushList(c))
}

func pushSave(c *bm.Context) {
	params := &push.PushDetail{}
	if err := c.Bind(params); err != nil {
		return
	}
	uid, username := util.UserInfo(c)
	if params.STime > params.ETime {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, pushSvc.PushSave(c, params, username, uid))
}

func pushDetail(c *bm.Context) {
	params := struct {
		ID int64 `form:"id"`
	}{}
	if err := c.Bind(&params); err != nil {
		return
	}
	c.JSON(pushSvc.PushDetail(c, params.ID))
}

func pushDelete(c *bm.Context) {
	params := struct {
		ID int64 `form:"id"`
	}{}
	if err := c.Bind(&params); err != nil {
		return
	}
	uid, username := util.UserInfo(c)
	c.JSON(nil, pushSvc.PushDelete(c, params.ID, uid, username))
}
