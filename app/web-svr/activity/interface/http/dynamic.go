package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	aecode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/service"

	dynmdl "go-gateway/app/web-svr/activity/interface/model/dynamic"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
)

func actIndex(c *bm.Context) {
	c.JSON(nil, aecode.NativePageOffline)
}

func actDynamic(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func natPages(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func newVideoAid(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func newVideoDyn(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func resourceAid(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func resourceDyn(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func seasonIDs(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func seasonSource(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func resourceRole(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func timelineSource(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func natModule(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func liveDyn(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func actLiked(c *bm.Context) {
	arg := &lmdl.ParamAddLikeAct{}
	if err := c.Bind(arg); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(service.LikeSvc.ActLiked(c, arg, mid))
}

func actList(c *bm.Context) {
	arg := &dynmdl.ParamActList{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.ActList(c, arg, mid))
}

func newActList(c *bm.Context) {
	arg := &dynmdl.ParamNewActList{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.NewActList(c, arg, mid))
}

func videoAct(c *bm.Context) {
	arg := &dynmdl.ParamVideoAct{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.VideoAct(c, arg, mid))
}
