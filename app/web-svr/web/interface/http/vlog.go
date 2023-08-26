package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"
)

func vlog(c *bm.Context) {
	var (
		err   error
		param = &model.VlogParam{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.MID = midInter.(int64)
		param.LoginEnvent = 2
	} else {
		param.LoginEnvent = 1
	}
	c.JSON(webSvc.Vlog(c, param))
}

func vlogRank(c *bm.Context) {
	var (
		err    error
		arcRes []*model.BvArc
		param  = &model.VlogRankParam{}
	)
	if err = c.Bind(param); err != nil {
		return
	}
	arcRes = webSvc.VlogRank(c, param)
	if arcRes == nil {
		arcRes = make([]*model.BvArc, 0)
	}
	c.JSON(arcRes, nil)
}
