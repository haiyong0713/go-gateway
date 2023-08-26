package http

import (
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/http/blademaster/render"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/service/mobile"
	"go-gateway/app/app-svr/app-wall/interface/service/offer"
	"go-gateway/app/app-svr/app-wall/interface/service/operator"
	pingSvr "go-gateway/app/app-svr/app-wall/interface/service/ping"
	"go-gateway/app/app-svr/app-wall/interface/service/telecom"
	"go-gateway/app/app-svr/app-wall/interface/service/unicom"
	"go-gateway/app/app-svr/app-wall/interface/service/wall"
)

var (
	// depend service
	verifySvc *verify.Verify
	authSvc   *auth.Auth
	// self service
	wallSvc     *wall.Service
	offerSvc    *offer.Service
	unicomSvc   *unicom.Service
	mobileSvc   *mobile.Service
	pingSvc     *pingSvr.Service
	telecomSvc  *telecom.Service
	operatorSvc *operator.Service
)

type Server struct {
	// depend service
	VerifySvc *verify.Verify
	AuthSvc   *auth.Auth
	// self service
	WallSvc     *wall.Service
	OfferSvc    *offer.Service
	UnicomSvc   *unicom.Service
	MobileSvc   *mobile.Service
	PingSvc     *pingSvr.Service
	TelecomSvc  *telecom.Service
	OperatorSvc *operator.Service
}

func Init(c *conf.Config, svr *Server) {
	initService(svr)
	// init external router
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(c, engineOut)
	// init Outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v)", err)
		panic(err)
	}
}

// initService init services.
func initService(svr *Server) {
	verifySvc = svr.VerifySvc
	authSvc = svr.AuthSvc
	// init self service
	wallSvc = svr.WallSvc
	offerSvc = svr.OfferSvc
	unicomSvc = svr.UnicomSvc
	mobileSvc = svr.MobileSvc
	pingSvc = svr.PingSvc
	telecomSvc = svr.TelecomSvc
	operatorSvc = svr.OperatorSvc
}

func outerRouter(c *conf.Config, e *bm.Engine) {
	e.Use(bm.CORS(), anticrawler.Report())
	e.Ping(ping)
	// formal api
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-app.bilibili.com")
	w := e.Group("/x/wall")
	{
		w.GET("/get", walls)
		op := w.Group("/operator", authSvc.Guest)
		{
			op.GET("/ip", userOperatorIP)
			op.GET("/ip/info", operatorIPInfo)
			op.GET("/m/ip", mOperatorIP)
			op.GET("/reddot", reddot)
		}
		of := w.Group("/offer")
		{
			of.GET("/exist", wallExist)
			of.POST("/click/shike", proxyHandler, wallShikeClick)
			of.GET("/click/dotinapp", wallDotinappClick)
			of.GET("/click/gdt", wallGdtClick)
			of.GET("/click/toutiao", wallToutiaoClick)
			of.POST("/active", proxyHandler, verifySvc.Verify, wallActive)
			of.GET("/active/test", wallTestActive)
			of.POST("/active2", proxyHandler, verifySvc.Verify, wallActive2)
		}
		uc := w.Group("/unicom", proxyHandler)
		{
			// unicomSync
			// 订单同步
			uc.POST("/orders", ordersSync)
			// deprecated
			uc.POST("/advance", advanceSync)
			// deprecated
			uc.POST("/flow", flowSync)
			// deprecated
			uc.POST("/ip", inIPSync)
			// coupon verify
			uc.POST("/coupon/verify", couponVerify)
			// unicom
			uc.GET("/userflow", verifySvc.Verify, userFlow)
			uc.GET("/user/userflow", userFlowState)
			uc.GET("/userstate", verifySvc.Verify, userState)
			uc.GET("/state", verifySvc.Verify, authSvc.GuestMobile, unicomState)
			uc.GET("/m/state", unicomStateM)
			uc.POST("/pack", authSvc.User, pack)
			uc.GET("/userip", isUnciomIP)
			uc.GET("/user/ip", userUnciomIP)
			uc.POST("/order/pay", orderPay)
			uc.POST("/order/cancel", orderCancel)
			uc.POST("/order/smscode", authSvc.Guest, smsCode)
			uc.POST("/order/bind", authSvc.User, bindUser)
			uc.POST("/order/untie", authSvc.User, unbindUser)
			uc.GET("/bind/info", authSvc.Guest, userBind)
			uc.GET("/pack/list", authSvc.Guest, packList)
			uc.POST("/order/pack/receive", authSvc.User, packReceive)
			uc.POST("/order/pack/flow", authSvc.User, flowPack)
			uc.GET("/order/userlog", authSvc.User, userBindLog)
			uc.GET("/pack/log", userPacksLog)
			uc.GET("/bind/state", verifySvc.Verify, welfareBindState)
			uc.GET("/bind/info/phone", userBindInfoByPhone)
			uc.POST("/bind/add/integral", addUserBindIntegral)
			uc.GET("/flow/sign", flowSign)
			uc.GET("/activate", verifySvc.Verify, authSvc.GuestMobile, activate)
			uc.GET("/active/state", verifySvc.Verify, authSvc.GuestMobile, unicomActiveState)
			uc.POST("/order/flow/tryout", verifySvc.Verify, authSvc.GuestMobile, unicomFlowTryout)
		}
		mb := w.Group("/mobile", proxyHandler)
		{
			mb.POST("/orders.so", ordersMobileSync)
			// deprecated
			mb.GET("/activation", verifySvc.Verify, mobileActivation)
			mb.GET("/status", verifySvc.Verify, authSvc.GuestMobile, mobileState)
			mb.GET("/user/status", userMobileState)
			mb.GET("/active/state", verifySvc.Verify, authSvc.GuestMobile, mobileActiveState)
		}
		tl := w.Group("/telecom", proxyHandler)
		{
			tl.POST("/orders.so", telecomOrdersSync)
			tl.POST("/flow.so", telecomMsgSync)
			tl.POST("/order/pay", telecomPay)
			tl.POST("/order/pay/cancel", cancelRepeatOrder)
			tl.GET("/order/consent", verifySvc.Verify, orderConsent)
			tl.GET("/order/list", verifySvc.Verify, orderList)
			tl.GET("/order/user/flow", phoneFlow)
			tl.POST("/send/sms", verifySvc.Verify, phoneSendSMS)
			tl.GET("/verification", verifySvc.Verify, phoneVerification)
			tl.GET("/order/state", orderState)
			tl.POST("/card/orders", telecomCardOrdersSync)
			tl.GET("/card/state", authSvc.GuestMobile, telecomCardOrder)
			tl.GET("/card/verification", authSvc.GuestMobile, telecomCardCodeOrder)
			tl.POST("/card/sms", phoneCardVerification)
			tl.GET("/pack/log/vip", vipPacksLog)
			tl.GET("/active/state", verifySvc.Verify, authSvc.GuestMobile, telecomActiveState)
		}
	}
}

// returnDataJSON return json no message
func returnDataJSON(c *bm.Context, data map[string]interface{}, err error) {
	code := http.StatusOK
	if data == nil {
		c.JSON(data, err)
		return
	}
	if _, ok := data["message"]; !ok {
		data["message"] = ""
	}
	if err != nil {
		c.Error = err
		bcode := ecode.Cause(err)
		data["code"] = bcode.Code()
	} else {
		if _, ok := data["code"]; !ok {
			data["code"] = ecode.OK
		}
		data["ttl"] = 1
	}
	c.Render(code, render.MapJSON(data))
}

func Close() {
	wallSvc.Close()
	unicomSvc.Close()
}
