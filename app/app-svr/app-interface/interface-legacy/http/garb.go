package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
)

func garbDetail(c *bm.Context) {
	param := &space.GarbDetailReq{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(spaceSvr.GarbDetail(c, param))
}

func userGarbList(c *bm.Context) {
	param := &space.GarbListReq{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(spaceSvr.UserGarbList(c, param))
}

func topphotoReset(c *bm.Context) {
	param := new(struct {
		Platform  string `form:"platform" validate:"required"`
		AccessKey string `form:"access_key"`
		Device    string `form:"device"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(nil, spaceSvr.TopphotoReset(c, mid, param.AccessKey, param.Platform, param.Device, 0))
}

func garbDress(c *bm.Context) {
	param := &space.GarbDressReq{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(nil, spaceSvr.GarbDress(c, param))
}

func garbTakeoff(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, spaceSvr.GarbTakeOff(c, mid))
}

func characterList(c *bm.Context) {
	param := &space.CharacterListReq{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(spaceSvr.UserCharacterList(c, param))
}

func characterSet(c *bm.Context) {
	param := &space.CharacterSetReq{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	c.JSON(spaceSvr.CharacterSet(c, param))
}

func characterRemove(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spaceSvr.CharacterRemove(c, mid))
}
