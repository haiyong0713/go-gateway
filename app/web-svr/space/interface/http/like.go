package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/space/interface/model"
)

func likeVideo(c *bm.Context) {
	req := new(model.LikeVideoReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.LikeVideo(c, req, mid))
}
