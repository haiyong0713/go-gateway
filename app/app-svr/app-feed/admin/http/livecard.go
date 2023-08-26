package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func popLiveCardSave(c *bm.Context) {
	uid, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		return
	}
	var (
		err   error
		param = &show.PopLiveCard{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	param.CreateBy = uname
	c.JSON(nil, popularSvc.PopLiveCardSave(c, param, uname, uid))
}

func popLiveCardList(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			ID       int64  `form:"id"`
			CreateBy string `form:"create_by"`
			State    int    `form:"state"`
			Pn       int    `form:"pn" default:"1"`
			Ps       int    `form:"ps" default:"20"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(popularSvc.PopLiveCardList(ctx, param.ID, param.State, param.CreateBy, param.Pn, param.Ps))
}

func popLiveCardOperate(c *bm.Context) {
	uid, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		return
	}
	var (
		err   error
		param = new(struct {
			ID    int64 `form:"id" validate:"required"`
			State int   `form:"state"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopLiveCardOperate(c, param.ID, param.State, uname, uid))
}
