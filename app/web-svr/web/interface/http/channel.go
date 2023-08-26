package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model/channel"
)

func channelDetail(c *bm.Context) {
	var params = &channel.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	c.JSON(webSvc.ChannelDetail(c, params))
}

func channelMultiple(c *bm.Context) {
	var params = &channel.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	c.JSON(webSvc.ChannelMultiple(c, params))
}

func channelSelected(c *bm.Context) {
	var params = &channel.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	c.JSON(webSvc.ChannelSelected(c, params))
}

func channelRed(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(chSvr.Red(c, mid))
}

func categoryList(c *bm.Context) {
	c.JSON(chSvr.CategoryList(c))
}

func channelArcList(c *bm.Context) {
	var (
		mid int64
		req = &channel.ChannelArcListReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.ChannelArcList(c, mid, req))
}

func channelList(c *bm.Context) {
	var (
		mid int64
		req = &channel.ChannelListReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.ChannelList(c, mid, req))
}

func subscribedList(c *bm.Context) {
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(chSvr.SubscribedList(c, mid))
}

func viewList(c *bm.Context) {
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(chSvr.ViewList(c, mid))
}

func stick(c *bm.Context) {
	var req = &channel.StickReq{}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, chSvr.Stick(c, mid, req))
}

func subscribe(c *bm.Context) {
	var req = &channel.SubscribeReq{}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, chSvr.Subscribe(c, mid, req))
}

func unsubscribe(c *bm.Context) {
	var req = &channel.UnsubscribeReq{}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, chSvr.Unsubscribe(c, mid, req))
}

func hotList(c *bm.Context) {
	var (
		mid int64
		req = &channel.HotListReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.HotList(c, mid, req))
}

func webDetail(c *bm.Context) {
	var (
		mid int64
		req = &channel.WebDetailReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.Detail(c, mid, req))
}

func featuredList(c *bm.Context) {
	var (
		mid int64
		req = &channel.FeaturedListReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.FeaturedList(c, mid, req))
}

func multipleList(c *bm.Context) {
	var (
		mid int64
		req = &channel.MultipleListReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.MultipleList(c, mid, req))
}

func searchChannel(c *bm.Context) {
	var (
		mid int64
		req = &channel.SearchReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.Search(c, mid, req))
}

func topList(c *bm.Context) {
	var (
		mid int64
		req = &channel.TopListReq{}
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(chSvr.TopList(c, mid, req))
}
