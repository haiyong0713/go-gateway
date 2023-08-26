package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/service"
	"go-gateway/app/web-svr/activity/admin/service/bnj"
	"go-gateway/app/web-svr/activity/admin/service/college"
	"go-gateway/app/web-svr/activity/admin/service/currency"
	"go-gateway/app/web-svr/activity/admin/service/datamanage"
	"go-gateway/app/web-svr/activity/admin/service/domain"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"
	"go-gateway/app/web-svr/activity/admin/service/fit"
	"go-gateway/app/web-svr/activity/admin/service/kfc"
	"go-gateway/app/web-svr/activity/admin/service/lottery"
	"go-gateway/app/web-svr/activity/admin/service/page"
	"go-gateway/app/web-svr/activity/admin/service/prediction"
	"go-gateway/app/web-svr/activity/admin/service/question"
	"go-gateway/app/web-svr/activity/admin/service/rank"
	rankv2 "go-gateway/app/web-svr/activity/admin/service/rank_v3"
	"go-gateway/app/web-svr/activity/admin/service/reward_conf"
	"go-gateway/app/web-svr/activity/admin/service/rewards"
	"go-gateway/app/web-svr/activity/admin/service/s10"
	"go-gateway/app/web-svr/activity/admin/service/system"
	"go-gateway/app/web-svr/activity/admin/service/task"
	"go-gateway/app/web-svr/activity/admin/service/taskv2"
	"go-gateway/app/web-svr/activity/admin/service/vogue"
	"go-gateway/pkg/idsafe/bvid"
)

var (
	actSrv        *service.Service
	exportSrv     *exporttask.Service
	dataMgeSrv    *datamanage.Service
	authSrv       *permit.Permit
	kfcSrv        *kfc.Service
	preSrv        *prediction.Service
	bnjSrv        *bnj.Service
	taskSrv       *task.Service
	currSrv       *currency.Service
	quesSrv       *question.Service
	lotterySrv    *lottery.Service
	vogueSrv      *vogue.Service
	verifySvc     *verify.Verify
	collegeSrv    *college.Service
	s10Svc        *s10.Service
	rankSrv       *rank.Service
	systemSrv     *system.Service
	taskv2Srv     *taskv2.Service
	rankv2Srv     *rankv2.Service
	domainSrv     *domain.Service
	pageSvc       *page.Service
	fitSrv        *fit.Service
	rewardConfSrv *reward_conf.Service
)

// Init init http sever instance.
func Init(c *conf.Config, s *service.Service) {
	actSrv = s
	exportSrv = exporttask.New(c)
	dataMgeSrv = datamanage.New(c)
	verifySvc = verify.New(nil)
	kfcSrv = kfc.New(c)
	preSrv = prediction.New(c)
	bnjSrv = bnj.New(c)
	taskSrv = task.New(c)
	currSrv = currency.New(c)
	quesSrv = question.New(c)
	lotterySrv = lottery.New(c)
	vogueSrv = vogue.New(c)
	authSrv = permit.New2(nil)
	collegeSrv = college.New(c)
	s10Svc = s10.New(c)
	rankSrv = rank.New(c)
	systemSrv = system.New(c)
	taskv2Srv = taskv2.New(c)
	rankv2Srv = rankv2.New(c)
	domainSrv = domain.New(c)
	pageSvc = page.New(c)
	fitSrv = fit.New(c)
	rewardConfSrv = reward_conf.New(c)
	engine := bm.DefaultServer(c.HTTPServer)
	route(engine)
	rewards.Init(c)
	if err := engine.Start(); err != nil {
		log.Error("httpx.Serve error(%v)", err)
		panic(err)
	}
}

func route(e *bm.Engine) {
	e.Ping(ping)
	e.GET("/fix/lottery/gift/task", fixLotteryGiftTask)
	g := e.Group("/x/admin/activity")
	{
		gTunnel := g.Group("/fix/tunnel/group")
		{
			gTunnel.GET("/add", fixTunnelAdd)
			gTunnel.GET("/up", fixTunnelUp)
			gTunnel.GET("/del", fixTunnelDel)
		}
		gappRelation := g.Group("/relation")
		{
			gappRelation.GET("/list", ListActRelation)
			gappRelation.GET("/get", GetActRelation)
			gappRelation.POST("/add", AddActRelation)
			gappRelation.POST("/update", UpdateActRelation)
			gappRelation.POST("/state", StateActRelation)
		}
		s10Group := g.Group("/s10")
		{
			s10Group.GET("/lottery/users", lotteryByRobin)
			s10Group.GET("/user/cost", userCostInfo)
			s10Group.GET("/user/cost/state", updateUserCostState)
			s10Group.GET("/user/cost/ack", ackCostInfo)
			s10Group.GET("/user/cost/cache", userCostCacheFlush)
			s10Group.GET("/user/gift", userGiftInfo)
			s10Group.GET("/user/gift/ack", ackGiftInfo)
			s10Group.GET("/user/gift/cache", userGiftInfoFlush)
			s10Group.GET("/user/cost/redelivery", redeliveryCostInfo)
			s10Group.GET("/gift/redelivery", redeliveryGiftInfo)
			s10Group.POST("/user/import", superUserImport)
			s10Group.GET("/lottery/users/check", checkUserLottery)
			s10Group.GET("/user/real/goods", realGoodsList)
			s10Group.GET("/sentout/goods", sentOutGoods)
			s10Group.GET("/goods/stock/cache", delGoodsStock)
			s10Group.GET("/backup/users", backupUsers)
			s10Group.GET("/backup/generate/users", genBackUsers)
		}
		g.GET("/arcs", archives)
		g.GET("/accounts", accounts)
		gapp := g.Group("/matchs", authSrv.Permit2("ACT_MATCHS_MGT_TEST"))
		{
			gapp.POST("/add", addMatch)
			gapp.POST("/save", saveMatch)
			gapp.GET("/info", matchInfo)
			gapp.GET("/list", matchList)
		}
		gappO := g.Group("/matchs/object", authSrv.Permit2("ACT_MATCHS_MGT_TEST"))
		{
			gappO.POST("/add", addMatchObject)
			gappO.POST("/save", saveMatchObject)
			gappO.GET("/info", matchObjectInfo)
			gappO.GET("/list", matchObjectList)
		}
		domain := g.Group("/domain")
		{
			domain.POST("/add", addDomain)
			domain.POST("/edit", editDomain)
			domain.POST("/stop", stopDomain)
			domain.GET("/search", searchDomain)
			domain.POST("/sync", syncCacheDomain)
		}
		gappSuject := g.Group("/subject")
		{
			gappSuject.GET("/list", listInfosAll)
			gappSuject.GET("/videos", videoList)
			gappSuject.POST("/add", addActSubject)
			gappSuject.POST("/up", updateInfoAll)
			gappSuject.GET("/protocol", subPro)
			gappSuject.GET("/conf", timeConf)
			gappSuject.GET("/articles", article)
			gappSuject.GET("/infos", subInfos)
			gappSuject.GET("/opt/videos/list", optVideoList)
			gappSuject.GET("/rule/list", subjectRules)
			gappSuject.GET("/rule/user/state", subjectRuleUserState)
			gappSuject.POST("/rule/add", addSubjectRule)
			gappSuject.POST("/rule/save", saveSubjectRule)
			gappSuject.POST("/rule/state/up", upSubRuleState)
			gappSuject.POST("/push/add", addPush)
			gappSuject.POST("/push/up", editPush)
			gappSuject.POST("/push/start", startPush)
			gappSuject.GET("/push/info", infoPush)
			gappSuject.GET("/push/template", pushTemplate)
		}
		reserve := g.Group("/reserve")
		{
			reserve.GET("", reserveList)
			reserve.POST("/add", addReserve)
			reserve.POST("/import", importReserve)
			reserve.POST("/score/update", reserveScoreUpdate)
			reserve.POST("/notify/update", reserveNotifyUpdate)
			reserve.POST("/notify/delete", reserveNotifyDelete)
			reserve.GET("/counter/group", reserveCounterGroupList)
			reserve.POST("/counter/group", reserveCounterGroupUpdate)
			reserve.GET("/counter/node", reserveCounterNodeList)
		}
		upReserve := g.Group("/up/reserve", authSrv.Permit2("UP_RESERVE"))
		{
			upReserve.GET("/list", upReserveList)
			upReserve.POST("/hang", upReserveHang)
			upReserve.GET("/hang/log/list", upReserveHangLogList)
		}
		gappLikes := g.Group("/likes")
		{
			gappLikes.GET("/list", likesList)
			gappLikes.GET("/lids", likes)
			gappLikes.POST("/add", addLike)
			gappLikes.POST("/up", upLike)
			gappLikes.POST("/up/reply", upListContent)
			gappLikes.POST("/up/wid", upWid)
			gappLikes.POST("/add/pic", addPic)
			gappLikes.POST("/batch/wid", batchLikes)
			gappLikes.GET("/export", likeExport)
			gappLikes.POST("/batch/edit", likeBatchEdit)
		}
		gappKfc := g.Group("kfc")
		{
			gappKfc.GET("/list", kfcList)
		}
		// 时尚活动后台管理
		groupVogue := g.Group("/vogue")
		{
			// 商品管理
			groupVogue.GET("/goods/list", authSrv.Permit2("ACT_VOGUE_GOODS"), goodsList)
			groupVogue.GET("/goods/list/csv", authSrv.Permit2("ACT_VOGUE_GOODS"), goodsExport)
			groupVogue.POST("/goods/add", authSrv.Permit2("ACT_VOGUE_GOODS"), goodsAdd)
			groupVogue.POST("/goods/del", authSrv.Permit2("ACT_VOGUE_GOODS"), goodsDel)
			groupVogue.POST("/goods/modify", authSrv.Permit2("ACT_VOGUE_GOODS"), goodsModify)
			groupVogue.POST("/goods/soldout", authSrv.Permit2("ACT_VOGUE_GOODS"), goodsSoldOut)
			// 配置管理
			groupVogue.GET("/config/list", authSrv.Permit2("ACT_VOGUE_SETTING"), configList)
			groupVogue.POST("/config/edit", authSrv.Permit2("ACT_VOGUE_SETTING"), modifyConfig)
			// 积分进度
			groupVogue.GET("/credit", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditList)
			groupVogue.GET("/credit/csv", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditListExport)
			groupVogue.GET("/credit/csv/async", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditListGenerate)
			groupVogue.GET("/credit/csv/async/data", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditListAvailable)
			groupVogue.GET("/credit/csv/async/download", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditListDownload)
			groupVogue.GET("/detail", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditDetail)
			groupVogue.GET("/detail/csv", authSrv.Permit2("ACT_VOGUE_PROGRESS"), creditDetailExport)
			groupVogue.GET("/prize", authSrv.Permit2("ACT_VOGUE_PROGRESS"), prizeList)
			groupVogue.GET("/prize/csv", authSrv.Permit2("ACT_VOGUE_PROGRESS"), prizeExport)
			groupVogue.GET("/winning/list", authSrv.Permit2("ACT_VOGUE_PROGRESS"), winningList)
			groupVogue.GET("/winning/csv", authSrv.Permit2("ACT_VOGUE_PROGRESS"), winningListExport)
			// 微信封禁
			groupVogue.GET("/wechatcheck", authSrv.Permit2("ACT_VOGUE_SETTING"), weChatBlockStatus)
		}
		groupPre := g.Group("/prediction")
		{
			groupPre.POST("/add", predictionAdd)
			groupPre.GET("/search", predSearch)
			groupPre.POST("/up", predUp)
			groupPre.POST("/item/add", itemAdd)
			groupPre.POST("/item/up", itemUp)
			groupPre.GET("/item/search", itemSearch)
		}
		quesGroup := g.Group("/question")
		{
			quesGroup.GET("/base/list", baseList)
			quesGroup.GET("/base/item", baseItem)
			quesGroup.POST("/base/add", baseAdd)
			quesGroup.POST("/base/save", baseSave)
			quesGroup.GET("/detail/list", detailList)
			quesGroup.POST("/detail/add", detailAdd)
			quesGroup.POST("/detail/save", detailSave)
			quesGroup.POST("/detail/import/csv", importDetailCSV)
			quesGroup.POST("/detail/del", detailDel)
			quesGroup.POST("/detail/online", detailOnline)
		}
		gappBws := g.Group("/bws")
		{
			gappBws.POST("/add", addBws)
			gappBws.POST("/save", saveBws)
			gappBws.GET("/info", bwsInfo)
			gappBws.GET("/list", bwsList)
			gappBws.GET("/task/list", bwsTasks)
			gappBws.POST("/task/add", bwsTaskAdd)
			gappBws.POST("/task/edit", bwsTaskEdit)
			gappBws.POST("/task/del", bwsTaskDel)
			gappBws.GET("/award/list", bwsAwards)
			gappBws.POST("/award/add", bwsAwardAdd)
			gappBws.POST("/award/edit", bwsAwardEdit)
			gappBws.POST("/award/del", bwsAwardDel)
			gappBws.POST("/import/csv", bwsUsersImport)
			gappBws.POST("/import/vip_user", bwsUsersVipImport)

			gappAchievement := gappBws.Group("/achievement")
			{
				gappAchievement.POST("/add", addBwsAchievement)
				gappAchievement.POST("/save", saveBwsAchievement)
				gappAchievement.GET("/info", bwsAchievement)
				gappAchievement.GET("/list", bwsAchievements)
			}
			gappField := gappBws.Group("/field")
			{
				gappField.POST("/add", addBwsField)
				gappField.POST("/save", saveBwsField)
				gappField.GET("/info", bwsField)
				gappField.GET("/list", bwsFields)
			}
			gappPoint := gappBws.Group("/point")
			{
				gappPoint.POST("/add", addBwsPoint)
				gappPoint.POST("/save", saveBwsPoint)
				gappPoint.GET("/info", bwsPoint)
				gappPoint.GET("/list", bwsPoints)
				pointLevel := gappPoint.Group("/level")
				{
					pointLevel.POST("/save", saveBwsPointLevel)
				}
			}
			gappUser := gappBws.Group("/user")
			{
				gappUser.POST("/add", addBwsUser)
				gappUser.POST("/save", saveBwsUser)
				gappUser.GET("/info", bwsUser)
				gappUser.GET("/list", bwsUsers)
				gappUserAchievement := gappUser.Group("/achievement")
				{
					gappUserAchievement.POST("/add", addBwsUserAchievement)
					gappUserAchievement.POST("/save", saveBwsUserAchievement)
					gappUserAchievement.GET("/info", bwsUserAchievement)
					gappUserAchievement.GET("/list", bwsUserAchievements)
				}
				gappUserPoint := gappUser.Group("/point")
				{
					gappUserPoint.POST("/add", addBwsUserPoint)
					gappUserPoint.POST("/save", saveBwsUserPoint)
					gappUserPoint.GET("/info", bwsUserPoint)
					gappUserPoint.GET("/list", bwsUserPoints)
				}
			}
			bluetooth := gappBws.Group("/bluetooth")
			{
				bluetooth.GET("/up/list", bwsBluetoothUpList)
				bluetooth.POST("/up/add", bwsBluetoothUpAdd)
				bluetooth.POST("/up/svee", bwsBluetoothUpSave)
				bluetooth.POST("/up/del", bwsBluetoothUpDel)
			}
		}
		groupTask := g.Group("/task")
		{
			groupTask.GET("/list", taskList)
			groupTask.POST("/add", addTask)
			groupTask.POST("/addv2", addTaskV2)
			groupTask.POST("/save", saveTask)
			groupTask.POST("/award", addAward)
		}

		groupNewTask := g.Group("/new_task")
		{
			groupNewTask.GET("/list", taskv2List)
			groupNewTask.POST("/save", saveTaskV2)
		}
		groupSpringFestival := g.Group("/springfestival")
		{
			groupSpringFestival.GET("/user", sp2021User)
			groupSpringFestival.GET("/invite_log", sp2021InviteLog)
			groupSpringFestival.GET("/cards_nums_log", sp2021CardsNumsLog)
			groupSpringFestival.GET("/compose_log", sp2021ComposeLog)
			groupSpringFestival.GET("/send_cards_log", sp2021SendCardLog)
			groupSpringFestival.GET("/is_follow", isReserveSpringfestival)

		}
		groupCards := g.Group("/cards")
		{
			groupCards.POST("/add", addCards)
			groupCards.POST("/edit", editCards)
			groupCards.GET("/get_by_lottery", getCardsByLotteryID)
			groupCards.GET("/user", cardsUser)
			groupCards.GET("/invite_log", cardsInviteLog)
			groupCards.GET("/cards_nums_log", cardsNumsLog)
			groupCards.GET("/compose_log", cardsComposeLog)
			groupCards.GET("/send_cards_log", cardsSendCardLog)
			groupCards.GET("/is_follow", isReserveCards)
			groupCards.GET("/compose_count", cardsComposeCount)

		}
		groupCurr := g.Group("/currency")
		{
			groupCurr.GET("/list", currencyList)
			groupCurr.GET("/one", currencyItem)
			groupCurr.POST("/add", addCurrency)
			groupCurr.POST("/save", saveCurrency)
			groupCurr.POST("/relation/add", addCurrRelation)
			groupCurr.POST("/relation/del", delCurrRelation)
		}
		bnjGroup := g.Group("/bnj")
		{
			bnjGroup.POST("/2021/ar", bnj2021ARSetting)
			bnjGroup.POST("/2021/score2coupon/rule", bnj2021ARScore2CouponRule)
			bnjGroup.DELETE("/2021/score2coupon/rule", bnj2021ARScore2CouponRuleDel)
		}
		g.GET("/bnj/pendant/check", pendantCheck)
		g.POST("/bnj20/hotpot/value/change", bnjValueChange)
		g.GET("/bnj20/hotpot/value", bnjValue)
		groupLottery := g.Group("/lottery")
		{
			groupLottery.GET("/list", list)
			groupLottery.GET("/draft/list", listDraft)

			groupLottery.POST("/draft/add", authSrv.Verify2(), addDraft)
			groupLottery.GET("/detail", detail)
			groupLottery.GET("/draft/detail", authSrv.Verify2(), detailDraft)
			groupLottery.POST("/draft/edit", authSrv.Verify2(), editDraft)
			groupLottery.POST("/draft/gift/add", authSrv.Verify2(), giftAddDraft)
			groupLottery.POST("/draft/gift/edit", authSrv.Verify2(), giftEditDraft)
			// groupLottery.POST("/add", authSrv.Verify2(), add)
			groupLottery.POST("/delete", authSrv.Verify2(), deleteLottery)
			// groupLottery.POST("/edit", authSrv.Verify2(), edit)
			// groupLottery.POST("/gift/add", authSrv.Verify2(), giftAdd)
			// groupLottery.POST("/gift/edit", authSrv.Verify2(), giftEdit)
			groupLottery.GET("/gift/list", giftList)
			groupLottery.GET("/draft/gift/list", giftListDraft)
			groupLottery.POST("/draft/audit", authSrv.Verify2(), lotteryDraftAudit)
			groupLottery.GET("/gift/win", giftWinList)
			groupLottery.POST("/gift/upload", authSrv.Verify2(), giftUpload)
			groupLottery.GET("/gift/win/download", authSrv.Verify2(), giftExport)
			groupLottery.POST("/draft/membergroup/edit", authSrv.Verify2(), memberGroupDraftEdit)
			// groupLottery.POST("/membergroup/edit", authSrv.Verify2(), memberGroupEdit)
			groupLottery.GET("/membergroup/list", memberGroupList)
			groupLottery.GET("/draft/membergroup/list", memberGroupListDraft)
			groupLottery.GET("/vip/check", vipCheck)
			// groupLottery.POST("/batch/add/times", batchAddTimes)
			groupLottery.GET("/gift/win/download/all", authSrv.Verify2(), giftExportAll)
			groupLottery.GET("/wx/log", wxLotteryLog)
			groupLottery.GET("/used", unsedLottery)
			groupLottery.POST("/addtimes", authSrv.Verify2(), addTimesBatch)
			groupLottery.POST("/retry", authSrv.Verify2(), addTimesBatchRetry)
			groupLottery.GET("/times", lotteryMidAddTimes)
			groupLottery.GET("/addtimes_log", addTimesLogList)
			groupLottery.GET("/addtimes/mid", addTimesMidList)

		}
		groupUp := g.Group("/up")
		{
			groupUp.GET("/act/list", upActlist)
			groupUp.POST("/act/edit", upActEdit)
			groupUp.POST("/act/offline", upActOffline)
		}
		groupAward := g.Group("/award")
		{
			groupAward.GET("/detail", awardDetail)
			groupAward.GET("/list", awardSubList)
			groupAward.POST("/add", authSrv.Permit2(""), awardAdd)
			groupAward.POST("/save", authSrv.Permit2(""), awardSave)
			groupAward.GET("/log/list", awardSubLog)
			groupAward.GET("/log/export", awardSubLogExport)
		}
		exportTask := g.Group("/export/task")
		{
			exportTask.POST("/add", exportTaskAdd)
			exportTask.GET("/add", exportTaskAdd)
			exportTask.GET("/state", exportTaskState)
			exportTask.GET("/redo", exportTaskRedo)
			exportTask.GET("/list", exportTaskList)
			exportTask.GET("/wechat/userid", exportTaskWeChatUserID)
			exportTask.GET("/wechat/update", exportTaskWeChatUpdateMemberInfo)
		}
		ticket := g.Group("/ticket")
		{
			ticket.GET("/create", ticketCreate)
			ticket.GET("/export", ticketExport)
		}
		groupCollege := g.Group("/college")
		{
			groupCollege.GET("/list", collegeList)
			groupCollege.POST("/save", saveCollege)
			groupCollege.POST("/import/csv", collegeImport)
			groupCollege.GET("/aid/list", collegeAIDList)
			groupCollege.POST("/aid/save", collegeSaveAID)
		}
		groupTag := g.Group("tag")
		{
			groupTag.GET("/status", authSrv.Verify2(), tagStatus)
			groupTag.POST("/toactivity", authSrv.Verify2(), tagToActivity)
			groupTag.POST("/tonormal", authSrv.Verify2(), tagToNormal)
		}
		groupWhite := g.Group("/white_list")
		{
			groupWhite.POST("/add", authSrv.Permit2(""), addWhiteList)
			groupWhite.POST("/add/outer", verifySvc.Verify, addWhiteListOuter)
			groupWhite.POST("/delete", authSrv.Permit2(""), deleteWhiteList)
			groupWhite.GET("/list", authSrv.Permit2(""), whiteList)
		}
		groupRank := g.Group("/rank")
		{
			groupRank.POST("/add", authSrv.Verify2(), rankCreate)
			groupRank.GET("/detail", rankDetail)
			groupRank.POST("/update", authSrv.Verify2(), rankUpdate)
			groupRank.POST("/offline", rankOffline)
			groupRank.GET("/intervention", rankIntervention)
			groupRank.POST("/update_intervention", updateIntervention)
			groupRank.GET("/result", rankResult)
			groupRank.POST("/result_update", rankResultUpdate)
			groupRank.POST("/publish", rankPublish)
			groupRank.GET("/export", authSrv.Verify2(), rankExport)
		}
		groupRankNew := g.Group("/rank_v2")
		{
			groupRankNew.POST("/add", authSrv.Verify2(), rankV2Create)
			groupRankNew.POST("/update", authSrv.Verify2(), rankV2Update)
			groupRankNew.GET("/list", rankV2List)
			groupRankNew.GET("/base", rankV2Rank)
			groupRankNew.POST("/update_rule", authSrv.Verify2(), rankV2UpdateRule)
			groupRankNew.POST("/black_white", authSrv.Verify2(), rankV2BlackWhite)
			groupRankNew.POST("/adjust", authSrv.Verify2(), rankV2UpdateAdjust)
			groupRankNew.GET("/detail", rankV2Detail)
			groupRankNew.POST("/publish", authSrv.Verify2(), rankV2Publish)
			groupRankNew.POST("/rank_offline", authSrv.Verify2(), rankV2RuleOffline)
			groupRankNew.POST("/export", authSrv.Verify2(), rankV2Export)
			groupRankNew.GET("/source", authSrv.Verify2(), rankV2Source)
			groupRankNew.POST("/upload_source", authSrv.Verify2(), uploadSource)
			groupRankNew.POST("/update_ruleshow", authSrv.Verify2(), UpdateRulesShowInfo)
			groupRankNew.GET("/export_result", authSrv.Verify2(), rankV2ExportResult)

		}
		dataManage := g.Group("/data/manage")
		{
			dataManage.GET("/export", dataManageExport)
			dataManage.GET("/select", dataManageSelect)
			dataManage.POST("/update", dataManageUpdate)
			dataManage.POST("/diff", dataManageDiff)
		}
		systemManage := g.Group("/system")
		{
			systemManage.POST("/import/vip/list", importSignVipList)     // 导入签到vipList
			systemManage.GET("/sign/list", signList)                     // 获取活动签到人数列表
			systemManage.GET("/sign/vip/list", signVipList)              // 获取vip人员签到状态
			systemManage.POST("/sign/user", signUser)                    // 手动签到
			systemManage.GET("/export/sign/list", exportSignList)        // 导出签到列表数据
			systemManage.GET("/export/sign/vip/list", exportSignVipList) // 导出vip签到列表数据

			systemManage.POST("/act/add", actAdd)     // 新增活动
			systemManage.POST("/act/edit", actEdit)   // 新增活动
			systemManage.POST("/act/state", actState) // 下线or删除活动
			systemManage.GET("/act/info", actInfo)    // 活动信息
			systemManage.GET("/act/list", actList)    // 活动列表

			systemManage.GET("/seat/list", seatList) // 座位表页面

			systemManage.GET("/vote/sum", voteSum)                    // 获取各个选项已投票人数
			systemManage.GET("/vote/option", voteOption)              // 查看投票明细
			systemManage.GET("/vote/detail/export", voteDetailExport) // 导出投票明细

			systemManage.GET("/question/list", questionList)              // 提问列表
			systemManage.POST("/question/state", questionState)           // 删除问题
			systemManage.GET("/export/question/list", exportQuestionList) // 导出问答列表
		}
		groupPage := g.Group("/page")
		{
			groupPage.GET("/list", pageList)
		}
		knowGroup := g.Group("/knowledge")
		{
			knowGroup.POST("/history/update", historyUpdate)
		}

		fitActivityManage := g.Group("/fit")
		{
			fitActivityManage.POST("/plan/add_one", addOnePlan)
			fitActivityManage.POST("/plan/update_by_id", updatePlanById)
		}
		addInternalRewardsRouter(g)
		addInternalVoteRouter(g)
		addAccountBindRouter(g)
		addMissionActivityRouter(g)
		addRewardConfRouter(g)
	}
}

func ping(c *bm.Context) {
	if err := actSrv.Ping(c); err != nil {
		c.Error = err
		c.AbortWithStatus(503)
	}
}

func bvArgCheck(aid int64, bv string) (res int64, err error) {
	res = aid
	if bv != "" {
		if res, err = bvid.BvToAv(bv); err != nil {
			log.Error("bvid.BvToAv(%s) aid(%d) error(%+v)", bv, aid, err)
			err = ecode.RequestErr
			return
		}
	}
	if res <= 0 {
		err = ecode.RequestErr
	}
	return
}
