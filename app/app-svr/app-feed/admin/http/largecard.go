package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func popLargeCardSave(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		return
	}
	var (
		err   error
		param = &show.PopLargeCard{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	param.CreateBy = uname
	params := c.Request.Form
	rid := params.Get("rid")
	param.RID, _ = common.GetAvID(rid)
	c.JSON(nil, popularSvc.PopLargeCardSave(c, param))
}

func popLargeCardList(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			ID       int64  `form:"id"`
			CreateBy string `form:"create_by"`
			Rid      string `form:"rid"`
			Pn       int    `form:"pn" default:"1"`
			Ps       int    `form:"ps" default:"20"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	var rid int64
	if param.Rid != "" {
		if rid, err = common.GetAvID(param.Rid); err != nil {
			c.JSONMap(map[string]interface{}{"message": "avid/bvid 非法"}, ecode.RequestErr)
			return
		}
	}
	c.JSON(popularSvc.PopLargeCardList(ctx, param.ID, param.CreateBy, rid, param.Pn, param.Ps))
}

func popLargeCardOperate(c *bm.Context) {
	_, uname := util.UserInfo(c)
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
	c.JSON(nil, popularSvc.PopLargeCardOperate(c, param.ID, param.State))
}
