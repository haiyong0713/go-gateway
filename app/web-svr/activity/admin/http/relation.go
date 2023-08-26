package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

func ListActRelation(c *bm.Context) {
	arg := new(model.ActRelationListArgs)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.ActRelationList(c, arg))
}

func GetActRelation(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"required,min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.ActRelationGet(c, arg.ID))
}

func AddActRelation(c *bm.Context) {
	arg := new(model.ActRelationSubject)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.ActRelationAdd(c, arg))
}

func UpdateActRelation(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"required,min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	data := c.Request.Form
	query := make(map[string]interface{})
	for k, v := range data {
		query[k] = v
	}
	c.JSON(actSrv.ActRelationUpdate(c, arg.ID, query))
}

func StateActRelation(c *bm.Context) {
	arg := new(struct {
		ID    int64 `form:"id" validate:"required,min=1"`
		State int64 `form:"state" validate:"gt=-2,lt=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.ActRelationState(c, arg.ID, arg.State))
}
