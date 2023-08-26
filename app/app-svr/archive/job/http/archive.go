package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/archive/job/model/databus"
)

func arcUpdate(c *bm.Context) {
	req := &databus.Videoup{}
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if req.Aid <= 0 || req.Route == "" {
		log.Error("req is error(%+v)", req)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	log.Info("http arcUpdate value(%+v)", req)
	arcJobSrv.Prom.Incr("from http arcUpdate")
	arcJobSrv.VideoUpService(req)
	c.JSON(nil, nil)
}
