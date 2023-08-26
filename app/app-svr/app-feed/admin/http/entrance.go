package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_deletedEntrance = 2
)

func popEntranceSave(c *bm.Context) {
	uid, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = &show.EntranceSave{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopEntranceSave(c, param, uid, uname))
}

func popularEntrance(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			State int `form:"state"`
			Pn    int `form:"pn" default:"1"`
			Ps    int `form:"ps" default:"20"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	if param.State == _deletedEntrance { // 不能查询已删除数据
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(popularSvc.PopularEntrance(ctx, param.State, param.Pn, param.Ps))
}

func popEntranceOperate(c *bm.Context) {
	uid, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID    int64  `form:"id" validate:"required"`
			Title string `form:"title" validate:"required"`
			State int    `form:"state"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopEntranceOperate(c, param.ID, uid, param.State, uname, param.Title))
}

func redDotUpdate(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			Operator string `form:"operator"`
			ModuleID string `form:"module_id" validate:"required"`
			ID       int64  `form:"id"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.RedDotUpdate(c, param.Operator, param.ModuleID, param.ID))
}

func redDotUpdateDisposable(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			Operator string `form:"operator"`
			ModuleID string `form:"module_id"`
			ID       int64  `form:"id"`
			Content  string `form:"red_dot_text"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.RedDotUpdateDisposable(c, param.Operator, param.ModuleID, param.ID, param.Content))
}

func popEntranceView(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID int64 `form:"id" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(popularSvc.PopularView(c, param.ID))
}

func popEntranceViewSave(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID       int64  `form:"id" validate:"required"`
			TopPhoto string `form:"top_photo" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopularViewSave(c, param.ID, param.TopPhoto))
}

func popEntranceViewAdd(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID  int64    `form:"id" validate:"required"`
			RID []string `form:"rid,split" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	var aids []int64
	for _, item := range param.RID {
		var aid int64
		if aid, err = common.GetAvID(item); err != nil {
			return
		}
		aids = append(aids, aid)
	}
	c.JSON(nil, popularSvc.PopularViewAdd(c, param.ID, aids))
}

func popEntranceViewOperate(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID    int64  `form:"id" validate:"required"`
			RID   string `form:"rid" validate:"required"`
			TagID int64  `form:"tag_id"`
			State int    `form:"state" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	var aid int64
	if aid, err = common.GetAvID(param.RID); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopularViewOperate(c, param.ID, aid, param.TagID, param.State))
}

func popEntranceTagAdd(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID     int64   `form:"id" validate:"required"`
			TagIDs []int64 `form:"tag_id,split" validate:"dive,gt=0,required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopularTagAdd(c, param.ID, param.TagIDs))
}

func popEntranceTagDel(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID    int64 `form:"id" validate:"required"`
			TagID int64 `form:"tag_id" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopularTagDel(c, param.ID, param.TagID))
}

func popMiddleSave(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID         int64  `form:"id" validate:"required"`
			LocationId int64  `form:"location_id" validate:"required"`
			TopPhoto   string `form:"top_photo" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, popularSvc.PopularMiddleSave(c, param.ID, param.LocationId, param.TopPhoto))
}

func popMiddleList(c *bm.Context) {
	_, uname := util.UserInfo(c)
	if uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			Pn int `form:"pn" default:"1"`
			Ps int `form:"ps" default:"20"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(popularSvc.PopularMiddleList(c, param.Pn, param.Ps))
}
