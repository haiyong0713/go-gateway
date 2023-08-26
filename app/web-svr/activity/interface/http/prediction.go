package http

import (
	bm "go-common/library/net/http/blademaster"
	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
	"go-gateway/app/web-svr/activity/interface/service"
)

func prediction(c *bm.Context) {
	arg := new(premdl.PreParams)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.LikeSvc.PreJudge(c, arg))
}

func preItemUp(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.PreItemUp(c, arg.ID))
}

func preUp(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.PreUp(c, arg.ID))
}

func preSetItem(c *bm.Context) {
	arg := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		Pid   int64 `form:"pid" validate:"min=1"`
		State int   `form:"state"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.PreSetItem(c, arg.ID, arg.Pid, arg.State))
}

func preSet(c *bm.Context) {
	arg := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		Sid   int64 `form:"sid" validate:"min=1"`
		State int   `form:"state"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.PreSet(c, arg.ID, arg.Sid, arg.State))
}
