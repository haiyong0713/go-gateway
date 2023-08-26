package http

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	"go-gateway/app/app-svr/hkt-note/interface/service/article"
	"go-gateway/app/app-svr/hkt-note/interface/service/image"
	"go-gateway/app/app-svr/hkt-note/interface/service/note"
	"go-gateway/pkg/idsafe/bvid"
)

var (
	noteSvr     *note.Service
	imgSvr      *image.Service
	artSvr      *article.Service
	authSvc     *auth.Auth
	publicToken string
)

type Server struct {
	NoteSvr *note.Service
	ImgSvr  *image.Service
	ArtSvr  *article.Service
}

// Init int http service
func Init(svr *Server) {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	initService(conf, svr)
	// init internal router
	engineOut := bm.NewServer(conf.HTTPServer)
	engineOut.Use(bm.Recovery(), bm.Trace(), bm.CORS(), bm.CSRF(), bm.Logger(), bm.Mobile())
	outerRouter(engineOut)
	// init internal server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v) | config(%v)", err, conf)
		panic(err)
	}
}

// initService init services.
func initService(c *conf.Config, svr *Server) {
	noteSvr = svr.NoteSvr
	imgSvr = svr.ImgSvr
	artSvr = svr.ArtSvr
	publicToken = c.Bfs.PublicToken
	authSvc = auth.New(c.Auth)
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	note := e.Group("/x/note", authSvc.User)
	{
		//PC 从收藏-公开笔记-跳转视频到稿件播放页调了这个
		note.GET("/list/archive", noteListArc)
		note.POST("/add", noteAdd)
		// PC 从收藏-公开笔记-跳转视频到稿件播放页详情调了这个
		note.GET("/info", noteInfo)
		// PC+APP  收藏-我的笔记列表
		note.GET("/list", noteList)
		note.POST("/del", noteDel)
		note.GET("/is_gray", isGray)
		note.GET("/count", noteCount)
		image := note.Group("/image")
		{
			image.POST("/upload", upload)
			//获取图片(PC端稿件播放页查看笔记
			image.GET("", download)
		}
	}
	publish := e.Group("/x/note/publish", authSvc.Guest)
	{
		// 查看公开笔记详情(PC三点公开笔记列表点击查看笔记；PC+APP 收藏-我的公开笔记列表-点了某个笔记
		// ; APP 收藏-公开笔记-跳转视频到稿件播放页拉起笔记详情）
		publish.GET("/info", publishNoteInfo) // 公开笔记详情
		// PC+ APP 稿件播放页三点获取稿件下的公开笔记列表(APP是点了三点调，PC是进播放页就调）
		publish.GET("/list/archive", publicListInArc) // 稿件下公开笔记列表
		// PC+APP 收藏-我的公开笔记列表
		publish.GET("/list/user", publicListInUser, authSvc.User) // 用户下公开笔记列表
		publish.POST("/del", publishDel, authSvc.User)
	}
	noteWithoutAuth := e.Group("/x/note", authSvc.Guest)
	{
		noteWithoutAuth.GET("/image/public", downloadPub) // 无用户鉴权的图片流
		noteWithoutAuth.GET("/links", links)              // 通用链接
		noteWithoutAuth.GET("/is_forbid", isForbid)
	}
}

// ping check server ok.
func ping(c *bm.Context) {}

func bvArgCheck(aid int64, bv string) (int64, error) {
	res := aid
	if bv != "" {
		var err error
		if res, err = bvid.BvToAv(bv); err != nil {
			log.Error("bvArgCheck bvid.BvToAv(%s) aid(%d) error(%+v)", bv, aid, err)
			return 0, ecode.RequestErr
		}
	}
	if res <= 0 {
		return 0, ecode.RequestErr
	}
	return res, nil
}

func Close() {
	noteSvr.Close()
}
