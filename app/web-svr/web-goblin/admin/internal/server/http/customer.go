package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web-goblin/admin/internal/model"
)

func addBusiness(ctx *bm.Context) {
	var (
		err   error
		param = new(model.GbCustomerBusiness)
	)
	if err = ctx.Bind(param); err != nil {
		return
	}
	if param.CustomerType == 3 && param.Logo == "" {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	if err := svc.AddBusiness(ctx, param); err != nil {
		res := map[string]interface{}{}
		res["message"] = "业务创建失败 " + err.Error()
		ctx.JSONMap(res, xecode.RequestErr)
		return
	}
	ctx.JSON(nil, nil)
}

func editBusiness(ctx *bm.Context) {
	var (
		err   error
		param = new(model.GbCustomerBusiness)
	)
	if err = ctx.Bind(param); err != nil {
		return
	}
	if param.CustomerType == 3 && param.Logo == "" {
		ctx.JSON(nil, xecode.RequestErr)
		ctx.Abort()
		return
	}
	if err := svc.EditBusiness(ctx, param); err != nil {
		res := map[string]interface{}{}
		res["message"] = "业务修改失败 " + err.Error()
		ctx.JSONMap(res, xecode.RequestErr)
		return
	}
	ctx.JSON(nil, nil)
}

func delBusiness(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, svc.DelBusiness(c, v.ID))
}

func listBusiness(c *bm.Context) {
	var (
		list []*model.GbCustomerBusiness
		err  error
	)
	v := new(struct {
		CustomerType int64 `form:"customer_type" default:"0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, err = svc.ListBusiness(c, v.CustomerType); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(list, nil)
}

func infoBusiness(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.InfoBusiness(c, v.ID))
}

func addCustomer(ctx *bm.Context) {
	var (
		err   error
		param = new(model.GbCustomerCenters)
	)
	if err = ctx.Bind(param); err != nil {
		return
	}
	if err := svc.AddCustomer(ctx, param); err != nil {
		res := map[string]interface{}{}
		res["message"] = "客服创建失败 " + err.Error()
		ctx.JSONMap(res, xecode.RequestErr)
		return
	}
	ctx.JSON(nil, nil)
}

func editCustomer(ctx *bm.Context) {
	var (
		err   error
		param = new(model.GbCustomerCenters)
	)
	if err = ctx.Bind(param); err != nil {
		return
	}
	if param.ID <= 0 {
		ctx.JSON(nil, xecode.RequestErr)
		return
	}
	if err := svc.EditCustomer(ctx, param); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修改创建失败 " + err.Error()
		ctx.JSONMap(res, xecode.RequestErr)
		return
	}
	ctx.JSON(nil, nil)
}

func delCustomer(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, svc.DelCustomer(c, v.ID))
}

func listCustomer(c *bm.Context) {
	var (
		list []*model.GbCustomerCenters
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn int64 `form:"pn" validate:"min=0" default:"1"`
		Ps int64 `form:"ps" validate:"min=0" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = svc.ListCustomer(c, v.Pn, v.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func infoCustomer(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.InfoCustomer(c, v.ID))
}
