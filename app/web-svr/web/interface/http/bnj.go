package http

import bm "go-common/library/net/http/blademaster"

func bnj2019(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.Bnj2019(c, mid))
}

func bnj2019Aids(c *bm.Context) {
	data := make(map[string]interface{}, 1)
	data["list"] = webSvc.Bnj2019Aids(c)
	c.JSON(data, nil)
}

func timeline(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	list, err := webSvc.Timeline(c, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"list": list}, nil)
}

func bnj2020(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.Bnj2020(c, mid))
}

func bnj2020Item(c *bm.Context) {
	var (
		mid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.Bnj2020Item(c, v.Aid, mid))
}

func bnj2020ElecShow(c *bm.Context) {
	c.JSON(webSvc.Bnj20ElecShow(c), nil)
}

func bnj2020Aids(c *bm.Context) {
	data := make(map[string]interface{}, 1)
	data["list"] = webSvc.Bnj2020Aids(c)
	c.JSON(data, nil)
}

func bnj2020Timeline(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	list, err := webSvc.Bnj20Timeline(c, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"list": list}, nil)
}
