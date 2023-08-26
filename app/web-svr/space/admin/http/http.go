package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/space/admin/conf"
	"go-gateway/app/web-svr/space/admin/service"
)

var (
	spcSvc *service.Service
	//idfSvc  *identify.Identify
	permitSvc *permit.Permit
	vfySvc    *verify.Verify
)

// Init init http sever instance.
func Init(c *conf.Config, s *service.Service) {
	spcSvc = s
	permitSvc = permit.New2(nil)
	vfySvc = verify.New(c.Verify)
	engine := bm.DefaultServer(c.BM)
	authRouter(engine)
	// init internal server
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func authRouter(e *bm.Engine) {
	e.Ping(func(*bm.Context) {})
	e.Use(bm.CORS())
	group := e.Group("/x/admin/space")
	{
		group.GET("/fans", vfySvc.Verify, fans)
		group.GET("/topphoto/arcs", vfySvc.Verify, topPhotoArcs)

		topPhotoAdmin := group.Group("/topphoto")
		{
			topPhotoAdmin.GET("/list", permitSvc.Permit2("SPACE_PHOTO_LIST"), getPhotoList)
			topPhotoAdmin.GET("/log", permitSvc.Permit2("SPACE_PHOTO_LIST"), vipAuditLogList)
			topPhotoAdmin.POST("/pass", permitSvc.Permit2("SPACE_PHOTO_LIST"), passPhoto)
			topPhotoAdmin.POST("/back", permitSvc.Permit2("SPACE_PHOTO_LIST"), backPhoto)
			topPhotoAdmin.POST("/repass", permitSvc.Permit2("SPACE_PHOTO_LIST"), rePass)

			topPhotoAdmin.GET("/log/action/list", actionLogList)
		}

		sysNotice := group.Group("sysNotice", permitSvc.Permit2("SPACE_SYSTEM_NOTICE"))
		{
			sysNotice.GET("", sysNoticeList)
			sysNotice.POST("add", addSysNotice)
			sysNotice.POST("update", updateSysNotice)
			sysNotice.POST("opt", optSysNotice)
			sysNotice.POST("/uid/add", addSysNoticeUid)
			sysNotice.GET("/uid", SysNoticeUid)
			sysNotice.POST("/uid/del", delSysNoticeUid)
		}

		noticeGroup := group.Group("/notice", permitSvc.Permit2("SPACE_NOTICE"))
		{
			noticeGroup.GET("", notice)
			noticeGroup.POST("/up", noticeUp)
		}
		group.GET("/relation", relation)
		blacklist := group.Group("/blacklist", permitSvc.Permit2("SPACE_BLACKLIST"))
		{
			blacklist.GET("", blacklistIndex)
			blacklist.POST("/add", blacklistAdd)
			blacklist.POST("/update", blacklistUp)
		}
		whitelist := group.Group("/whitelist", permitSvc.Permit2("SPACE_WHITELIST"))
		{
			whitelist.GET("/list", whitelistIndex)
			whitelist.POST("/add", whitelistAdd)
			whitelist.POST("/update", whitelistUp)
			whitelist.POST("/delete", whitelistDel)
		}
		group.POST("/log/add", addLog)
		official := group.Group("/official", permitSvc.Permit2("SPACE_OFFICIAL_DOWNLOAD"))
		{
			official.GET("", officialIndex)
			official.POST("/add", addOfficial)
			official.POST("/update", updateOfficial)
			official.POST("/del", delOfficial)
		}
		group.GET("/channel", permitSvc.Permit2("SPACE_NOTICE"), channel)
		group.GET("/top/arc", permitSvc.Permit2("SPACE_NOTICE"), topArc)
		group.GET("/masterpiece", permitSvc.Permit2("SPACE_NOTICE"), masterpiece)
		group.POST("/clear/msg", permitSvc.Permit2("SPACE_NOTICE"), clearMessage)
		// User native tab页
		usertab := group.Group("/tab", permitSvc.Permit2("SPACE_USER_TAB"))
		{
			usertab.GET("list", userTabList)
			usertab.POST("add", userTabAdd)
			usertab.POST("modify", userTabModify)
			usertab.POST("online", userTabOnline)
			usertab.POST("delete", userTabDelete)
			//usertab.POST("log", userTabLog)
			usertab.GET("info", userMidInfo)
		}
		usertabAdmin := group.Group("/tab/admin", permitSvc.Permit2("SPACE_USER_TAB_ADMIN"))
		{
			usertabAdmin.GET("list", userTabList)
			usertabAdmin.POST("add", userTabAdd)
			usertabAdmin.POST("modify", userTabModify)
			usertabAdmin.POST("online", userTabOnline)
			usertabAdmin.POST("delete", userTabDelete)
			//usertAdminab.POST("log", userTabLog)
			usertabAdmin.GET("info", userMidInfo)
		}
		commercial := group.Group("/tab/commercial")
		{
			commercial.GET("list", userTabList)
			commercial.POST("add", commercialTabAdd)
			commercial.POST("modify", commercialTabModify)
			commercial.POST("online", commercialTabOnline)
			commercial.POST("delete", commercialTabDelete)
			//usertAdminab.POST("log", userTabLog)
			commercial.GET("info", userMidInfo)
		}
		bannedRcmd := group.Group("/banned_rcmd", permitSvc.Permit2("SPACE_BANNED_RCMD"))
		{
			bannedRcmd.GET("list", bannedRcmdList)
			bannedRcmd.POST("add", bannedRcmdAdd)
			bannedRcmd.POST("delete", bannedRcmdDelete)
			//bannedRcmd.POST("midInfo", bannedRcmdSearchMids)
		}
	}
}
