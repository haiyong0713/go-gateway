package http

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"go-common/library/conf/paladin"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/dance-taiko/interface/api"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/service"
	"go-gateway/app/web-svr/dance-taiko/interface/pkg"
)

var (
	svc     *service.Service
	idfSvc  *verify.Verify
	authSvc *auth.Auth
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	idfSvc = verify.New(nil)
	authSvc = auth.New(nil)
	engine = bm.NewServer(&cfg)
	engine.Use(bm.Recovery(), bm.Logger(), bm.Trace(), bm.Mobile(), pkg.CORS())
	initRouter(engine)
	api.RegisterDanceTaikoBMServer(engine, svc)
	err = engine.Start()

	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/dance")
	{
		game := g.Group("/game")
		{
			game.POST("/stat", idfSvc.Verify, gameStat)
			game.POST("/upload", idfSvc.Verify, upload)
			game.GET("/top_rank", idfSvc.Verify, arcTopRank)
		}
	}
	e.Inject("/x/dance/game/current", idfSvc.Verify)
	e.Inject("/x/dance/game/create")
	e.Inject("/x/dance/game/join", idfSvc.Verify)
	e.Inject("/x/dance/game/start", idfSvc.Verify)
	e.Inject("/x/dance/game/status", idfSvc.Verify)
	e.Inject("/x/dance/game/finish", idfSvc.Verify)
	e.Inject("/x/dance/game/restart", idfSvc.Verify)

	ott := e.Group("/x/dance_ott")
	{
		game := ott.Group("/game", idfSvc.Verify)
		{
			game.GET("/rank", authSvc.Guest, loadRank)
			game.GET("/load", authSvc.Guest, loadKeyFrames)
			game.GET("/qrcode", authSvc.Guest, loadQRCode)
			game.GET("/create", authSvc.Guest, gameCreate)
			game.POST("/start", authSvc.Guest, gameStart)
			game.GET("/status", authSvc.Guest, gameStatus)
			game.GET("/join", authSvc.User, gameJoin)
			game.POST("/finish", authSvc.Guest, gameFinish)
			game.POST("/stat", authSvc.User, ottGameStat)
			game.POST("/pkg_upload", authSvc.Guest, pkgUpload)
			game.POST("/restart", authSvc.Guest, gameRestart)
		}
	}
}

func ping(ctx *bm.Context) {
	svc.Ping(ctx, nil)
}

func upload(c *bm.Context) {
	var (
		fileType string
		body     []byte
		file     multipart.File
		err      error
	)
	if file, _, err = c.Request.FormFile("file"); err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("c.Request.FormFile(\"file\") error(%v)", err)
		return
	}
	defer file.Close()
	if body, err = ioutil.ReadAll(file); err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("ioutil.ReadAll(c.Request.Body) error(%v)", err)
		return
	}
	fileType = http.DetectContentType(body)

	var url string
	url, err = svc.Upload(c, fileType, bytes.NewReader(body))
	if err != nil {
		log.Error("upload file fail")
	}

	c.JSON(struct {
		URL string `json:"url"`
	}{url}, err)
	return
}
