package http

import (
	"reflect"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/app-svr/fawkes/service/api/app/auth"
	"go-gateway/app/app-svr/fawkes/service/api/app/open"
	"go-gateway/app/app-svr/fawkes/service/api/app/tribe"
	"go-gateway/app/app-svr/fawkes/service/api/app/webcontainer"
	"go-gateway/app/app-svr/fawkes/service/conf"
	apmSvr "go-gateway/app/app-svr/fawkes/service/service/apm"
	appSvr "go-gateway/app/app-svr/fawkes/service/service/app"
	authSvr "go-gateway/app/app-svr/fawkes/service/service/auth"
	bizapkSvr "go-gateway/app/app-svr/fawkes/service/service/bizapk"
	buglySvr "go-gateway/app/app-svr/fawkes/service/service/bugly"
	busSvr "go-gateway/app/app-svr/fawkes/service/service/business"
	cdSvr "go-gateway/app/app-svr/fawkes/service/service/cd"
	ciSvr "go-gateway/app/app-svr/fawkes/service/service/ci"
	configSvr "go-gateway/app/app-svr/fawkes/service/service/config"
	feedbackSvr "go-gateway/app/app-svr/fawkes/service/service/feedback"
	ffSvr "go-gateway/app/app-svr/fawkes/service/service/ff"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	laserSvr "go-gateway/app/app-svr/fawkes/service/service/laser"
	mngSvr "go-gateway/app/app-svr/fawkes/service/service/manager"
	modSvr "go-gateway/app/app-svr/fawkes/service/service/mod"
	mdlSvr "go-gateway/app/app-svr/fawkes/service/service/modules"
	openSvr "go-gateway/app/app-svr/fawkes/service/service/open"
	pingSvr "go-gateway/app/app-svr/fawkes/service/service/ping"
	prometheusSvr "go-gateway/app/app-svr/fawkes/service/service/prometheus"
	statisticsSvr "go-gateway/app/app-svr/fawkes/service/service/statistics"
	tribeSvr "go-gateway/app/app-svr/fawkes/service/service/tribe"
	webContainerSvr "go-gateway/app/app-svr/fawkes/service/service/webcontainer"
	taskSvr "go-gateway/app/app-svr/fawkes/service/task"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/middleware"
)

var (
	verifySvc *verify.Verify
	permitSvc *permit.Permit
	s         *Servers
	cfg       *conf.Config
	OpenPaths []string
)

// Servers all service.
type Servers struct {
	ApmSvr          *apmSvr.Service
	AppSvr          *appSvr.Service
	BizapkSvr       *bizapkSvr.Service
	BusSvr          *busSvr.Service
	CDSvr           *cdSvr.Service
	CiSvr           *ciSvr.Service
	ConfigSvr       *configSvr.Service
	FFSvr           *ffSvr.Service
	GitSvr          *gitSvr.Service
	LaserSvr        *laserSvr.Service
	MngSvr          *mngSvr.Service
	MdlSvr          *mdlSvr.Service
	ModSvr          *modSvr.Service
	PrometheusSvr   *prometheusSvr.Service
	PingSvr         *pingSvr.Service
	StatisticsSvr   *statisticsSvr.Service
	FeedbackSvr     *feedbackSvr.Service
	BuglySvr        *buglySvr.Service
	TribeSvr        *tribeSvr.Service
	OpenSvr         *openSvr.Service
	TaskSvr         *taskSvr.Service
	AuthSvr         *authSvr.Service
	WebContainerSvr *webContainerSvr.Service
}

// Init http init
func Init(c *conf.Config, ss *Servers) {
	s = ss
	cfg = c
	verifySvc = verify.New(nil)
	permitSvc = permit.New2(nil)
	engineInner := bm.DefaultServer(c.HTTPServers.Inner)
	registerBMServer(engineInner)
	innerRouter(engineInner)
	initOpenApi(engineInner, ss.OpenSvr)
	// init internal server
	if err := engineInner.Start(); err != nil {
		log.Error("engineInner.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

// 注册bm的相关server
func registerBMServer(e *bm.Engine) {
	e.Inject("app/tribe/", permitSvc.Verify2(), middleware.ContextValues())
	e.Inject("business/tribe/", middleware.ContextValues())
	e.Inject("/x/admin/fawkes/app/open/", permitSvc.Verify2(), middleware.ContextValues())
	e.Inject("/x/admin/fawkes/auth", permitSvc.Verify2(), middleware.ContextValues())
	e.Inject("/x/admin/fawkes/app/webcontainer/", permitSvc.Verify2(), middleware.ContextValues())
	tribe.RegisterTribeBMServer(e, s.TribeSvr)
	open.RegisterOpenBMServer(e, s.OpenSvr)
	auth.RegisterAuthBMServer(e, s.AuthSvr)
	webcontainer.RegisterWhiteListBMServer(e, s.WebContainerSvr)
}

func initOpenApi(engine *bm.Engine, openSvr *openSvr.Service) {
	openSvr.OpenPaths = reflect.ValueOf(engine).Elem().FieldByName("metastore").MapKeys()
}

// Fawkes 统一校验入口
func FawkesVerify() bm.HandlerFunc {
	handlers := []bm.HandlerFunc{
		permitSvc.Verify2(),
		middleware.VisitorCheck(),
		middleware.AuthVerify(), // 权限点校验
	}
	return func(ctx *bm.Context) {
		for _, h := range handlers {
			h(ctx)
			if ctx.IsAborted() {
				break
			}
		}
	}
}

// innerRouter init outer router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(bm.CORS())
	af := e.Group("/x/admin/fawkes")
	{
		app := af.Group("/app")
		{
			// 应用信息
			app.GET("", FawkesVerify(), appInfo)
			app.GET("/list", FawkesVerify(), appList)
			app.POST("/add", FawkesVerify(), appAdd)
			app.POST("/edit", FawkesVerify(), appEdit)
			app.POST("/update/highest_peak", FawkesVerify(), appUpdateIsHighestPeak)

			// 应用审核
			app.GET("/audit/list", FawkesVerify(), appAuditList)
			app.POST("/audit", FawkesVerify(), appAudit)
			// 关注列表
			app.GET("/follow/list", FawkesVerify(), appFollowList)
			app.POST("/follow/add", FawkesVerify(), appFollowAdd)
			app.POST("/follow/del", FawkesVerify(), appFollowDel)
			// 应用渠道
			app.GET("/channels", FawkesVerify(), appChannelList)
			app.GET("/channels/list", FawkesVerify(), appChannelListV2)
			app.POST("/channels/add", FawkesVerify(), appChannelAdd)
			app.POST("/channels/del", FawkesVerify(), appChannelDelete)
			app.POST("/channels/group/relate", FawkesVerify(), appChannelGroupRelate)
			app.GET("/channels/group/list", FawkesVerify(), appChannelGroupList)
			app.POST("/channels/group/add", FawkesVerify(), appChannelGroupAdd)
			app.POST("/channels/group/update", FawkesVerify(), appChannelGroupUpdate)
			app.POST("/channels/group/del", FawkesVerify(), appChannelGroupDel)
			// 平台公告
			app.GET("/notification/list", FawkesVerify(), appNotificationList)
			app.POST("/notification/update", FawkesVerify(), appNotificationUpdate)
			app.POST("/notification/add", FawkesVerify(), appNotificationAdd)
			// 邮件通知
			app.GET("/mailto/list", FawkesVerify(), appMailtoList)
			app.POST("/mailto/update", FawkesVerify(), appMailtoUpdate)
			app.POST("/mail/config/add", FawkesVerify(), appMailConfigAdd)
			app.POST("/mail/config/del", FawkesVerify(), appMailConfigDel)
			app.POST("/mail/config/update", FawkesVerify(), appMailConfigUpdate)
			app.GET("/mail/config/list", FawkesVerify(), appMailConfigList)
			app.GET("/mail/list", FawkesVerify(), appMailList)
			// 企业微信通知
			app.POST("/wxapp/notify", FawkesVerify(), appWXAppNotify)
			app.POST("/wxapp/picnotify", FawkesVerify(), appWXAppPictureNotify)
			// 机器人管理
			app.POST("/robot/notify", FawkesVerify(), appRobotNotify)
			app.POST("/robot/set", FawkesVerify(), appRobotSet)
			app.POST("/robot/upload", FawkesVerify(), appRobotUpload)
			app.GET("/robot/list", FawkesVerify(), appRobotList)
			app.POST("/robot/add", FawkesVerify(), appRobotAdd)
			app.POST("/robot/update", FawkesVerify(), appRobotUpdate)
			app.POST("/robot/del", FawkesVerify(), appRobotDel)
			// 其他
			app.GET("/key", FawkesVerify(), appKeys)
			app.GET("/system", FawkesVerify(), system)
			app.GET("/branch", FawkesVerify(), branchTagList)
			app.POST("/trigger/pipeline", FawkesVerify(), appTriggerPipeline)
			app.POST("/file/upload", FawkesVerify(), appFileUpload) // bfs上传
			app.POST("/service/ping", FawkesVerify(), appServicePing)

			// CI 相关接口
			ci := app.Group("/ci")
			{
				ci.GET("/info", appCIPackInfo)
				ci.GET("/list", FawkesVerify(), buildPackList)
				ci.POST("/add", FawkesVerify(), createBuildPack)              // 老的CI构建目前有部分其他业务方还在脚本使用
				ci.POST("/common/add", FawkesVerify(), createBuildPackCommon) // 平台构建全部使用common/add
				ci.POST("/cancel", FawkesVerify(), cancelBuildPack)
				ci.POST("/del", FawkesVerify(), deleteBuildPack)
				ci.POST("/notifygroup", FawkesVerify(), notifyGroup)
				ci.POST("/parse/mainbbr", FawkesVerify(), parseBBR)
				// 环境变量管理
				ci.GET("/env/list", FawkesVerify(), ciEnvList)
				ci.POST("/env/add", FawkesVerify(), addCiEnv)
				ci.POST("/env/update", FawkesVerify(), UpdateCiEnv)
				ci.POST("/env/del", FawkesVerify(), DeleteCiEnv)
				ci.POST("/env/delbyappkey", FawkesVerify(), DeleteCiEnvByAppKey)
				// 定时任务
				ci.GET("/crontab/list", FawkesVerify(), ciCrontabList)
				ci.POST("/crontab/add", FawkesVerify(), ciCrontabAdd)
				ci.POST("/crontab/status", FawkesVerify(), ciCrontabStatus)
				ci.POST("/crontab/del", FawkesVerify(), ciCrontabDel)
				// 自动化测试
				ci.GET("/monkey/list", FawkesVerify(), getMonkeyList)
				ci.POST("/monkey/add", FawkesVerify(), addMonkey)
				ci.POST("/monkey/update/status", updateMonkeyStatus)
				// 其他
				ci.GET("/job/info", FawkesVerify(), ciJobInfo)
				ci.GET("/pack/report/info", FawkesVerify(), packReportInfo)
				ci.POST("/dependency/publish", FawkesVerify(), publishDepandency)
				ci.POST("/version/info", FawkesVerify(), getAppBuildPackVersionInfo)

				// Git pipline 调用
				ci.POST("/record", recordBuildPack)
				ci.POST("/update", updateBuildPackInfo)
				ci.POST("/update/status", updateBuildPackStatus)
				ci.POST("/upload", uploadBuildPack)
				ci.POST("/upload/buildfile", uploadBuildFile)
				ci.POST("/upload/mobile/ep/business", uploadMobileEPBusiness)
				ci.POST("/sendmail", sendmail)
				ci.POST("/subrepo/pushhook", repoPushHook)
				ci.POST("/subrepo/mrhook", subRepoMRHook)
				ci.POST("/mainrepo/pushhook", repoPushHook)
				ci.POST("/mainrepo/mrhook", mainRepoMRHook)
				ci.POST("/mainrepo/commenthook", mainRepoCommentHook)
				ci.POST("/mainrepo/branchhook", releaseBranchHook)
				ci.POST("/mainrepo/rebuild", mainRepoRebuild)
				ci.GET("/mainrepo/build", mainRepoBuild)
				ci.GET("/mainrepo/buildstatus", pipelineStatus)
				ci.POST("/mainrepo/checkout", checkoutBranch)
				ci.POST("/lint/mrhook", lintMRHook)
				ci.POST("/mr/create", relatedMRCreate)
				ci.GET("/branch/commit", branchCommit)
				// 废弃接口
				ci.POST("/test/update", updateTestStatus)
			}
			// CD 相关接口
			cd := app.Group("/cd")
			{
				cd.POST("/portal/test", FawkesVerify(), appPortalTest)
				cd.GET("/list", FawkesVerify(), appCDList)
				cd.GET("/list/filter", FawkesVerify(), appCDListFilter)
				// ？？
				cd.GET("/versions", FawkesVerify(), appCDVersions)
				cd.GET("/builds", FawkesVerify(), appCDBuilds)
				cd.GET("/hotfix/versions", FawkesVerify(), appCDHotfixVersions)
				// 升级配置
				cd.POST("/config/pack/steadystate/set", FawkesVerify(), appCDPackSteadyStateSet)
				cd.POST("/config/switch/set", FawkesVerify(), appCDConfigSwitchSet)
				cd.POST("/config/upgrad/set", FawkesVerify(), appCDUpgradConfigSet)
				cd.GET("/config/upgrad", FawkesVerify(), appCDUpgradConfig)
				cd.POST("/config/filter/set", FawkesVerify(), appCDFilterConfigSet)
				cd.GET("/config/filter", FawkesVerify(), appCDFilterConfig)
				cd.POST("/config/flow/set", FawkesVerify(), appCDFlowConfigSet)
				cd.GET("/config/flow", FawkesVerify(), appCDFlowConfig)
				cd.POST("/evolution", FawkesVerify(), appCDEvolution)
				// 渠道包
				cd.GET("/generate/list", FawkesVerify(), appCDGenerate)
				cd.POST("/generate/add", FawkesVerify(), appCDGenerateAdd)
				cd.POST("/generate/add/git", FawkesVerify(), appCDGenerateAddGit)
				cd.POST("/generate/adds", FawkesVerify(), appCDGenerateAdds)
				cd.POST("/generate/status", FawkesVerify(), appCDGenerateStatus)
				cd.POST("/generate/upload", FawkesVerify(), appCDGenerateUpload)
				cd.POST("/generate/publish", FawkesVerify(), appCDGeneratePublish)
				cd.GET("/generate/publish/list", appCDGeneratePublishList) // 渠道包外部反查
				// 自建渠道包
				cd.GET("/customchannel/list", FawkesVerify(), appCDCustomChannelList)
				cd.POST("/customchannel/add", FawkesVerify(), appCDCustomChannelAdd)
				cd.POST("/customchannel/upload", FawkesVerify(), appCDCustomChannelUpload)
				// Patch包
				cd.GET("/patch/list", FawkesVerify(), appCDPatchList)
				cd.POST("/patch/build", FawkesVerify(), appCDPatchBuild)
				// Testflight
				cd.POST("/testflight/app/set", FawkesVerify(), testflightAppSet)
				cd.GET("/testflight/app", FawkesVerify(), testflightAppInfo)
				cd.GET("/testflight/testing", testflightTestingPack) // 内部脚本服务
				cd.POST("/testflight/betareview", FawkesVerify(), testflightBetaReview)
				cd.POST("/testflight/distribute", FawkesVerify(), testflightDistribute)
				cd.POST("/testflight/stop", FawkesVerify(), testflightStop)
				cd.POST("/testflight/remindupdate", FawkesVerify(), testflightRemindUpdate)
				cd.POST("/testflight/forceupdate", FawkesVerify(), testflightForceUpdate)
				cd.POST("/testflight/betagroups/set", FawkesVerify(), testflightBetagroupSet)
				cd.POST("/testflight/setupdtxt", FawkesVerify(), testflightSetUpdTxt)
				cd.POST("/testflight/bwlist/add", FawkesVerify(), testflightBWAdd)
				cd.GET("/testflight/bwlist/list", FawkesVerify(), testflightBWList)
				cd.POST("/testflight/bwlist/del", FawkesVerify(), testflightBWDel)
				cd.POST("/testflight/uploadbugly", FawkesVerify(), testflightUploadBugly)
				// 资源包推送
				cd.POST("/assets/evolution", assetsEvolution) // 内部脚本调用
				// Windows发布
				cd.POST("/windows/appinstaller/upload", FawkesVerify(), windowsAppinstallerUpload)
				cd.POST("/windows/appinstaller/publish", FawkesVerify(), windowsAppinstallerPublish)
				// 其他
				cd.POST("/sync/macross", FawkesVerify(), appCDSyncMacross)
				cd.POST("/sync/manager", FawkesVerify(), appCDSyncManager)
				cd.POST("/cdn/refresh", FawkesVerify(), appCDRefreshCDN)
				cd.POST("/release/notify", FawkesVerify(), releaseNotify)
				// 灰度包信息查询
				cd.GET("/pack/grey/list", FawkesVerify(), packGreyList)
				// 废弃接口
				// cd.POST("/generate/teststate/set", FawkesVerify(), appCDGenerateTestStateSet)
				cd.POST("/cdn/publish", FawkesVerify(), appCDCDNPublish)
			}
			// business apk 相关接口
			bizapk := app.Group("/bizapk")
			{
				bizapk.GET("/list", FawkesVerify(), bizapkBuildsList)
				bizapk.POST("/add", FawkesVerify(), bizApkAdd)
				bizapk.POST("/cancel", FawkesVerify(), bizApkCancel)
				bizapk.POST("/del", FawkesVerify(), bizApkDelete)
				bizapk.POST("/evolution", FawkesVerify(), bizApkEvolution)
				bizapk.GET("/settings/list", FawkesVerify(), bizApkList)
				bizapk.POST("/settings/set", FawkesVerify(), bizApkSet)
				bizapk.POST("/config/filter/set", FawkesVerify(), bizApkFilterConfigSet)
				bizapk.GET("/config/filter", FawkesVerify(), bizApkFilterConfig)
				bizapk.POST("/config/flow/set", FawkesVerify(), bizApkFlowConfigSet)
				bizapk.GET("/config/flow", FawkesVerify(), bizApkFlowConfig)
				bizapk.POST("/upload", bizApkUpload)
				bizapk.POST("/update", bizApkUpdate)
			}
			// modules 相关接口
			modules := app.Group("/modules")
			{
				modules.POST("/group/add", FawkesVerify(), groupAdd)
				modules.POST("/group/change", FawkesVerify(), groupChange)
				modules.POST("/group/edit", FawkesVerify(), groupEdit)
				modules.POST("/group/del", FawkesVerify(), groupDel)
				modules.GET("/list", moduleGroupList)
				modules.GET("/list/groups", FawkesVerify(), listGroups)
				modules.GET("/list/sizetype", FawkesVerify(), listSizeType)
				modules.GET("/size/module", FawkesVerify(), sizeModule)
				modules.GET("/size/group", FawkesVerify(), sizeGroup)
				modules.GET("/size/version", FawkesVerify(), groupsSizeInBuild)
				modules.GET("/size/groupversion", FawkesVerify(), modulesSizeInGroupVersion)
				modules.POST("/config/totalsize/set", FawkesVerify(), modulesConfTotalSizeSet)
				modules.POST("/config/set", FawkesVerify(), modulesConfSet)
				modules.GET("/config", FawkesVerify(), modulesConfGet)
				// Git 上报
				modules.POST("/size/record", sizeRecord)
			}
			// hotfix 相关接口
			hf := app.Group("/hotfix")
			{
				hf.GET("/list", FawkesVerify(), hotfixList)
				hf.POST("/build", FawkesVerify(), hotfixBuild)
				hf.POST("/cancel", FawkesVerify(), hotfixCancel)
				hf.POST("/del", FawkesVerify(), hotfixDel)
				hf.POST("/evolution", FawkesVerify(), hotfixPushEnv)
				hf.POST("/config/set", FawkesVerify(), hotfixConfSet)
				hf.GET("/config", FawkesVerify(), hotfixConfGet)
				hf.POST("/effect", FawkesVerify(), hotfixEffect)
				// Git 上报
				hf.POST("/update", hotfixUpdate)
				hf.POST("/upload", hotfixUpload)
				hf.GET("/origin/get", hotfixOrigGet)
			}
			// config 相关接口
			cf := app.Group("/config")
			{
				cf.GET("", FawkesVerify(), appConfig)
				cf.GET("/list", FawkesVerify(), appConfigVersionList)
				cf.POST("/add", FawkesVerify(), appConfigVersionAdd)
				cf.POST("/fastadd", FawkesVerify(), appConfigFastAdd)
				cf.POST("/del", FawkesVerify(), appConfigVersionDel)
				cf.POST("/save", FawkesVerify(), appConfigSave)
				cf.GET("/diff", FawkesVerify(), appConfigDiff)
				cf.POST("/publish", FawkesVerify(), appConfigPublish)
				cf.GET("/publish/diff", FawkesVerify(), appConfigPublishView)
				cf.POST("/publish/multiple", FawkesVerify(), appConfigPublishMultiple)
				cf.GET("/history", appConfigVersionHistory)
				cf.GET("/historys", FawkesVerify(), appConfigVersionHistorys)
				cf.GET("/history/cid", FawkesVerify(), appConfigVersionHistoryByID)
				cf.GET("/file", FawkesVerify(), appConfigFile)
				cf.GET("/modify/count", FawkesVerify(), appConfigModifyCount)
				cf.GET("/key/publish/history", FawkesVerify(), appConfigKeyPublishHistory)
				// cf.GET("/paladin/fe", FawkesVerify(), appPaladinFeConfig)
			}
			// ab 相关接口
			ff := app.Group("/ff")
			{
				ff.GET("", FawkesVerify(), appFFConfig)
				ff.GET("/list", FawkesVerify(), appFFList)
				ff.POST("/set", FawkesVerify(), appFFConfigSet)
				ff.POST("/del", FawkesVerify(), appFFConfigDel)
				ff.POST("/publish", FawkesVerify(), appFFPublish)
				ff.GET("/diff", FawkesVerify(), appFFDiff)
				ff.GET("/publish/history", FawkesVerify(), appFFHistory)
				ff.GET("/publish/history/ffid", FawkesVerify(), appFFHistoryByID)
				ff.GET("/publish/diff", FawkesVerify(), appFFPublishDiff)
				ff.GET("/whithlist", FawkesVerify(), appFFWhithlist)
				ff.POST("/whithlist/add", FawkesVerify(), appFFWhithlistAdd)
				ff.POST("/whithlist/del", FawkesVerify(), appFFWhithlistDel)
				ff.GET("/modify/count", FawkesVerify(), appFFModifyCount)
			}
			ls := app.Group("/laser")
			{
				ls.GET("/list", FawkesVerify(), appLaserList)
				ls.POST("/add", FawkesVerify(), appLaserAdd)
				ls.POST("/del", FawkesVerify(), appLaserDel)
				ls.GET("/active/list", FawkesVerify(), appLaserActiveList)
				ls.GET("/command/list", FawkesVerify(), appLaserCmdList)
				ls.POST("/command/add", FawkesVerify(), appLaserCmdAdd)
				ls.POST("/command/del", FawkesVerify(), appLaserCmdDel)
				ls.GET("/command/action/list", FawkesVerify(), appLaserCmdActionList)
				ls.POST("/command/action/add", FawkesVerify(), appLaserCmdActionAdd)
				ls.POST("/command/action/update", FawkesVerify(), appLaserCmdActionUpdate)
				ls.POST("/command/action/del", FawkesVerify(), appLaserCmdActionDel)
				// ???
				// ls.POST("/parsestatus/update", appLaserParseStatusUpdate)
				// ls.GET("/broadcast/pending/list", appLaserPendingList)
			}
			mod := app.Group("/mod")
			{
				// 这三个接口存在早上10点定时外部调用，为了避免影响业务运行，暂时忽略校验
				mod.GET("/pool/list", permitSvc.Verify2(), modPoolList)
				mod.GET("/module/list", permitSvc.Verify2(), modModuleList)
				mod.GET("/version/list", permitSvc.Verify2(), modVersionList)
				// 访问强校验
				mod.GET("/patch/list", FawkesVerify(), modPatchList)
				mod.POST("/version/add", FawkesVerify(), modVersionAdd)
				mod.POST("/version/release", FawkesVerify(), modVersionRelease)
				mod.POST("/version/release/check", FawkesVerify(), modVersionReleaseCheck)
				mod.POST("/version/push", FawkesVerify(), modVersionPush)
				mod.GET("/version/config", FawkesVerify(), modVersionConfig)
				mod.POST("/version/config/add", FawkesVerify(), modVersionConfigAdd)
				mod.GET("/version/gray", FawkesVerify(), modVersionGray)
				mod.POST("/version/gray/whitelist/upload", FawkesVerify(), modGrayWhitelistUpload)
				mod.POST("/version/gray/add", FawkesVerify(), modVersionGrayAdd)
				mod.POST("/module/delete", FawkesVerify(), modModuleDelete)
				mod.POST("/module/state", FawkesVerify(), modModuleState)
				mod.POST("/module/update", FawkesVerify(), modModuleUpdate)
				mod.POST("/module/add", FawkesVerify(), modModuleAdd)
				mod.POST("/module/push/offline", FawkesVerify(), modModulePushOffline)
				mod.POST("/pool/add", FawkesVerify(), modPoolAdd)
				mod.POST("/pool/update", FawkesVerify(), modPoolUpdate)
				mod.GET("/permission/list", FawkesVerify(), modPermissionList)
				mod.POST("/permission/add", FawkesVerify(), modPermissionAdd)
				mod.POST("/permission/delete", FawkesVerify(), modPermissionDelete)
				mod.GET("/permission/role", FawkesVerify(), modPermissionRole)
				mod.POST("/global/push", FawkesVerify(), modGlobalPush)
				mod.POST("/version/apply/add", FawkesVerify(), modVersionApplyAdd)
				mod.GET("/version/apply/notify", FawkesVerify(), modVersionApplyNotify)
				mod.GET("/version/apply/list", FawkesVerify(), modVersionApplyList)
				mod.GET("/version/apply/overview", FawkesVerify(), modVersionApplyOverview)
				mod.POST("/version/apply/pass", FawkesVerify(), modVersionApplyPass)
				mod.POST("/version/apply/refuse", FawkesVerify(), modVersionApplyRefuse)
				mod.GET("/sync/pool", FawkesVerify(), modSyncPool)
				mod.GET("/sync/version/list", FawkesVerify(), modSyncVersionList)
				mod.POST("/sync/add", FawkesVerify(), modSyncAdd)
				mod.GET("/sync/version/info", FawkesVerify(), modSyncVersionInfo)

				open := mod.Group("/open")
				{
					open.GET("/pool/list", verifySvc.Verify, permitSvc.Verify2(), modOpenPoolList)
					open.GET("/module/list", verifySvc.Verify, permitSvc.Verify2(), modOpenModuleList)
					open.GET("/version", verifySvc.Verify, permitSvc.Verify2(), modOpenVersion)
					open.GET("/version/list", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionList)
					open.GET("/patch/list", verifySvc.Verify, permitSvc.Verify2(), modOpenPatchList)
					open.POST("/module/add", verifySvc.Verify, permitSvc.Verify2(), modOpenModuleAdd)
					open.POST("/module/state", verifySvc.Verify, permitSvc.Verify2(), modOpenModuleState)
					open.POST("/module/push/offline", verifySvc.Verify, permitSvc.Verify2(), modOpenModulePushOffline)
					open.POST("/version/add", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionAdd)
					open.POST("/version/release", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionRelease)
					open.POST("/version/push", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionPush)
					open.GET("/version/config", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionConfig)
					open.POST("/version/config/add", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionConfigAdd)
					open.GET("/version/gray", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionGray)
					open.POST("/version/gray/add", verifySvc.Verify, permitSvc.Verify2(), modOpenVersionGrayAdd)
					open.POST("/version/gray/whitelist/upload", verifySvc.Verify, permitSvc.Verify2(), modOpenGrayWhitelistUpload)
				}
				outer := open.Group("/outer")
				{
					outer.GET("/pool/list", verifySvc.Verify, modOpenPoolList)
					outer.GET("/module/list", verifySvc.Verify, modOpenModuleList)
					outer.GET("/version", verifySvc.Verify, modOpenVersion)
					outer.GET("/version/list", verifySvc.Verify, modOpenVersionList)
					outer.GET("/patch/list", verifySvc.Verify, modOpenPatchList)
					outer.POST("/module/add", verifySvc.Verify, modOpenModuleAdd)
					outer.POST("/module/state", verifySvc.Verify, modOpenModuleState)
					outer.POST("/module/push/offline", verifySvc.Verify, modOpenModulePushOffline)
					outer.POST("/version/add", verifySvc.Verify, modOpenVersionAdd)
					outer.POST("/version/release", verifySvc.Verify, modOpenVersionRelease)
					outer.POST("/version/push", verifySvc.Verify, modOpenVersionPush)
					outer.GET("/version/config", verifySvc.Verify, modOpenVersionConfig)
					outer.POST("/version/config/add", verifySvc.Verify, modOpenVersionConfigAdd)
					outer.GET("/version/gray", verifySvc.Verify, modOpenVersionGray)
					outer.POST("/version/gray/add", verifySvc.Verify, modOpenVersionGrayAdd)
					outer.POST("/version/gray/whitelist/upload", verifySvc.Verify, modOpenGrayWhitelistUpload)
				}
			}
			fb := app.Group("/feedback", FawkesVerify())
			{
				fb.GET("/list", FeedbackList)
				fb.POST("/add", FeedbackAdd)
				fb.POST("/update", FeedbackUpdate)
				fb.GET("/info", FeedbackInfo)
				fb.POST("/del", FeedbackDel)
				fb.POST("/tapd/bug/create", FeedBackTapdBugCreate)
			}
		}
		// 静态渠道相关
		channel := af.Group("/channels")
		{
			channel.GET("", FawkesVerify(), channelList)
			channel.POST("/add", FawkesVerify(), channelAdd)
			channel.POST("/del", FawkesVerify(), channelDelete)
		}
		// 管理信息
		mng := af.Group("/manager")
		{
			mng.GET("/tree/auth", FawkesVerify(), treeAuth)
			mng.GET("/tree/auths", FawkesVerify(), treeAuths)
			mng.GET("/tree/list", FawkesVerify(), treeList)
			mng.GET("/auth/user/list", FawkesVerify(), authUserList)
			mng.GET("/auth/user/list/role", FawkesVerify(), authUserListByRole)
			mng.POST("/auth/user/set", FawkesVerify(), authUserSet)
			mng.POST("/auth/user/del", FawkesVerify(), authUserDel)
			mng.GET("/auth/user", FawkesVerify(), authUser)
			mng.GET("/auth/role", FawkesVerify(), authRole)
			mng.GET("/auth/supervisor", FawkesVerify(), authSupervisor)
			mng.GET("/auth/supervisor/role", FawkesVerify(), authSupervisorRole)
			mng.GET("/auth/role/apply/user", FawkesVerify(), authRoleApply)
			mng.GET("/auth/role/apply/list", FawkesVerify(), authRoleApplyList)
			mng.POST("/auth/role/apply/add", FawkesVerify(), authRoleApplyAdd)
			mng.POST("/auth/role/apply/pass", FawkesVerify(), authRoleApplyPass)
			mng.POST("/auth/role/apply/refuse", FawkesVerify(), authRoleApplyRefuse)
			mng.POST("/auth/role/apply/ignore", FawkesVerify(), authRoleApplyIgnore)
			mng.POST("/event/apply/add", FawkesVerify(), eventApplyAdd)
			mng.POST("/event/apply/recall", FawkesVerify(), eventApplyRecall)
			mng.POST("/bfs/cdn/refresh", FawkesVerify(), bfsRefreshCDN)
			mng.POST("/auth/message/push", FawkesVerify(), authMessagePush)
			mng.GET("/auth/user/name/list", FawkesVerify(), userNameList)
			mng.POST("/auth/user/name/add", FawkesVerify(), userNameSet)
			mng.POST("/auth/admin/apply", FawkesVerify(), authAdminApply)
			// 疑似废弃
			mng.GET("/log/list", logList)
			mng.POST("/auth/nickname/set", authNickNameSet)
		}
		mngMod := mng.Group("/mod")
		{
			mngMod.GET("/role/apply/list", FawkesVerify(), modRoleApplyList)
			mngMod.POST("/role/add", FawkesVerify(), modRoleAdd)
			mngMod.POST("/role/apply/add", FawkesVerify(), modRoleApplyAdd)
			mngMod.GET("/role/apply/process", FawkesVerify(), modRoleApplyProcess)
			mngMod.GET("/role/operator/list", FawkesVerify(), modRoleOperatorList)
			mngMod.POST("/role/apply/pass", FawkesVerify(), modRoleApplyPass)
			mngMod.POST("/role/apply/refuse", FawkesVerify(), modRoleApplyRefuse)
		}
		monitor := af.Group("/apm")
		{
			monitor.GET("/bus/list", FawkesVerify(), apmBusList)
			monitor.POST("/bus/add", FawkesVerify(), apmBusAdd)
			monitor.POST("/bus/del", FawkesVerify(), apmBusDel)
			monitor.POST("/bus/update", FawkesVerify(), apmBusUpdate)
			monitor.GET("/command/group/advanced/list", FawkesVerify(), apmCommandGroupAdvancedList)
			monitor.POST("/command/group/advanced/add", FawkesVerify(), apmCommandGroupAdvancedAdd)
			monitor.POST("/command/group/advanced/del", FawkesVerify(), apmCommandGroupAdvancedDel)
			monitor.POST("/command/group/advanced/update", FawkesVerify(), apmCommandGroupAdvancedUpdate)
			monitor.GET("/command/group/list", FawkesVerify(), apmCommandGroupList)
			monitor.POST("/command/group/add", FawkesVerify(), apmCommandGroupAdd)
			monitor.POST("/command/group/del", FawkesVerify(), apmCommandGroupDel)
			monitor.POST("/command/group/update", FawkesVerify(), apmCommandGroupUpdate)
			monitor.GET("/command/list", FawkesVerify(), apmCommandList)
			monitor.GET("/moni/calculate", FawkesVerify(), apmMoniCalculate)
			monitor.POST("/moni/line", FawkesVerify(), apmMoniLine)
			monitor.POST("/moni/pie", FawkesVerify(), apmMoniPie)
			// 列表详情
			monitor.POST("/moni/metric/info/list", FawkesVerify(), apmMoniMetricInfoList)
			monitor.POST("/moni/count/info/list", FawkesVerify(), apmMoniCountInfoList)
			monitor.GET("/moni/net/info/list", FawkesVerify(), apmMoniNetInfoList)
			monitor.GET("/moni/statistics/info/list", FawkesVerify(), apmMoniStatisticsInfoList)

			monitor.GET("/moni/aggregate/net/list", FawkesVerify(), apmAggregateNetList)
			monitor.GET("/moni/aggregate/crash/list", FawkesVerify(), apmAggregateCrashList)
			monitor.GET("/moni/aggregate/anr/list", FawkesVerify(), apmAggregateANRList)
			monitor.GET("/moni/aggregate/setup/list", FawkesVerify(), apmAggregateSetupList)
			monitor.GET("/moni/flawmap/route/list", FawkesVerify(), apmFlowmapRouteList)
			monitor.GET("/moni/detail/setup", FawkesVerify(), apmDetailSetup)
			monitor.GET("/flowmap/route/alias/list", FawkesVerify(), apmFlowmapRouteAliasList)
			monitor.POST("/flowmap/route/alias/add", FawkesVerify(), apmFlowmapRouteAliasAdd)
			monitor.POST("/flowmap/route/alias/update", FawkesVerify(), apmFlowmapRouteAliasUpdate)
			monitor.POST("/flowmap/route/alias/del", FawkesVerify(), apmFlowmapRouteAliasDel)
			monitor.POST("/fawkes/track", FawkesVerify(), apmWebTrack)
			// monitor.POST("/fawkes/track", FawkesVerify(), apmWebTrack)
			event := monitor.Group("event")
			{
				event.GET("", FawkesVerify(), apmEvent)
				event.GET("/list", FawkesVerify(), apmEventList)
				event.POST("/add", FawkesVerify(), apmEventAdd)
				event.POST("/del", FawkesVerify(), apmEventDel)
				event.POST("/update", FawkesVerify(), apmEventUpdate)
				event.POST("/field/set", FawkesVerify(), apmEventFieldSet)
				event.POST("/field/type/sync", FawkesVerify(), apmEventFieldTypeSync) // 临时接口
				event.GET("/field/list", FawkesVerify(), apmEventFieldList)
				event.POST("/field/publish", FawkesVerify(), apmEventFieldPublish)
				event.GET("/field/publish/history", FawkesVerify(), apmEventFieldPublishHistory)
				event.GET("/field/publish/diff", FawkesVerify(), apmEventFieldPublishDiff)
				event.GET("/field/diff", FawkesVerify(), apmEventFieldDiff)
				event.POST("/field/billions/sync", FawkesVerify(), apmEventFieldBillionsSync) // 日志平台 同步event field
				event.POST("/table/ck/create", FawkesVerify(), apmEventCKTableCreate)
				event.GET("/convert/sql", FawkesVerify(), apmEventSql)
				event.GET("/advanced/list", FawkesVerify(), apmEventAdvancedList)
				event.POST("/advanced/add", FawkesVerify(), apmEventAdvancedAdd)
				event.POST("/advanced/del", FawkesVerify(), apmEventAdvancedDel)
				event.POST("/advanced/update", FawkesVerify(), apmEventAdvancedUpdate)
				event.GET("/setting", FawkesVerify(), apmEventSetting)
				// app内部event接口
				event.GET("/app/list", FawkesVerify(), apmAppEventList)
				event.POST("/app/relation/add", FawkesVerify(), apmAppEventRelAdd)
				// app内部基础字段的事件组
				event.GET("/app/commonfield/group", FawkesVerify(), apmAppCommonFieldGroup)
				event.GET("/app/commonfield/group/list", FawkesVerify(), apmAppCommonFieldGroupList)
				event.POST("/app/commonfield/group/add", FawkesVerify(), apmAppCommonFieldGroupAdd)
				event.POST("/app/commonfield/group/update", FawkesVerify(), apmAppCommonFieldGroupUpdate)
				event.POST("/app/commonfield/group/del", FawkesVerify(), apmAppCommonFieldGroupDel)
				// 日志告警操作
				event.GET("/alert", FawkesVerify(), apmEventAlert)
				event.GET("/alert/list", FawkesVerify(), apmEventAlertList)
				event.POST("/alert/add", FawkesVerify(), apmEventAlertAdd)
				event.POST("/alert/update", FawkesVerify(), apmEventAlertUpdate)
				event.POST("/alert/del", FawkesVerify(), apmEventAlertDel)
				event.POST("/alert/switch", FawkesVerify(), apmEventAlertSwitch)
				event.POST("/samplerate/add", FawkesVerify(), apmEventSampleRateAdd)
				event.GET("/samplerate/list", FawkesVerify(), apmEventSampleRateList)
				event.POST("/samplerate/delete", FawkesVerify(), apmEventSampleRateDel)
				event.GET("/samplerate/config", FawkesVerify(), apmEventSampleRateConfig)
				// 埋点监控配置
				event.GET("/monitor/notify/config", FawkesVerify(), apmEventMonitorNotifyConfig)
				event.GET("/monitor/notify/config/list", FawkesVerify(), apmEventMonitorNotifyConfigList)
				event.POST("/monitor/notify/config/set", FawkesVerify(), apmEventMonitorNotifyConfigSet)
			}
			prometheus := monitor.Group("protheusme")
			{
				prometheus.GET("/metric/list", FawkesVerify(), apmMetricList)
				prometheus.POST("/metric/set", FawkesVerify(), apmMetricSet)
				prometheus.POST("/metric/del", FawkesVerify(), apmMetricDel)
				prometheus.POST("/metric/publish", FawkesVerify(), apmMetricPublish)
				prometheus.GET("/metric/publish/list", FawkesVerify(), apmMetricPublishList)
				prometheus.GET("/metric/publish/diff", FawkesVerify(), apmMetricPublishDiff)
				prometheus.POST("/metric/publish/rollback", FawkesVerify(), apmMetricPublishRollback)
			}
			flinkJob := monitor.Group("/flink/job")
			{
				flinkJob.GET("/list", FawkesVerify(), apmFlinkJobList)
				flinkJob.POST("/add", FawkesVerify(), apmFlinkJobAdd)
				flinkJob.POST("/update", FawkesVerify(), apmFlinkJobUpdate)
				flinkJob.POST("/del", FawkesVerify(), apmFlinkJobDel)
				flinkJob.GET("/relation/list", FawkesVerify(), apmFlinkJobRelationList)
				flinkJob.POST("/relation/add", FawkesVerify(), apmFlinkJobRelationAdd)
				flinkJob.POST("/relation/del", FawkesVerify(), apmFlinkJobRelationDel)
				flinkJob.POST("/publish", FawkesVerify(), apmFlinkJobPublish)
				flinkJob.GET("/publish/list", FawkesVerify(), apmFlinkJobPublishList)
				flinkJob.GET("/publish/diff", FawkesVerify(), apmFlinkJobPublishDiff)
			}
			crashRule := monitor.Group("/crash/rule")
			{
				crashRule.GET("", FawkesVerify(), apmCrashRule)
				crashRule.GET("/list", FawkesVerify(), apmCrashRuleList)
				crashRule.POST("/add", FawkesVerify(), apmCrashRuleAdd)
				crashRule.POST("/del", FawkesVerify(), apmCrashRuleDel)
				crashRule.POST("/update", FawkesVerify(), apmCrashRuleUpdate)
			}
			alertRule := monitor.Group("/alert/rule")
			{
				alertRule.GET("/list", FawkesVerify(), apmAlertRuleList)
				alertRule.POST("/set", FawkesVerify(), apmAlertRuleSet)
				alertRule.POST("/del", FawkesVerify(), apmAlertRuleDel)
			}
			alert := monitor.Group("/alert")
			{
				alert.GET("/indicator/info", FawkesVerify(), apmAlertIndicatorInfo)
				alert.GET("/list", FawkesVerify(), apmAlertList)
				alert.POST("/add", FawkesVerify(), apmAlertAdd)
				alert.POST("/update", FawkesVerify(), apmAlertUpdate)
				alert.GET("/reason", FawkesVerify(), apmAlertReason)
				// 告警根因配置相关接口
				alert.GET("/reason/config", FawkesVerify(), apmAlertReasonConfig)
				alert.POST("/reason/config/add", FawkesVerify(), apmAlertReasonConfigAdd)
				alert.POST("/reason/config/update", FawkesVerify(), apmAlertReasonConfigUpdate)
				alert.POST("/reason/config/delete", FawkesVerify(), apmAlertReasonConfigDelete)
			}
		}
		// 统计接口
		statistics := app.Group("/statistics")
		{
			statistics.GET("/line", statisticsLine)
			statistics.GET("/pie", statisticsPie)
			statistics.GET("/common/info/list", commonInfoList)
		}
		// Veda
		veda := app.Group("/veda")
		{
			veda.POST("/crash/index/list", FawkesVerify(), crashIndexList)
			veda.POST("/crash/info/list", FawkesVerify(), crashInfoList)
			veda.POST("/crash/index/update", FawkesVerify(), updateCrashIndex)
			veda.POST("/crash/laser/relation/add", FawkesVerify(), crashLaserRelationAdd)
			veda.POST("/jank/index/list", FawkesVerify(), jankIndexList)
			veda.POST("/jank/info/list", FawkesVerify(), jankInfoList)
			veda.POST("/jank/index/update", FawkesVerify(), updateJankIndex)
			veda.POST("/oom/index/list", FawkesVerify(), oomIndexList)
			veda.POST("/oom/info/list", FawkesVerify(), oomInfoList)
			veda.POST("/oom/index/update", FawkesVerify(), updateOOMIndex)
			veda.POST("/index/status/update", FawkesVerify(), updateIndex)
			veda.GET("/index/status", FawkesVerify(), solveStatus)
			veda.GET("/crash/log/list", FawkesVerify(), crashLogList)
		}
		// 业务接口
		bus := af.Group("/business")
		{
			// app-resource
			bus.GET("/config/version", newestVersion)  // 6000+ ( 以下均为15分钟的请求量 ）
			bus.GET("/version/all", versionAll)        // 1100+
			bus.GET("/upgrade/all", upgradeAll)        // 1100+
			bus.GET("/pack/all", packAll)              // 1100+
			bus.GET("/filter/all", filterAll)          // 1100+
			bus.GET("/patch/all2", patchAll)           // 1100+
			bus.GET("/channel/all", channelAll)        // 1100+
			bus.GET("/flow/all", flowAll)              // 1100+
			bus.GET("/hotfix/all", hotfixAll)          // 1100+
			bus.GET("/laser", laser)                   // 1100+
			bus.GET("/bizapk/list/all", bizApkListAll) // 1100+
			bus.GET("/tribe/list/all", tribeListAll)
			bus.GET("/tribe/relation/all", tribeRelationAll)
			bus.GET("/app/useable/tribes", appUseableTribes)
			bus.GET("/tribe/hosts", tribeHosts)
			bus.GET("/testflight", testflightAll)                 // 2400+
			bus.POST("/laser/report", laserReport)                // 5
			bus.POST("/laser/report2", laserReport2)              // 1100+
			bus.POST("/laser/report/silence", laserReportSilence) // 0
			bus.POST("/laser/cmd/report", laserCmdReport)         // 300+
			bus.GET("/mod/appkey/list", modAppKeyList)
			bus.GET("/mod/appkey/file/list", modAppkeyFileList)

			// git-pipline
			bus.GET("/pack/latestStable", packLatestStable)
			bus.GET("/cd/generates/publish", generatesPunlish)
			bus.POST("/cd/generates/update", generatesUpdate)
			bus.POST("/patch/upload", patchUpload)
			bus.POST("/patch/setStatus", patchSetStatus)
			bus.POST("/ci/job/record", ciJobRecord)
			bus.POST("/ci/compile/record", ciCompileRecord)

			// 活动业务平台 -- 活动配置同步至Config （待迁移）
			bus.GET("/config/default", defaultConfig)
			bus.GET("/config/default/history", defaultConfigHistory)
			bus.POST("/config/default/add", addDefaultConfig)
			bus.POST("/config/manager/share/add", addManagerShareConfig)

			// ipdb
			bus.GET("/ip/parse", parseIP)

			// 废弃接口 （ 部分迁移至 openapi ）
			// bus.POST("/channel", verifySvc.Verify, appChannelList)
			// bus.GET("/patch/all", patchAll)
			// bus.GET("/patch/all4", patchAll4) // pach/all2待替换接口
			// bus.GET("/patch/all/cache", patchAll3)   // 测试patchAll3
			// bus.GET("/patch/all/nocache", patchAll2) // 测试patchAll2
			// bus.POST("/cd/generate/teststate/set", generateTestStateSet)
			// bus.POST("/hawkeye/webhook/crash", hawkeyeWebhookCrash)
			// bus.GET("/list", appCDList)
			// bus.GET("/gitlab/getfile", gitlabGetFile)
			// bus.GET("/veda/crash/index/list", crashIndexListByHashList)
			// bus.GET("/ci/list", buildPackList)
			// bus.GET("/ci/subrepo/list", buildPackSubRepoList)
			// bus.GET("/modules/size/groupversion", middleware.ReqLog(), modulesSizeInGroupVersion)
			// bus.GET("/cd/list/filter", middleware.ReqLog(), middleware.ReqHeaderCheck(), appCDList)
		}
		// 开放接口
		openapi := af.Group("/openapi", middleware.Logger(), middleware.OpenAuth(), middleware.AccessControl())
		{
			// APP信息
			openapi.GET("/app", appInfo)
			openapi.GET("/app/list", appList)

			// APP渠道包
			openapi.GET("/channels", appChannelList) // 待废弃
			openapi.GET("/app/channels", appChannelList)
			openapi.POST("/app/channels/add", busAppChannelAdd)

			// 文件上传
			openapi.POST("/file/upload", appFileUpload) // 视频云点播

			// 企微应用通知
			openapi.POST("/app/wxapp/notify", appWXAppNotify)
			openapi.POST("/app/wxapp/picnotify", appWXAppPictureNotify)

			// 机器人配置信息
			openapi.GET("/robot/list", appRobotList)
			openapi.POST("/robot/notify", appRobotNotify)

			// 包体积信息
			openapi.GET("/modules/config", modulesConfGet)
			openapi.GET("/modules/size/groupversion", modulesSizeInGroupVersion)

			// CI信息
			openapi.GET("/ci/info", appCIPackInfo)
			openapi.GET("/ci/list", buildPackList)
			openapi.GET("/ci/subrepo/list", buildPackSubRepoList)
			openapi.POST("/ci/common/add", createBuildPackCommon) // 灰度版本 数据平台上使用
			openapi.POST("/ci/track", ciTrack)                    // ci日志上报

			// CD发版信息
			openapi.GET("/cd/list/filter", appCDListFilter)
			openapi.GET("/cd/version/list", appCDVersionList)
			openapi.POST("/cd/generate/pack/upload", appCDGeneratePackUpload)
			openapi.GET("/cd/pack/grey/list", packGreyList)

			// Config FF
			openapi.GET("/app/ff", appFFConfig) // 待废弃
			openapi.GET("/ff", appFFConfig)
			openapi.POST("/config/add", addConfig)
			openapi.GET("/config/publish/history", appConfigVersionHistorys)
			openapi.GET("/ff/publish/history", appFFHistory)

			// SRE 修改降级 fallback_list
			openapi.POST("/config/sre/fallback/set", appConfigSreFallbackSet)
			openapi.POST("/config/sre/fallback/publish", appConfigPublishDefault) // TODO: Config发布隔离后替换

			// 数据平台修改 埋点事件采样率
			openapi.POST("/config/datacenter/samplerate/sync", appConfigDCSampleRateSync)
			openapi.POST("/config/datacenter/publish", appConfigPublishDefault) // TODO: Config发布隔离后替换

			// Laser信息
			openapi.GET("/laser/user", laserUser)
			openapi.GET("/laser/active/list", appLaserActiveList)

			// APM
			openapi.GET("/apm/event", apmEvent)
			openapi.GET("/apm/event/list", apmEventList)
			openapi.GET("/apm/event/monitor/notify/config/list", apmEventMonitorNotifyConfigList)
			openapi.POST("/apm/event/add", apmEventAdd)
			openapi.POST("/apm/event/update", apmEventUpdate)
			openapi.POST("/apm/event/field/billions/sync", apmEventFieldBillionsSync)
			openapi.POST("/apm/event/app/relation/add", apmAppEventRelAdd)

			openapi.POST("/apm/field/set", setApmEventField)

			openapi.GET("/apm/crash/index/list", crashIndexListByHashList)
			openapi.POST("/apm/crash/list", crashIndexList)
			openapi.POST("/apm/crash/info/list", crashInfoList)
			openapi.POST("/apm/oom/index/list", oomIndexList)
			openapi.GET("/apm/jank/index/list", jankIndexListByHashList)
			openapi.GET("/apm/moni/detail/setup", apmDetailSetup)
			openapi.GET("/apm/moni/calculate", apmMoniCalculate)

			openapi.GET("/apm/alert/rule/list", apmAlertRuleList)
			openapi.GET("/apm/alert/reason/config", apmAlertReasonConfig)
			openapi.POST("/apm/alert/add", apmAlertAdd)

			fb := openapi.Group("/feedback")
			{
				fb.GET("/list", FeedbackList)
				fb.POST("/add", FeedbackAdd)
				fb.POST("/tapd/bug/create", FeedBackTapdBugCreate)
			}

			// Comet 工单插件
			openapi.POST("/webcontainer/whitelist/add", AddWhiteList)
			openapi.POST("/event/samplerate/add", apmEventSampleRateAdd)
			openapi.POST("/event/samplerate/delete", apmEventSampleRateDel)

			// pcdn
			openapi.GET("/pcdn/file/list", pcdnFileList)
			openapi.POST("/pcdn/file/set", addPcdnFile)
		}
		// railgun定时任务
		railgun := af.Group("/railgun")
		{
			railgun.Any("/user/reload", userInfoReloadTask)

			railgun.Any("/nas/clean", nasCIPackDeleteTask)
			railgun.Any("/nas/clean/patch", nasPatchPackDeleteTask)
			railgun.Any("/nas/clean/channel", nasChannelPackDeleteTask)

			railgun.Any("/tribe/pack/move", tribePackMoveTask)

			railgun.Any("/apm/event/monitor", apmEventMonitorTask)
			railgun.Any("/apm/event/completion", apmEventCompletionTask)
			railgun.Any("/apm/event/monitor/notify/update", apmEventMonitorNotifyConfigTask)

			railgun.Any("/apm/veda/status/update", apmVedaStatusUpdate)

			railgun.Any("/pcnd/bender/resource/sync", benderResourceSyncTask)
		}
	}
}
