package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/queue/databus"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	appmid "go-gateway/app/app-svr/app-resource/interface/http/middleware"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/service/report"
	"go-gateway/app/app-svr/app-view/interface/service/view"
	"go-gateway/app/app-svr/app-view/interface/tools"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	viewSvr   *view.Service
	reportSvr *report.Service
	authSvr   *auth.Auth
	verifySvc *verify.Verify
	// databus
	userActPub *databus.Databus
	dislikePub *databus.Databus
	cfg        *conf.Config
	featureSvr *feature.Feature
)

type Server struct {
	ViewSvr   *view.Service
	ReportSvr *report.Service
	AuthSvr   *auth.Auth
	VerifySvc *verify.Verify
	// databus
	UserActPub *databus.Databus
	DislikePub *databus.Databus
	FeatureSvc *feature.Feature
}

type userAct struct {
	Client   string `json:"client"`
	Buvid    string `json:"buvid"`
	Mid      int64  `json:"mid"`
	Time     int64  `json:"time"`
	From     string `json:"from"`
	Build    string `json:"build"`
	ItemID   string `json:"item_id"`
	ItemType string `json:"item_type"`
	Action   string `json:"action"`
	ActionID string `json:"action_id"`
	Extra    string `json:"extra"`
	IsRisk   string `json:"is_risk"`
}

type cmDislike struct {
	ID         int64  `json:"id"`
	Buvid      string `json:"buvid"`
	Goto       string `json:"goto"`
	Mid        int64  `json:"mid"`
	ReasonID   int64  `json:"reason_id"`
	CMReasonID int64  `json:"cm_reason_id"`
	UpperID    int64  `json:"upper_id"`
	Rid        int64  `json:"rid"`
	TagID      int64  `json:"tag_id"`
	ADCB       string `json:"ad_cb"`
	State      int64  `json:"state"`
}

// Init init http
func Init(c *conf.Config, svr *Server) {
	cfg = c
	initService(svr)
	// init external router
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(engineOut)
	// init outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v)", err)
		panic(err)
	}
}

func initService(svr *Server) {
	verifySvc = svr.VerifySvc
	authSvr = svr.AuthSvr
	viewSvr = svr.ViewSvr
	reportSvr = svr.ReportSvr
	// databus
	userActPub = svr.UserActPub
	dislikePub = svr.DislikePub
	featureSvr = svr.FeatureSvc
}

func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(anticrawler.Report())
	// view
	view := e.Group("/x/v2/view", appmid.InjectTimestamp(), featureSvr.BuildLimitHttp(),
		tools.CheckMid64Support(viewSvr.VersionMapClient.AppkeyVersion))
	view.GET("", verifySvc.Verify, authSvr.GuestMobile, arcmid.BatchPlayArgs(), viewIndex)
	view.GET("/page", verifySvc.Verify, authSvr.GuestMobile, viewPage)
	view.GET("/video/shot", verifySvc.Verify, videoShot)
	view.POST("/share/add", verifySvc.Verify, authSvr.GuestMobile, addShare)
	view.GET("/share/icon", verifySvc.Verify, authSvr.UserMobile, shareIcon)
	view.POST("/coin/add", authSvr.UserMobile, addCoin)
	view.POST("/ad/dislike", authSvr.GuestMobile, adDislike)
	view.GET("/report", verifySvc.Verify, copyWriter)
	view.POST("/report/add", authSvr.UserMobile, addReport)
	view.POST("/report/upload", verifySvc.Verify, upload)
	view.POST("/like", authSvr.UserMobile, like)
	view.POST("/dislike", authSvr.UserMobile, dislike)
	view.POST("/vip/playurl", authSvr.UserMobile, vipPlayURL)
	view.GET("/follow", authSvr.GuestMobile, follow)
	view.GET("/upper/recmd", authSvr.GuestMobile, upperRecmd)
	view.POST("/like/triple", authSvr.UserMobile, likeTriple)
	view.GET("/material", verifySvc.Verify, material)
	view.POST("/share/click", verifySvc.Verify, authSvr.GuestMobile, shareClick)
	view.POST("/share/complete", verifySvc.Verify, authSvr.GuestMobile, shareComplete)
	view.POST("/like/nologin", authSvr.GuestMobile, likeNoLogin)
	view.GET("/video/download", verifySvc.Verify, authSvr.GuestMobile, videoDownload)
	view.POST("/ad/dislike/cancel", authSvr.GuestMobile, adDislikeCancel)
	// ar
	view.POST("/ar/do", verifySvc.Verify, authSvr.UserMobile, doAr)
	//三点下的内容
	view.GET("/dots", verifySvc.Verify, authSvr.GuestMobile, dots)
	//展示在线观看人数
	view.GET("/video/online", verifySvc.Verify, authSvr.GuestMobile, videoOnline)
	//互动弹幕投票
	view.POST("/dm/vote", verifySvc.Verify, authSvr.UserMobile, dmVote)
	//stat数据
	view.GET("/stat", verifySvc.Verify, authSvr.GuestMobile, arcStat)
	view.POST("/user/action/add", verifySvc.Verify, authSvr.GuestMobile, addUserAction)

	//首映分享落地页查询
	corsFn := func(c *bm.Context) {
		bm.CORS().ServeHTTP(c)
	}
	view.GET("/share/info", corsFn, shareInfo)
	view.OPTIONS("/share/info", corsFn)
	//付费UGC
	trade := view.Group("/trade")
	{
		trade.GET("/product/info", verifySvc.Verify, authSvr.GuestMobile, productInfo)
		trade.GET("/order/state", verifySvc.Verify, authSvr.UserMobile, orderState)
		trade.POST("/order/create", verifySvc.Verify, authSvr.UserMobile, orderCreate)
	}
}
