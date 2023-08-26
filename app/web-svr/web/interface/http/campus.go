package http

import (
	bm "go-common/library/net/http/blademaster"
	cpmodel "go-gateway/app/web-svr/web/interface/model/campus"

	campusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

func pages(c *bm.Context) {
	var (
		err error
		rs  *cpmodel.PagesReply
	)
	param := &cpmodel.CampusRcmdReq{}
	if err = c.Bind(param); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		param.Mid = midStr.(int64)
	}
	if rs, err = webSvc.Pages(c, param); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func schoolSearch(c *bm.Context) {
	var (
		err error
		rs  *cpmodel.SchoolSearchRep
	)
	v := new(struct {
		Keywords string `form:"keywords"`
		Ps       int    `form:"ps" default:"10" validate:"min=1,max=50"`
		Offset   int    `form:"offset" default:"0" validate:"min=0"`
		FromType string `form:"from_type" default:"PC"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if rs, err = webSvc.SchoolSearch(c, v.Keywords, uint64(v.Ps), uint64(v.Offset), v.FromType); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func schoolRecommend(c *bm.Context) {
	var (
		err error
		rs  []*campusgrpc.CampusInfo
		mid int64
	)
	v := new(struct {
		Lat float32 `form:"lat"`
		Lng float32 `form:"lng"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if rs, err = webSvc.SchoolRecommend(c, uint64(mid), v.Lat, v.Lng); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func OfficialAccounts(c *bm.Context) {
	var (
		err error
		rs  []*cpmodel.OfficialAccountInfo
	)
	v := &cpmodel.CampusOfficialReq{}
	if err = c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		v.Mid = midStr.(int64)
	}
	if rs, err = webSvc.OfficialAccounts(c, v); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func OfficialDynamics(c *bm.Context) {
	var (
		err error
		rs  *cpmodel.OfficialDynamicsReply
	)
	v := &cpmodel.CampusOfficialReq{}
	if err = c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		v.Mid = midStr.(int64)
	}
	if rs, err = webSvc.OfficialDynamics(c, v); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func CampusTopicList(c *bm.Context) {
	var (
		err error
		rs  *campusgrpc.TopicListReply
		mid int64
	)
	v := new(struct {
		CampusId uint64 `form:"campus_id" validate:"required"`
		Offset   int64  `form:"offset"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if rs, err = webSvc.CampusTopicList(c, uint64(mid), v.CampusId, v.Offset); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func CampusBillboard(c *bm.Context) {
	var (
		err error
		rs  *cpmodel.CampusBillBoardReply
		mid int64
	)
	v := new(struct {
		CampusId    int64  `form:"campus_id" validate:"required"`
		VersionCode string `form:"version_code"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if rs, err = webSvc.CampusBillboard(c, mid, v.CampusId, v.VersionCode); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func CampusFeedback(c *bm.Context) {
	req := &cpmodel.CampusFeedbackReq{}
	var (
		err error
		rs  *cpmodel.CampusFeedbackReply
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		req.Mid = midStr.(int64)
	}
	if rs, err = webSvc.CampusFeedback(c, req); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func CampusNearbyRcmd(c *bm.Context) {
	req := &cpmodel.CampusNearbyRcmdReq{}
	var (
		err error
		rs  *cpmodel.CampusNearbyRcmdReply
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		req.Mid = midStr.(int64)
	}
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		req.Buvid = ck.Value
	}
	if rs, err = webSvc.CampusNearbyRcmd(c, req); err != nil {
		// rs.Items = make([]*rcmd.Item, 0)
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func CampusRedDot(c *bm.Context) {
	req := &cpmodel.CampusRedDotReq{}
	var (
		err error
		rs  *cpmodel.CampusRedDotReply
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		req.Mid = midStr.(int64)
	}
	if rs, err = webSvc.CampusRedDot(c, req); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}
