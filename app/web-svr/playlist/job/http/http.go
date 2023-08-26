package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/playlist/job/conf"
	"go-gateway/app/web-svr/playlist/job/service"
)

var pjSrv *service.Service

// Init .
func Init(c *conf.Config, s *service.Service) {
	pjSrv = s
	engineOut := bm.DefaultServer(c.HTTPServer)
	outerRouter(engineOut)
	// init Outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start error(%v)", err)
		panic(err)
	}
}

func outerRouter(e *bm.Engine) {
	e.Ping(ping)
}

func ping(c *bm.Context) {

}
