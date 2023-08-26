package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const (
	_headerBuvid = "Buvid"
)

func nodeinfo(c *bm.Context) {
	params := &model.NodeInfoParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	var (
		mid      int64
		newBuvid string
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params.HandlerCursor() // 处理Cursor
	if err := params.HandlerBvid(); err != nil {
		c.JSON(nil, err)
		return
	}
	newBuvid = buildBuvid(c, params, mid)
	result, err := svc.RouterInfo(c, params, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if newBuvid != "" { // 生成的buvid下发
		result.Buvid = newBuvid
	}
	c.JSON(result, nil)
}

func nodeinfoPreview(c *bm.Context) {
	params := &model.NodeinfoPreReq{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params.HandlerCursor() // 处理Cursor
	if err := params.HandlerBvid(); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(svc.RouterInfoPreview(c, params, mid))
}

func edgeV2infoPreview(c *bm.Context) {
	params := &model.EdgeInfoV2PreReq{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params.HandlerCursor() // 处理Cursor
	c.JSON(svc.RouterInfoV2Preview(c, params, mid))
}

func mark(c *bm.Context) {
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Mark int64  `form:"mark" validate:"min=1,max=5"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if v.Aid == 0 {
		if v.Bvid == "" { // 如果aid和bvid都没有，直接返回
			c.JSON(nil, ecode.AidBvidNil)
			return
		}
		if v.Aid, _ = model.GetAvID(v.Bvid); v.Aid == 0 {
			c.JSON(nil, ecode.BvidIllegal)
			return
		}
	}
	c.JSON(nil, svc.AddMark(c, v.Aid, mid, v.Mark))
}

func edgeinfoV2(c *bm.Context) {
	params := new(model.EdgeInfoV2Param)
	if err := c.Bind(params); err != nil {
		return
	}
	var (
		mid      int64
		newBuvid string
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	newBuvid = buildBuvid(c, &params.NodeInfoParam, mid)
	params.HandlerCursor()
	if err := params.HandlerBvid(); err != nil {
		c.JSON(nil, err)
		return
	}
	result, err := svc.RouterInfoV2(c, params, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if newBuvid != "" {
		result.Buvid = newBuvid // 生成的buvid下发
	}
	c.JSON(result, nil)
}

func buildBuvid(c *bm.Context, params *model.NodeInfoParam, mid int64) (newBuvid string) {
	if buvid := c.Request.Header.Get(_headerBuvid); buvid != "" { // 优先取header中buvid，其次取参数中的
		params.Buvid = buvid
	}
	if params.Portal == 0 && params.EdgeID == 0 && params.MobiApp == "" && params.Buvid == "" { // web端不带node id请求下发buvid
		if ck, ckErr := c.Request.Cookie("buvid3"); ckErr == nil { // 默认从cookie获取
			newBuvid = ck.Value
			svc.IncrPromBusiness("cookie_buvid")
		} else {
			newBuvid = svc.GenBuvid(mid, params.AID) // cookie获取失败继续下发
			svc.IncrPromBusiness("cookie_no_buvid")
		}
		params.Buvid = newBuvid // 补上buvid
	}
	return

}
