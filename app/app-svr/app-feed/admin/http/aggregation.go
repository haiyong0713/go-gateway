package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/aggregation"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func aggregationList(c *bm.Context) {
	var (
		err   error
		param = &aggregation.AggListReq{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(aggSvc.AggregationList(c, param))
}

func aggregationSave(c *bm.Context) {
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = &aggregation.AggSaveReq{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	if param.ID == 0 {
		c.JSON(nil, aggSvc.AddAggregation(c, param.AggPub, param.TagID, name, uid))
	} else {
		c.JSON(nil, aggSvc.UpdateAggregation(c, param.AggPub, param.TagID, name, uid))
	}
}

func aggOperate(c *bm.Context) {
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	var (
		err   error
		param = new(struct {
			ID       int64  `form:"id" validate:"required"`
			State    int    `form:"state"`
			HotTitle string `form:"hot_title"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, aggSvc.AggOperate(c, param.ID, uid, param.State, name, param.HotTitle))
}

func aggView(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			ID int64 `form:"id" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(aggSvc.AggView(ctx, param.ID))
}

func aggViewAdd(c *bm.Context) {
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
	//nolint:gomnd
	if len(param.RID) > 20 {
		c.JSON(nil, ecode.LimitExceed)
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
	c.JSON(nil, aggSvc.AggViewAdd(c, param.ID, aids))
}

func aggViewOp(c *bm.Context) {
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
	c.JSON(nil, aggSvc.AggViewOp(c, param.ID, aid, param.TagID, param.State))
}

func tag(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			TagName string `form:"tag_name" validate:"required"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(aggSvc.Tag(ctx, param.TagName))
}

func aggTagAdd(c *bm.Context) {
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
	c.JSON(nil, aggSvc.AggTagAdd(c, param.ID, param.TagIDs))
}

func aggTagDel(c *bm.Context) {
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
	c.JSON(nil, aggSvc.AggTagDel(c, param.ID, param.TagID))
}
