package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/currency"
)

func currencyList(c *bm.Context) {
	v := new(struct {
		Pn int64 `form:"pn" default:"1" validate:"min=1"`
		Ps int64 `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := currSrv.CurrencyList(c, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func currencyItem(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(currSrv.CurrencyItem(c, v.ID))
}

func addCurrency(c *bm.Context) {
	v := new(currency.AddArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, currSrv.AddCurrency(c, v))
}

func saveCurrency(c *bm.Context) {
	v := new(currency.SaveArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, currSrv.SaveCurrency(c, v))
}

func addCurrRelation(c *bm.Context) {
	v := new(currency.RelationArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, currSrv.AddRelation(c, v))
}

func delCurrRelation(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, currSrv.DelRelation(c, v.ID))
}
