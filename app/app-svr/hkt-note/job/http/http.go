package http

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/hkt-note/job/conf"
	"go-gateway/app/app-svr/hkt-note/job/service/article"
	"go-gateway/app/app-svr/hkt-note/job/service/note"
)

var (
	NoteSrv *note.Service
	ArtSrv  *article.Service
)

// Init init http router.
func Init() {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note-job.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	initService(conf)
	e := bm.DefaultServer(conf.HTTPServer)
	innerRouter(e)
	// init internal server
	e.Any("railgun/note_add", NoteSrv.NoteAddRailgunHttp())
	e.Any("railgun/note_audit", NoteSrv.NoteAuditRailgunHttp())
	if err := e.Start(); err != nil {
		log.Error("hkt-job error(%v)", err)
		panic(err)
	}
}

// innerRouter init inner router.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
}

// ping check server ok.
func ping(c *bm.Context) {}

func initService(c *conf.Config) {
	NoteSrv = note.New(c)
	ArtSrv = article.New(c)
}
