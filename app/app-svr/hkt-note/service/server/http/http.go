package http

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/hkt-note/service/conf"
	"go-gateway/app/app-svr/hkt-note/service/service/article"
	"go-gateway/app/app-svr/hkt-note/service/service/image"
	"go-gateway/app/app-svr/hkt-note/service/service/note"
)

type Server struct {
	NoteSvr *note.Service
	ImgSvr  *image.Service
	ArtSvr  *article.Service
}

// Init int http service
func Init() {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note-service.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	// init internal router
	engineInner := bm.DefaultServer(conf.HTTPServer)
	innerRouter(engineInner)
	// init internal server
	if err := engineInner.Start(); err != nil {
		log.Error("engineInner.Start() error(%v) | config(%v)", err, conf)
		panic(err)
	}
}

// innerRouter init outer router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
}
