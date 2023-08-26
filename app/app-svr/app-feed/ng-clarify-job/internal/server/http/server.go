package http

import (
	"time"

	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	pb "go-gateway/app/app-svr/app-feed/ng-clarify-job/api"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/internal/model"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/internal/service"

	"github.com/pkg/errors"
)

var svc *service.Service

// New new a bm server.
func New(s pb.AppFeedNGClarifyJobServer) (*bm.Engine, error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}
	svc = s.(*service.Service)
	engine := bm.DefaultServer(&cfg)
	initRouter(engine)
	if err := engine.Start(); err != nil {
		return nil, errors.WithStack(err)
	}
	return engine, nil
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/internal/feed-ng-clarify-job")
	{
		g.POST("/save-session", saveSession)
		g.GET("/archive/*archivePath", downloadURL)
		g.GET("/list/archive/index", listArchvieIndex)
	}
}

func ping(ctx *bm.Context) {}

func saveSession(ctx *bm.Context) {
	req := &model.IndexSession{}
	if err := ctx.BindWith(req, binding.JSON); err != nil {
		return
	}
	ctx.JSON(nil, svc.SaveSession(ctx, req))
}

func downloadURL(ctx *bm.Context) {
	archivePath := ctx.Params.ByName("archivePath")
	req := &struct {
		TTL int64 `form:"ttl" default:"900"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.DownloadURL(ctx, archivePath, req.TTL))
}

func listArchvieIndex(ctx *bm.Context) {
	req := &struct {
		StartTS int64  `form:"start_ts"`
		EndTS   int64  `form:"end_ts"`
		LastKey string `form:"last_key"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	now := time.Now()
	if req.StartTS <= 0 {
		req.StartTS = now.Unix()
	}
	if req.EndTS <= 0 {
		//nolint:gomnd
		req.EndTS = now.Unix() - 3600
	}
	ctx.JSON(svc.ScanArchvieIndex(ctx, req.StartTS, req.EndTS, req.LastKey))
}
