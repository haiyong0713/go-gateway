package http

import (
	abtest "go-common/component/tinker/middleware/http"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	appmid "go-gateway/app/app-svr/app-resource/interface/http/middleware"
	absvr "go-gateway/app/app-svr/app-resource/interface/service/abtest"
	auditsvr "go-gateway/app/app-svr/app-resource/interface/service/audit"
	broadcastsvr "go-gateway/app/app-svr/app-resource/interface/service/broadcast"
	deeplinksvr "go-gateway/app/app-svr/app-resource/interface/service/deeplink"
	displaysvr "go-gateway/app/app-svr/app-resource/interface/service/display"
	domainsvr "go-gateway/app/app-svr/app-resource/interface/service/domain"
	entrancesvr "go-gateway/app/app-svr/app-resource/interface/service/entrance"
	fksvr "go-gateway/app/app-svr/app-resource/interface/service/fawkes"
	fpsvr "go-gateway/app/app-svr/app-resource/interface/service/fingerprint"
	fissisvr "go-gateway/app/app-svr/app-resource/interface/service/fission"
	guidesvc "go-gateway/app/app-svr/app-resource/interface/service/guide"
	locationsvr "go-gateway/app/app-svr/app-resource/interface/service/location"
	"go-gateway/app/app-svr/app-resource/interface/service/mod"
	"go-gateway/app/app-svr/app-resource/interface/service/notice"
	"go-gateway/app/app-svr/app-resource/interface/service/param"
	pingsvr "go-gateway/app/app-svr/app-resource/interface/service/ping"
	pluginsvr "go-gateway/app/app-svr/app-resource/interface/service/plugin"
	privacysvr "go-gateway/app/app-svr/app-resource/interface/service/privacy"
	showsvr "go-gateway/app/app-svr/app-resource/interface/service/show"
	sidesvr "go-gateway/app/app-svr/app-resource/interface/service/sidebar"
	"go-gateway/app/app-svr/app-resource/interface/service/splash"
	staticsvr "go-gateway/app/app-svr/app-resource/interface/service/static"
	"go-gateway/app/app-svr/app-resource/interface/service/version"
	whitesvr "go-gateway/app/app-svr/app-resource/interface/service/white"
	"go-gateway/app/app-svr/app-resource/interface/service/widget"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	// depend service
	authSvc *auth.Auth
	// self service
	pgSvr          *pluginsvr.Service
	pingSvr        *pingsvr.Service
	sideSvr        *sidesvr.Service
	verSvc         *version.Service
	paramSvc       *param.Service
	ntcSvc         *notice.Service
	splashSvc      *splash.Service
	auditSvc       *auditsvr.Service
	abSvc          *absvr.Service
	modSvc         *mod.Service
	guideSvc       *guidesvc.Service
	staticSvc      *staticsvr.Service
	domainSvc      *domainsvr.Service
	whiteSvc       *whitesvr.Service
	showSvc        *showsvr.Service
	broadcastSvc   *broadcastsvr.Service
	fingerPrintSvc *fpsvr.Service
	locationSvc    *locationsvr.Service
	fkSvc          *fksvr.Service
	fissionSvc     *fissisvr.Service
	verifySvc      *verify.Verify
	//nolint:unused
	privacySvc  *privacysvr.Service
	displaySvc  *displaysvr.Service
	deeplinkSvc *deeplinksvr.Service
	widgetSvc   *widget.Service
	entranceSvc *entrancesvr.Service
	featureSvc  *feature.Feature
	config      *conf.Config
)

type Server struct {
	// depend service
	AuthSvc *auth.Auth
	// self service
	PgSvr          *pluginsvr.Service
	PingSvr        *pingsvr.Service
	SideSvr        *sidesvr.Service
	VerSvc         *version.Service
	ParamSvc       *param.Service
	NtcSvc         *notice.Service
	SplashSvc      *splash.Service
	AuditSvc       *auditsvr.Service
	AbSvc          *absvr.Service
	ModSvc         *mod.Service
	GuideSvc       *guidesvc.Service
	StaticSvc      *staticsvr.Service
	DomainSvc      *domainsvr.Service
	WhiteSvc       *whitesvr.Service
	ShowSvc        *showsvr.Service
	BroadcastSvc   *broadcastsvr.Service
	FingerPrintSvc *fpsvr.Service
	LocationSvc    *locationsvr.Service
	FkSvc          *fksvr.Service
	FissionSvc     *fissisvr.Service
	VerifySvr      *verify.Verify
	PrivacySvc     *privacysvr.Service
	DisplaySvc     *displaysvr.Service
	DeeplinkSvc    *deeplinksvr.Service
	WidgetSvc      *widget.Service
	EntranceSvc    *entrancesvr.Service
	FeatureSvc     *feature.Feature
	Config         *conf.Config
}

// Init is
func Init(c *conf.Config, svr *Server) {
	initService(c, svr)
	config = c
	// init external router
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(engineOut)
	// init Outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

// initService init services.
func initService(_ *conf.Config, svr *Server) {
	// init self service
	authSvc = svr.AuthSvc
	pgSvr = svr.PgSvr
	pingSvr = svr.PingSvr
	sideSvr = svr.SideSvr
	verSvc = svr.VerSvc
	paramSvc = svr.ParamSvc
	ntcSvc = svr.NtcSvc
	splashSvc = svr.SplashSvc
	auditSvc = svr.AuditSvc
	abSvc = svr.AbSvc
	modSvc = svr.ModSvc
	guideSvc = svr.GuideSvc
	staticSvc = svr.StaticSvc
	domainSvc = svr.DomainSvc
	broadcastSvc = svr.BroadcastSvc
	whiteSvc = svr.WhiteSvc
	showSvc = svr.ShowSvc
	fingerPrintSvc = svr.FingerPrintSvc
	locationSvc = svr.LocationSvc
	fkSvc = svr.FkSvc
	fissionSvc = svr.FissionSvc
	verifySvc = svr.VerifySvr
	privacySvc = svr.PrivacySvc
	displaySvc = svr.DisplaySvc
	deeplinkSvc = svr.DeeplinkSvc
	widgetSvc = svr.WidgetSvc
	entranceSvc = svr.EntranceSvc
	featureSvc = svr.FeatureSvc
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(bm.CORS(), anticrawler.Report())
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-app.bilibili.com")
	r := e.Group("/x/resource")
	{
		r.GET("/peak/download", authSvc.GuestMobile, download)
		r.GET("/abtest/tiny", abtest.Handler(), tinyAbtest)
		r.GET("/plugin", plugin)
		r.GET("/sidebar", authSvc.GuestMobile, sidebar)
		r.GET("/topbar", topbar)
		r.GET("/abtest", abTest)
		r.GET("/abtest/v2", authSvc.GuestMobile, abtest.Handler(), abTestV2)
		r.GET("/abtest/list", authSvc.GuestMobile, abTestList)
		r.GET("/abtest/abserver", authSvc.GuestMobile, abserver)
		r.POST("/fingerprint", authSvc.GuestMobile, fingerprint)
		r.POST("/laser", laserReport)
		r.POST("/laser2", authSvc.GuestMobile, laserReport2)
		r.POST("/laser/silence", laserReportSilence)
		r.POST("/laser/cmd/report", laserCmdReport)
		m := r.Group("/module")
		{
			m.POST("", authSvc.GuestMobile, module)
			m.POST("/list", authSvc.GuestMobile, list)
		}
		g := r.Group("/guide", authSvc.GuestMobile)
		{
			g.GET("/interest", interest)
			g.GET("/interest2", interest2)
		}
		f := r.Group("/fission")
		{
			f.GET("/check/new", verifySvc.Verify, authSvc.UserMobile, checkNew)
			f.GET("/check/device", verifySvc.Verify, checkDevice)
		}
		r.GET("/static", getStatic)
		r.GET("/domain", domain)
		r.GET("/broadcast/servers", serverList)
		r.GET("/white/list", whiteList)
		r.GET("/show/tab", authSvc.GuestMobile, tabs)
		r.GET("/show/tab/v2", authSvc.GuestMobile, abtest.Handler(), tabsV2)
		r.POST("/show/click", verifySvc.Verify, clickTab)
		r.GET("/show/tab/bubble", proxyHandler, authSvc.GuestMobile, tabBubble)
		r.GET("/show/skin", verifySvc.Verify, authSvc.GuestMobile, featureSvc.BuildLimitHttp(), skin)
		r.GET("/ip", IpInfo)
		r.GET("/top/activity", verifySvc.Verify, authSvc.GuestMobile, topActivity)
		r.GET("/pop/up", verifySvc.Verify, authSvc.GuestMobile, indexPopUp)
		d := r.Group("/deeplink")
		{
			d.GET("/huawei", deeplinkHW)
			d.GET("/button", deeplinkButton)
			d.GET("/ai", deeplinkAi)
		}
		r.GET("/widget", widgets)
		r.GET("/widget/android", verifySvc.Verify, authSvc.GuestMobile, widgetAndroid)
		r.POST("/entrance/infoc", verifySvc.Verify, authSvc.GuestMobile, entranceInfoc)
		dolby := r.Group("/dolby")
		{
			dolby.GET("/config", dolbyConfig)
		}
		r.GET("/vivo/popular/badge", vivoPopularBadge)
		r.GET("/service/dependencies", serviceDependencies)
	}
	v := e.Group("/x/v2/version", featureSvc.BuildLimitHttp())
	{
		v.GET("", getVersion)
		v.GET("/h5", bm.CSRF(), getVersion)
		v.GET("/update", versionUpdate)
		v.GET("/update.pb", versionUpdatePb)
		v.GET("/so", versionSo)
		v.GET("/rn/update", versionRn)
		v.GET("/fawkes/upgrade", fawkesUpgrade)
		v.GET("/fawkes/upgrade/ios", verifySvc.Verify, fawkesUpgradeIOS)
		v.GET("/fawkes/hotfix/upgrade", fawkesHfUpgrade)
		v.GET("/fawkes/bizapk", apkList)
		v.GET("/fawkes/tribe", tribeList)
		v.GET("/testflight", authSvc.GuestMobile, testFlight)
		v.GET("/fawkes/upgrade/tiny", fawkesTinyUpgrade)

	}
	p := e.Group("/x/v2/param", authSvc.GuestMobile)
	{
		p.GET("", getParam)
	}
	n := e.Group("/x/v2/notice", authSvc.GuestMobile)
	{
		n.GET("", getNotice)
		n.GET("/package", getPackagePushMsg)
	}
	s := e.Group("/x/v2/splash", appmid.InjectTimestamp())
	{
		s.GET("", splashs)
		s.GET("/birthday", birthSplash)
		s.GET("/list", authSvc.GuestMobile, splashList)
		s.GET("/state", authSvc.GuestMobile, splashState)
		s.GET("/show", authSvc.GuestMobile, splashRtShow)
		b := s.Group("/brand", verifySvc.Verify, featureSvc.BuildLimitHttp())
		{
			b.GET("/list", authSvc.GuestMobile, brandList)
			b.GET("/set", authSvc.GuestMobile, brandSet)
			b.POST("/save", authSvc.GuestMobile, brandSave)
		}
		s.GET("/event/list", authSvc.GuestMobile, eventList)
		s.GET("/event/list2", authSvc.GuestMobile, eventList2)
	}
	a := e.Group("/x/v2/audit")
	{
		a.GET("", audit)
	}
	oldDis := e.Group("/x/display", bm.CSRF())
	{
		oldDis.GET("/id", displayId)
		oldDis.GET("/wechat/sign", wechatAuth)
	}
}
