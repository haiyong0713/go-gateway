package http

import (
	"net/http"

	"go-gateway/app/web-svr/activity/interface/service"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/antispam"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/middleware"
	"go-gateway/app/web-svr/activity/interface/service/poll"
)

var (
	pollSvc   *poll.Poll
	authSvc   *auth.Auth
	vfySvc    *verify.Verify
	antispamM *antispam.Antispam
)

// Init int http service
func Init(c *conf.Config) {
	authSvc = auth.New(c.Auth)
	vfySvc = verify.New(c.Verify)
	pollSvc = poll.New(c)
	antispamM = antispam.New(c.Antispam)
	engine := bm.NewServer(c.HTTPServer)
	limiter := quota.New(c.Limiter)
	engine.Use(bm.Recovery(), bm.Trace(), bm.Logger(), bm.Mobile(), bm.NewRateLimiter(nil).Limit(), limiter.Handler())
	outerRouter(engine)
	internalRouter(engine)
	// init Outer serve
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Use(middleware.ThirdPartOriginHandler())
	e.Use(bm.CORS())
	e.Ping(ping)
	e.Register(register)
	e.POST("/x/activity/invite/bind", inviteBind)
	e.GET("/x/activity/invite/inviter", inviteGetInviter)
	group := e.Group("/x/activity", bm.CSRF())
	{
		group.GET("/subject", subject)
		group.POST("/archive/list", getArchiveByBvid)
		group.GET("/subject/info", subjectInfo)
		group.GET("/protocols", subjectProtos)
		group.POST("/vote", authSvc.User, vote)
		group.GET("/ltime", ltime)
		group.GET("/object/group", groupData)
		group.GET("/tag/object", tagList)
		group.GET("/tag/object/stats", tagStats)
		group.GET("/region/object", regionList)
		group.GET("/subject/total/stat", subjectStat)
		group.GET("/view/rank", viewRank)
		group.POST("/likeact", authSvc.User, likeAct)
		group.POST("/likeact/batch", authSvc.User, batchLikeAct)
		group.GET("/likeact/likes", authSvc.User, likeActLikes)
		group.GET("/likeact/list", authSvc.Guest, likeActList)
		group.POST("/likeact/cv", authSvc.User, likeActBySidCVId)
		group.POST("/like/del", authSvc.User, likeDel)
		group.GET("/like/one", authSvc.Guest, likeOne)
		group.GET("/like/slider", authSvc.Guest, likeSlider)
		group.GET("/like/mylist", authSvc.User, likeMyList)
		group.GET("/like/check", authSvc.User, likeCheckJoin)
		group.POST("/missiongroup/like", authSvc.User, missionLike)
		group.POST("/missiongroup/likeact", authSvc.User, missionLikeAct)
		group.GET("/missiongroup/info", authSvc.User, missionInfo)
		group.GET("/missiongroup/tops", missionTops)
		group.GET("/missiongroup/user", missionUser)
		group.GET("/missiongroup/rank", authSvc.User, missionRank)
		group.GET("/missiongroup/friends", authSvc.User, missionFriends)
		group.GET("/missiongroup/award", authSvc.User, missionAward)
		group.POST("/missiongroup/achievement", authSvc.User, missionAchieve)
		group.POST("/up/act", authSvc.User, storyKingAct)
		group.POST("/up/addtimes", authSvc.User, upAddTimes)
		group.GET("/up/left", authSvc.User, storyKingLeft)
		group.GET("/up/list", authSvc.Guest, upList)
		group.GET("/likeact/up/list", authSvc.Guest, likeactUpList)
		group.GET("/up/list/relation", authSvc.Guest, upListRelation)
		group.GET("/up/list/group", authSvc.Guest, upListGroup)
		group.GET("/up/list/his", upListHis)
		group.GET("/up/songfestival/process", authSvc.Guest, festivalProcess)
		group.GET("/prediction", prediction)
		group.GET("/article/giant", authSvc.User, articleGiant)
		group.GET("/up/special", authSvc.User, upSpecial)
		group.POST("/likeact/token", authSvc.User, likeActToken)
		group.GET("/invite/times", authSvc.Guest, inviteTimes)
		group.GET("/native/rcmdInfo", authSvc.Guest, rcmdInfo)
		group.GET("/actdomain/list", listDomain)
		group.GET("/actdomain/search", searchDomain)
		group.POST("/reserve", authSvc.User, reserve)
		group.POST("/reserve/cancel", authSvc.User, reserveCancel)
		group.GET("/reserve/following", authSvc.Guest, reserveFollowing)
		group.GET("/reserve/progress", authSvc.Guest, reserveProgress)
		group.GET("/reserve/group/progress", authSvc.Guest, reserveGroupProgress)
		group.POST("/reserve/send/point", authSvc.User, reserveSendPoint)
		group.POST("/reserve/dove/award", authSvc.User, reserveDoveAward)
		group.GET("/resource/res/audit", resourceAudit)
		group.GET("/iir", resourceIir)
		group.GET("/subject/award/state", authSvc.User, awardSubjectState)
		group.POST("/subject/award/reward", authSvc.User, rewardSubject)
		group.GET("/subject/award/user/state", authSvc.User, awardSubjectStateByID)
		group.POST("/subject/award/user/reward", authSvc.User, rewardSubjectByID)
		group.POST("article/day2/join", authSvc.User, articleDayJoin)
		group.GET("article/day2/info", authSvc.User, articleDayInfo)
		group.GET("/preheat/downinfo", downloadInfo)
		group.POST("/ticket/sign", authSvc.User, ticketSign)
		group.GET("/yellow/green/vote", yellowGreenVote)
		group.POST("/filter", actFilter)
		group.GET("/web/view/data", viewData)
		subGroup := group.Group("/likes")
		{
			subGroup.POST("/add/questionnaire", authSvc.User, likeAddText)
			subGroup.POST("/add/text", authSvc.User, likeAddText)
			subGroup.POST("/add/other", authSvc.User, likeAddOther)
			subGroup.GET("/total", likeTotal)
		}
		spGroup := group.Group("/sports")
		{
			spGroup.GET("/qq", qq)
			spGroup.GET("/news", news)
		}
		knowGroup := group.Group("/knowledge/badge")
		{
			knowGroup.GET("/progress", authSvc.Guest, badgeProgress)
			knowGroup.POST("/share", authSvc.User, badgeShare)
			knowGroup.GET("/user", authSvc.Guest, knowledgeUser)
		}
		matchGroup := group.Group("/match")
		{
			matchGroup.GET("", matchs)
			matchGroup.GET("/unstart", authSvc.Guest, unStart)
			matchGroup.POST("/suits", authSvc.User, addSuits)
			guGroup := matchGroup.Group("/guess")
			{
				guGroup.GET("", authSvc.User, guess)
				guGroup.GET("/list", authSvc.User, listGuess)
				guGroup.POST("/add", authSvc.User, addGuess)
			}
			foGroup := matchGroup.Group("/follow")
			{
				foGroup.GET("", authSvc.User, follow)
				foGroup.POST("/add", authSvc.User, addFollow)
			}
			s10Group := matchGroup.Group("/s10")
			{
				s10Group.POST("/sign", authSvc.User, signed)
				s10Group.GET("/tasks", authSvc.User, tasksProgress)
				s10Group.GET("/total", authSvc.User, totalPoints)
				s10Group.GET("/stage", authSvc.Guest, matchesStage)
				s10Group.GET("/lottery", authSvc.User, userMatchesLotteryInfo)
				s10Group.POST("/stage/lottery/state", authSvc.User, updateUserLooteryState)
				s10Group.POST("/stage/lottery", authSvc.User, stageLottery)
				s10Group.GET("/goods", authSvc.Guest, actGoods)
				s10Group.POST("/goods/exchange", authSvc.User, exchangeGoods)
				s10Group.POST("/user/flow/check", checkUserProfile)
				s10Group.POST("/user/flow/ack", ackSendFreeFlow)
				s10Group.GET("/user/flow", authSvc.Guest, userFlow)
				s10Group.GET("/other/activity", otherActivity)
			}
		}
		tmGroup := group.Group("/timemachine")
		{
			tmGroup.GET("/2018", authSvc.User, timemachine2018)
			tmGroup.GET("/2019", authSvc.User, timemachine2019)
			tmGroup.GET("/2019/raw", authSvc.User, tmRaw2019)
			tmGroup.GET("/2019/cache", authSvc.User, tmCache2019)
			tmGroup.GET("/2019/reset", authSvc.User, tmReset2019)
			tmGroup.GET("/2020/user/report", authSvc.User, userReport2020)
			tmGroup.GET("/2020/before/publish", authSvc.User, beforePublish2020)
			tmGroup.POST("/2020/publish", authSvc.User, publish2020)
			tmGroup.GET("/2020/user/report/cache", userReport2020Cache)
		}
		bwsGroup := group.Group("/bws")
		{
			bwsGroup.GET("/user", authSvc.Guest, user)
			bwsGroup.GET("/reserve", authSvc.User, bwsUserReserve)
			bwsGroup.POST("/reserve/check", authSvc.User, bwsCheckReserve)
			bwsGroup.GET("/points", points)
			bwsGroup.GET("/point", point)
			bwsGroup.GET("/achievements", achievements)
			bwsGroup.GET("/achievement", achievement)
			bwsGroup.POST("/point/unlock", authSvc.User, unlock)
			bwsGroup.POST("/binding", authSvc.User, binding)
			bwsGroup.POST("/award", authSvc.User, award)
			bwsGroup.GET("/lottery", authSvc.User, lottery)
			bwsGroup.GET("/lottery/v1", authSvc.User, lotteryV1)
			bwsGroup.GET("/lottery/check", authSvc.User, lotteryCheck)
			bwsGroup.GET("/redis/check", authSvc.User, redisInfo)
			bwsGroup.GET("/key/info", authSvc.User, keyInfo)
			bwsGroup.GET("/admin/check", authSvc.User, adminInfo)
			bwsGroup.GET("/recharge/award", authSvc.Guest, rechargeAward)
			bwsGroup.GET("/achieve/rank", authSvc.Guest, achieveRank)
			bwsGroup.GET("/fields", fields)
			bwsGroup.POST("/grade/enter", authSvc.User, gradeEnter)
			bwsGroup.GET("/grade/show", authSvc.Guest, gradeShow)
			bwsGroup.POST("/grade/fix", authSvc.User, gradeFix)
			bwsGroup.GET("/2020/user", authSvc.User, user2020)
			bwsGroup.GET("/admin/2020/user", authSvc.User, adminUser2020)
			bwsGroup.GET("/2020/award/list", authSvc.User, awardList)
			bwsGroup.POST("/2020/task/award", authSvc.User, taskAward)
			bwsGroup.POST("/2020/lottery", authSvc.User, lottery2020)
			bwsGroup.POST("/2020/award/send", authSvc.User, awardSend)
			bwsGroup.POST("/2020/point/unlock", authSvc.User, unlock2020)
			bwsGroup.POST("/2020/create/token", authSvc.User, createUserToken)
			bwsGroup.POST("/2020/offline/award", authSvc.User, offlineAwardSend)
			bwsGroup.GET("/2020/user/rank", offlineUserRank)
			bwsGroup.GET("/2020/member", authSvc.User, bws2020Member)
			bwsGroup.GET("admin/2020/member", authSvc.User, adminBws2020Member)
			bwsGroup.POST("/game/play", authSvc.User, bws2020PlayGame)
			bwsGroup.POST("/heart/add", authSvc.User, bws2020AddHeart)
			bwsGroup.GET("/2020/store", bws2020Store)
			bwsGroup.GET("/user/points", authSvc.User, bwsUserPoints)
			bwsGroup.GET("/admin/user/points", authSvc.User, bwsUserPointsAdmin)
			bwsGroup.POST("/2020/vip_add_heart", authSvc.User, bws2020VipAddHeart)

			bwsGroup.POST("/vote", authSvc.User, bwsVote)
			bwsGroup.POST("/vote/clear", authSvc.User, bwsVoteClear)
			buleGroup := bwsGroup.Group("/bluetooth", authSvc.Guest)
			{
				buleGroup.POST("/up/catch", catchUp)
				buleGroup.GET("/up/catch/list", catchList)
				buleGroup.GET("/up/catch/key", catchBluetooth)
				buleGroup.GET("/up/keys", bluetoothUps)
			}
		}
		bnjGroup := group.Group("/bnj2019")
		{
			bnjGroup.GET("/preview", authSvc.Guest, previewInfo)
			bnjGroup.GET("/timeline", timeline)
			bnjGroup.POST("/fail", fail)
			bnjGroup.POST("/reset", authSvc.User, reset)
			bnjGroup.POST("/reward", authSvc.User, reward)
		}
		bnj20Group := group.Group("/bnj2020")
		{
			bnj20Group.GET("/main", authSvc.Guest, bnj20Main)
			bnj20Group.POST("/reward", authSvc.User, bnj20Reward)
			bnj20Group.GET("/material", authSvc.Guest, bnj20Material)
			bnj20Group.POST("/material/unlock", authSvc.User, bnj20MaterialUnlock)
			bnj20Group.POST("/material/reddot/clear", authSvc.User, bnj20MaterialRedDot)
			bnj20Group.POST("/hotpot/increase", authSvc.User, bnj20HotpotIncrease)
			bnj20Group.POST("/hotpot/decrease", authSvc.User, bnj20HotpotDecrease)
		}
		kfcGroup := group.Group("/kfc")
		{
			kfcGroup.GET("/info", authSvc.User, kfcInfo)
			kfcGroup.GET("/use", kfcUse)
		}
		singleGroup := group.Group("/single")
		{
			singleGroup.POST("/grant/pid", authSvc.User, grantPid)
			singleGroup.POST("/image/lottery", authSvc.User, imageLottery)
			singleGroup.POST("/image/task", authSvc.User, doImageTask)
			singleGroup.GET("/fate", fateData)
			singleGroup.GET("/birthday", authSvc.Guest, allCurrency)
			singleGroup.POST("/receive/coupon", authSvc.User, receiveCoupon)
			singleGroup.GET("/star", authSvc.User, starState)
			singleGroup.GET("/star/arc", authSvc.User, starArc)
			singleGroup.GET("/star/spring", authSvc.User, starSpring)
			singleGroup.POST("/task/do", authSvc.User, doTask)
			singleGroup.GET("/miku/list", authSvc.Guest, mikuList)
			singleGroup.GET("/scholarship/list", authSvc.Guest, specialList)
			singleGroup.POST("/scholarship/award", authSvc.User, specialAward)
			singleGroup.GET("/stein/list", steinList)
			singleGroup.GET("/user/match", authSvc.User, userMatch)
			singleGroup.POST("/award/receive", authSvc.User, singleAward)
			singleGroup.GET("/award", authSvc.User, singleAwardState)
			singleGroup.GET("/certificate/wall", authSvc.Guest, certificateWall)
			singleGroup.GET("/star/more/arc", authSvc.User, starMoreArc)
			singleGroup.GET("/arc/list", authSvc.Guest, archiveList)
			singleGroup.GET("/arc/lists", arcLists)
			singleGroup.GET("/channel/arcs", channelArcs)
			singleGroup.POST("/point/lottery", authSvc.User, pointLottery)
			singleGroup.GET("/article/list", authSvc.User, upArtLists)
			singleGroup.POST("/article/list/add", authSvc.User, addArtList)
			singleGroup.GET("/web/data", singleWebData)
			singleGroup.GET("/group/web/data", authSvc.Guest, singleGroupWebData)
			singleGroup.POST("/task/token/do", authSvc.User, taskTokenDo)
			singleGroup.GET("/card/num", authSvc.Guest, cardNum)
			singleGroup.GET("/special/arc/list", specialArcList)
			singleGroup.GET("/read/day", authSvc.Guest, readDay)
			singleGroup.POST("/bml2020/follow", authSvc.User, bml20Follow)
			singleGroup.GET("/image/user/rank", authSvc.Guest, imageUserRank)
			singleGroup.GET("/childhood/list", childhoodList)
			singleGroup.GET("/giant/article/list", authSvc.User, giantArticleList)
			singleGroup.POST("/giant/article/choose", authSvc.User, giantArticleChoose)
			singleGroup.GET("/stupid/list", authSvc.Guest, stupidList)
			singleGroup.GET("/stupid/status", authSvc.Guest, stupidStatus)
		}
		bdfGroup := group.Group("/bdf")
		{
			bdfGroup.GET("/school/list", authSvc.Guest, schoolList)
			bdfGroup.GET("/school/arcs", authSvc.Guest, schoolArcs)
		}
		taskGroup := group.Group("/task")
		{
			taskGroup.GET("/list", authSvc.Guest, taskList)
			taskGroup.POST("/award", authSvc.User, awardTask)
			taskGroup.POST("/award/special", authSvc.User, awardTaskSpecial)
			taskGroup.GET("check", authSvc.User, taskCheck)
			taskGroup.POST("/send_points", authSvc.User, sendPoints)
			taskGroup.GET("/detail", authSvc.User, taskResult)
		}
		currGroup := group.Group("/currency")
		{
			currGroup.GET("/amount", authSvc.User, actCurrency)
		}
		questionGroup := group.Group("/question")
		{
			questionGroup.POST("/start", authSvc.User, questionStart)
			questionGroup.POST("/answer", authSvc.User, questionAnswer)
			questionGroup.POST("/next", authSvc.User, questionNext)
			questionGroup.GET("/myrecords", authSvc.User, questionMyRecords)
			questionGroup.GET("/qa", questionAndAnswer)
			questionGroup.GET("/gaokao/qa", gaokaoQuestion)
			questionGroup.POST("/gaokao/report", authSvc.User, gaokaoReport)
			questionGroup.GET("/gaokao/rank", authSvc.User, gaokaoRank)
		}
		dynamicGroup := group.Group("/dynamic")
		{
			dynamicGroup.GET("/index", authSvc.Guest, actIndex)
			dynamicGroup.POST("/liked", authSvc.User, actLiked)
			dynamicGroup.GET("/act", authSvc.Guest, actList)
			dynamicGroup.GET("/new/act", authSvc.Guest, newActList)
			dynamicGroup.GET("/video/act", authSvc.Guest, videoAct)
			// 以下接口服务已迁移 start
			dynamicGroup.GET("/topic", authSvc.Guest, actDynamic)
			dynamicGroup.GET("/pages", natPages)
			dynamicGroup.GET("/new/video/aid", authSvc.Guest, newVideoAid)
			dynamicGroup.GET("/new/video/dyn", authSvc.Guest, newVideoDyn)
			dynamicGroup.GET("/resource/aid", authSvc.Guest, resourceAid)
			dynamicGroup.GET("/resource/dyn", authSvc.Guest, resourceDyn)
			dynamicGroup.GET("/season/ssid", authSvc.Guest, seasonIDs)
			dynamicGroup.GET("/season/source", authSvc.Guest, seasonSource)
			dynamicGroup.GET("/resource/role", authSvc.Guest, resourceRole)
			dynamicGroup.GET("/timeline/source", timelineSource)
			dynamicGroup.GET("/module", natModule)
			dynamicGroup.GET("/live", authSvc.Guest, liveDyn)
			// 以下接口服务已迁移 end
		}
		lotteryGroup := group.Group("/lottery")
		{
			lotteryGroup.POST("/do", authSvc.User, doLottery)
			lotteryGroup.POST("/do/simple", authSvc.User, doSimpleLottery)
			lotteryGroup.POST("/addtimes", authSvc.User, addLotteryTimes)
			lotteryGroup.POST("/task/addtimes", authSvc.User, syncTask)
			lotteryGroup.GET("/task/info", authSvc.User, taskInfo)
			lotteryGroup.GET("/progressrate", progressRate)
			lotteryGroup.POST("/addtimes_yuandan", internalAddLotteryTimes)
			lotteryGroup.POST("/addtimes/extra", authSvc.User, addExtraTimes)
			lotteryGroup.GET("/mylist", authSvc.User, lotteryGetMyList)
			lotteryGroup.GET("/mytimes", authSvc.User, lotteryGetUnusedTimes)
			lotteryGroup.GET("/win/list", authSvc.Guest, lotteryWinList)
			lotteryGroup.POST("/gift/address/add", authSvc.User, addLotteryAddress)
			lotteryGroup.GET("/gift/address", authSvc.User, lotteryAddress)
			lotteryGroup.POST("/wx/do", authSvc.User, wxLotteryDo)
			lotteryGroup.GET("/wx/award", authSvc.Guest, wxLotteryAward)
			lotteryGroup.GET("/wx/gift", wxLotteryGifts)
			lotteryGroup.POST("/wx/play/window", authSvc.Guest, wxLotteryPlayWindow)
			lotteryGroup.GET("/my/win/list", authSvc.User, lotteryMyWinList)
			lotteryGroup.GET("/my/count", authSvc.User, lotteryMyCount)
			lotteryGroup.GET("/my/can_addtimes", authSvc.User, lotteryCanAddTimes)
			lotteryGroup.GET("/my/coupon/list", authSvc.User, lotteryCouponWinList)
			lotteryGroup.GET("/orderno", lotteryOrderNo)
			lotteryGroup.POST("/supplyment", supplymentWin)
			lotteryGroup.GET("/gift", lotteryGift)

		}
		if service.EsportSvc != nil {
			esportsGroup := group.Group("/esports_arena")
			{
				esportsGroup.GET("/user_info", authSvc.User, UserInfo)
				esportsGroup.POST("/fav", authSvc.User, addFavGame)
				esportsGroup.POST("/lotterys/add", authSvc.User, settleHistory)
			}
		}
		pollGroup := group.Group("/poll")
		{
			pollGroup.GET("/meta", pollMeta)
			pollGroup.GET("/options", pollOptions)
			pollGroup.GET("/option/stat/top", pollOptionStatTop)
			pollGroup.POST("/s9/vote", authSvc.User, pollS9Vote)
			pollGroup.GET("/voted", authSvc.User, pollVoted)

			pm := pollGroup.Group("/m", authSvc.User, pollM)
			pm.GET("/options", pollMOptions)
			pm.POST("/options/delete", pollMOptionsDelete)
			pm.POST("/options/add", pollMOptionsAdd)
			pm.POST("/options/update", pollMOptionsUpdate)
		}
		upGroup := group.Group("/up")
		{
			upGroup.GET("/launch/check", authSvc.User, upLaunchCheck)
			upGroup.GET("/check", upCheck)
			upGroup.POST("/launch", authSvc.User, upLaunch)
			upGroup.GET("/archive/list", authSvc.User, upArchiveList)
			upGroup.GET("/act/info", authSvc.Guest, upActPage)
			upGroup.GET("/act/rank", authSvc.Guest, upActRank)
			upGroup.POST("/do", authSvc.User, upDo)
		}
		// 手机厂商领取大会员时长
		mvGroup := group.Group("/appstore")
		{
			mvGroup.GET("/state", authSvc.UserMobile, appStoreState)
			mvGroup.POST("/receive", authSvc.UserMobile, antispamM.Handler(), appStoreReceive)
		}
		vogueGroup := group.Group("/vogue")
		{
			vogueGroup.GET("/state", authSvc.User, vogueState)
			vogueGroup.GET("/prizes", voguePrizes)
			vogueGroup.GET("/share", vogueShare)
			vogueGroup.POST("/select", authSvc.User, vogueSelectPrizes)
			vogueGroup.POST("/add", authSvc.User, vogueAddtimes)
			vogueGroup.POST("/exchange", authSvc.User, vogueExchange)
			vogueGroup.POST("/address", authSvc.User, vogueAddress)
			vogueGroup.GET("/list/invite", authSvc.User, vogueInviteList)
			vogueGroup.GET("/list/prize", voguePrizeList)
		}
		bmlGroup := group.Group("/bml/online")
		{
			bmlGroup.GET("/guess/myrecord", authSvc.User, bmlOnlineMyGuessList)
			bmlGroup.POST("/guess/do", authSvc.User, bmlOnlineGuessDo)
		}
		bwsOnGroup := group.Group("/bws/online")
		{
			bwsOnGroup.GET("/main", authSvc.User, bwsOnlineMain)
			bwsOnGroup.POST("/piece/find", authSvc.User, bwsOnlinePieceFind)
			bwsOnGroup.POST("/piece/find/free", authSvc.User, bwsOnlinePieceFindFree)
			bwsOnGroup.GET("/award/list", authSvc.Guest, bwsOnlineAwardList)
			bwsOnGroup.GET("/award/my/list", authSvc.User, bwsOnlineMyAwardList)
			bwsOnGroup.POST("/reward", authSvc.User, bwsOnlineReward)
			bwsOnGroup.POST("/currency/find", authSvc.User, bwsOnlineCurrencyFind)
			bwsOnGroup.POST("/ticket/reward", authSvc.User, bwsOnlineTicketReward)
			bwsOnGroup.GET("/my/dress/list", authSvc.User, bwsOnlineMyDress)
			bwsOnGroup.POST("/dress/up", authSvc.User, bwsOnlineDressUp)
			bwsOnGroup.GET("/print/list", authSvc.Guest, bwsOnlinePrintList)
			bwsOnGroup.GET("/print/detail", authSvc.Guest, bwsOnlinePrintDetail)
			bwsOnGroup.POST("/print/unlock", authSvc.User, bwsOnlinePrintUnlock)
			bwsOnGroup.POST("/park/ticket/bind", authSvc.User, bwsOnlineTicketBind)
			bwsOnGroup.GET("/park/reserve/info", authSvc.User, bwsOnlineReserveInfo)
			bwsOnGroup.POST("/park/reserve/do", authSvc.User, bwsOnlineReserveDo)
			bwsOnGroup.GET("/park/myreserve", authSvc.User, bwsOnlineReservedList)
		}
		brandGroup := group.Group("/brand")
		{
			brandGroup.POST("/coupon", authSvc.User, coupon)
		}
		handWriteGroup := group.Group("/handwrite")
		{
			handWriteGroup.GET("/member_count", authSvc.Guest, memberCount)
			handWriteGroup.GET("/rank", authSvc.Guest, rank)
			handWriteGroup.GET("/personal", authSvc.User, personal)
			handWriteGroup.POST("/addlotterytimes", authSvc.User, addHwLotteryTimes)
			handWriteGroup.GET("/coins", authSvc.User, coin)
		}
		handWrite2021Group := group.Group("/handwrite2021")
		{
			handWrite2021Group.GET("/member_count", authSvc.Guest, handwriteMemberCount)
			handWrite2021Group.GET("/personal", authSvc.User, handwritePersonal)
		}
		remixGroup := group.Group("/remix")
		{
			remixGroup.GET("/member_count", authSvc.Guest, remixMemberCount)
			remixGroup.GET("/rank", authSvc.Guest, remixRank)
			remixGroup.GET("/child_rank", authSvc.Guest, remixChildRank)
			remixGroup.GET("/personal", authSvc.User, remixPersonal)
		}
		gameHolidayGroup := group.Group("/gameholiday")
		{
			gameHolidayGroup.POST("/addlotterytimes", authSvc.User, addGhLotteryTimes)
			gameHolidayGroup.GET("/like", authSvc.User, ghLikes)
		}
		newStarGroup := group.Group("/newstar")
		{
			newStarGroup.POST("/join", authSvc.User, newstarJoin)
			newStarGroup.GET("/creation/list", authSvc.User, newstarCreation)
			newStarGroup.GET("/invite/list", authSvc.User, newstarInvite)
		}
		lolGroup := group.Group("/s10")
		{
			lolGroup.GET("/coin/wins", authSvc.User, coinWins)
			lolGroup.GET("/coin/predict_list", authSvc.User, coinPredictList)
			lolGroup.GET("/point/list", authSvc.User, pointList)
			lolGroup.GET("/guess/predict/list", authSvc.User, guessPredictList)
		}
		inviteGroup := group.Group("/invite")
		{
			inviteGroup.POST("/token", authSvc.User, inviteToken)
		}
		if service.CollegeSvc != nil {
			collegeGroup := group.Group("/college")
			{
				collegeGroup.POST("/bind", authSvc.User, collegeBind)
				collegeGroup.GET("/province", authSvc.Guest, collegeProvinceRank)
				collegeGroup.GET("/nationwide", authSvc.Guest, collegeNationwideRank)
				collegeGroup.GET("/list", authSvc.Guest, collegeList)
				collegeGroup.GET("/people", authSvc.Guest, collegePeopleRank)
				collegeGroup.GET("/personal", authSvc.User, collegePersonal)
				collegeGroup.GET("/detail", authSvc.Guest, collegeDetail)
				collegeGroup.GET("/archive", authSvc.Guest, collegeTabArchive)
				collegeGroup.GET("/aidisactivity", authSvc.Guest, collegeArchiveIsActivity)
				collegeGroup.GET("/upload_version", authSvc.Guest, collegeUploadVersion)
				collegeGroup.GET("/task", authSvc.User, collegeTask)
				collegeGroup.GET("/inviter", authSvc.User, collegeInviter)
				collegeGroup.POST("/follow", authSvc.User, collegeFollow)
			}
		}
		contriGroup := group.Group("/contribution")
		{
			contriGroup.GET("/archives/info", authSvc.Guest, archiveInfo)
			contriGroup.POST("/addlotterytimes", authSvc.User, addContriTimes)
			contriGroup.GET("/like", authSvc.User, contriLikes)
			contriGroup.GET("/light_bcut", authSvc.Guest, lightBcut)
			contriGroup.GET("/rank", totalRank)
		}
		dubbingGroup := group.Group("/dubbing")
		{
			dubbingGroup.GET("/personal", authSvc.User, dubbingPersonal)
			dubbingGroup.GET("/rank", dubbingRank)
		}
		funnyGroup := group.Group("/funny")
		{
			funnyGroup.GET("/page", pageInfo)
			funnyGroup.GET("/likes", authSvc.User, getLikeCount)
			funnyGroup.POST("/add", authSvc.User, incrDrawTime)
		}
		if service.AcgSvc != nil {
			acgGroup := group.Group("/acg")
			{
				acgGroup.GET("/2020/task", authSvc.Guest, acg2020Task)
			}
		}
		s10AnswerGroup := group.Group("/answer")
		{
			s10AnswerGroup.GET("/user/info", authSvc.User, answerUserInfo)
			s10AnswerGroup.GET("/question", authSvc.User, answerQuestion)
			s10AnswerGroup.POST("/result", authSvc.User, answerResult)
			s10AnswerGroup.GET("/rank", answerRank)
			s10AnswerGroup.POST("/pendant/add", authSvc.User, answerPendant)
			s10AnswerGroup.POST("/know/rule", authSvc.User, knowRule)
			s10AnswerGroup.POST("/share/add", authSvc.User, shareAddHP)
			s10AnswerGroup.GET("/week/top", weekTop)
		}
		rankGroup := group.Group("/rank")
		{
			rankGroup.GET("/result", rankResult)
			rankGroup.GET("/personal", authSvc.User, rankPersonal)
		}
		selectionGroup := group.Group("/selection")
		{
			selectionGroup.GET("/user/info", authSvc.User, selectionInfo)
			selectionGroup.POST("/sensitive", sensitive)
			selectionGroup.POST("/submit", authSvc.User, selectionSubmit)
			selectionGroup.GET("/list", authSvc.Guest, seleList)
			selectionGroup.POST("/vote", authSvc.User, selectionVote)
			selectionGroup.GET("/rank", authSvc.User, selectionRank)
			selectionGroup.GET("/assistance", seleAssistance)
		}
		relationGroup := group.Group("/relation")
		{
			relationGroup.GET("/info", authSvc.Guest, getRelationInfo)
			//relationGroup.GET("/get_reserve_info", authSvc.Guest, getRelationReserveInfo)
			relationGroup.POST("/do", authSvc.User, doRelation)
		}
		if service.SystemSvc != nil {
			commonGroup := group.Group("/system")
			{
				// 权限验证
				commonGroup.GET("/oauth", WXAuth)
				// JSSDK初始化配置下发
				commonGroup.GET("/config", getConfig)
				// 获取活动基本信息
				commonGroup.GET("/activity", activityInfo)
				// 签到
				commonGroup.POST("/sign", sign)
				// 投票
				commonGroup.POST("/vote", systemVote)
				// 提问
				commonGroup.POST("/question", systemQuestion)
				// 列表
				commonGroup.GET("/question/list", systemQuestionList)
			}
		}
		newyear2021Group := group.Group("/happy2021")
		{
			newyear2021Group.GET("game/task", authSvc.Guest, Bnj2021TaskStatus)
			newyear2021Group.POST("mall/visit", authSvc.User, Bnj2021PubMallVisit)
			newyear2021Group.POST("reward", authSvc.User, Bnj2021ReceiveReward)
			newyear2021Group.GET("/game/prepare", authSvc.User, Bnj2021PreExchange)
			newyear2021Group.GET("/game/confirm", authSvc.User, Bnj2021ARConfirm)
			newyear2021Group.GET("/config", authSvc.User, BnjARConfig)
			newyear2021Group.GET("/device/adapt_level", ARAdaptLevel)
			newyear2021Group.POST("/game", authSvc.User, Bnj2021Exchange)
			newyear2021Group.GET("/profile", authSvc.User, Bnj2021Profile)
			newyear2021Group.GET("/game/quota", authSvc.User, Bnj2021ARQuota)
			newyear2021Group.GET("/draw", authSvc.Guest, Bnj2021LiveLotteryDetail)
			newyear2021Group.GET("/draw/detail", authSvc.Guest, Bnj2021LotteryUserHistory)
			newyear2021Group.POST("/draw/pay", authSvc.User, Bnj2021RewardPay)
			newyear2021Group.POST("/publicize/draw", authSvc.User, Bnj2021PublicizeRewardPay)
			newyear2021Group.GET("/publicize", authSvc.Guest, Bnj2021PublicizeBiz)
			newyear2021Group.GET("/support", authSvc.User, Bnj2021Support)
			newyear2021Group.GET("/live/status", BnjLiveStatus)
			newyear2021Group.GET("/exam", authSvc.Guest, BnjLiveExamDetail)
			newyear2021Group.POST("/exam/commit", authSvc.User, BnjLiveExamCommit)
			newyear2021Group.GET("/reserve", authSvc.Guest, BnjReserveStatus)
		}
		upReserve := group.Group("/up/reserve")
		{
			upReserve.POST("/create", authSvc.User, CreateUpActReserve)
			upReserve.POST("/update", authSvc.User, UpdateUpActReserve)
			upReserve.GET("/info", authSvc.User, UpActReserveInfo)
			upReserve.POST("/cancel", authSvc.User, CreateUpActCancel)
			upReserve.GET("/available", authSvc.User, UpActReserveRelationContinuing)
			upReserve.GET("/attachment", authSvc.User, UpActReserveRelationOthers)
			upReserve.GET("/relation/info", authSvc.Guest, UpActReserveRelationInfo)
		}
		if service.SpringFestival2021Svc != nil {
			springfestial2021Group := group.Group("/spring2021")
			{
				springfestial2021Group.POST("/join", authSvc.User, springFestivalJoin)
				springfestial2021Group.POST("/draw", authSvc.User, springFestivalIsJoin, springFestivalDraw)
				springfestial2021Group.POST("/addtimes", authSvc.User, springFestivalIsJoin, springFestivalAddTimes)
				springfestial2021Group.GET("/times", authSvc.User, springFestivalIsJoin, springFestivalTimes)
				springfestial2021Group.POST("/follow", authSvc.User, springFestivalIsJoin, springFestivalFollow)
				springfestial2021Group.GET("/follower", authSvc.User, springFestivalFollower)
				springfestial2021Group.GET("/task", authSvc.User, springFestivalIsJoin, springFestivalTask)
				springfestial2021Group.GET("/cards", authSvc.User, springFestivalIsJoin, springFestivalCards)
				springfestial2021Group.GET("/invite_share", authSvc.User, springFestivalIsJoin, springFestivalInviteShare)
				springfestial2021Group.POST("/bind", authSvc.User, springFestivalBind)
				springfestial2021Group.POST("/compose", authSvc.User, springFestivalIsJoin, springFestivalCompose)
				springfestial2021Group.POST("/click", authSvc.User, springFestivalIsJoin, springFestivalClick)
				springfestial2021Group.GET("/card_share", authSvc.User, springFestivalIsJoin, springFestivalCardShare)
				springfestial2021Group.POST("/get_card", authSvc.User, springFestivalGetCard)
				springfestial2021Group.GET("/card_token", springFestivalCardToken)
				springfestial2021Group.GET("/share_token", springFestivalShareToken)
			}
		}

		if service.CardSvc != nil {
			cardsGroup := group.Group("/cards")
			{
				cardsGroup.POST("/join", authSvc.User, cardsJoin)
				cardsGroup.POST("/draw", authSvc.User, cardsIsJoin, cardsDraw)
				cardsGroup.POST("/addtimes", authSvc.User, cardsIsJoin, cardsAddTimes)
				cardsGroup.GET("/times", authSvc.User, cardsIsJoin, cardsTimes)
				cardsGroup.POST("/follow", authSvc.User, cardsIsJoin, cardsFollow)
				cardsGroup.GET("/follower", authSvc.User, cardsFollower)
				cardsGroup.GET("/task", authSvc.User, cardsIsJoin, cardsTask)
				cardsGroup.GET("/cards", authSvc.User, cardsIsJoin, cardsCards)
				cardsGroup.GET("/invite_share", authSvc.User, cardsIsJoin, cardsInviteShare)
				cardsGroup.POST("/bind", authSvc.User, cardsBind)
				cardsGroup.POST("/compose", authSvc.User, cardsIsJoin, cardsCompose)
				cardsGroup.POST("/click", authSvc.User, cardsIsJoin, cardsClick)
				cardsGroup.GET("/card_share", authSvc.User, cardsIsJoin, cardsCardShare)
				cardsGroup.POST("/get_card", authSvc.User, cardsGetCard)
				cardsGroup.GET("/card_token", cardsCardToken)
				cardsGroup.GET("/share_token", cardsShareToken)
			}
		}

		if service.CardV2Svr != nil {
			cardsNewGroup := group.Group("/cards_new")
			{
				cardsNewGroup.POST("/join", authSvc.User, cardsNewJoin)
				cardsNewGroup.POST("/draw", authSvc.User, cardsNewIsJoin, cardsDrawNew)
				cardsNewGroup.POST("/addtimes", authSvc.User, cardsNewIsJoin, cardsAddTimesNew)
				cardsNewGroup.GET("/times", authSvc.User, cardsNewIsJoin, cardsTimesNew)
				cardsNewGroup.GET("/cards", authSvc.User, cardsNewIsJoin, cardsCardsNew)
				cardsNewGroup.GET("/invite_share", authSvc.User, cardsNewIsJoin, cardsInviteShareNew)
				cardsNewGroup.POST("/bind", authSvc.User, cardsNewBind)
				cardsNewGroup.POST("/compose", authSvc.User, cardsNewIsJoin, cardsComposeNew)
				cardsNewGroup.GET("/card_share", authSvc.User, cardsNewIsJoin, cardsCardShareNew)
				cardsNewGroup.POST("/get_card", authSvc.User, cardsGetCardNew)
				cardsNewGroup.GET("/card_token", cardsCardTokenNew)
				cardsNewGroup.GET("/share_token", cardsShareTokenNew)
				cardsNewGroup.POST("/send_points", authSvc.User, cardsNewIsJoin, cardsSendPoints)

			}
		}

		winterGroup := group.Group("/winter")
		{
			winterGroup.GET("/course", authSvc.User, winterCourse)
			winterGroup.POST("/join", authSvc.User, winterJoin)
			winterGroup.GET("/progress", authSvc.User, winterProgress)
		}
		commonActivityManuScript := group.Group("/manu_script")
		{
			commonActivityManuScript.GET("/aggregation", authSvc.User, ManuScriptAggregation)
			commonActivityManuScript.POST("/log/commit", authSvc.User, CommonActivityUserCommitContent)
			commonActivityManuScript.POST("/commit", authSvc.User, ManuScriptCommit)
			commonActivityManuScript.GET("/live", ManuScriptListInLive)
		}
		rankv3Group := group.Group("/rank_v3")
		{
			rankv3Group.GET("/result", rankv3Result)
		}
		examinationGroup := group.Group("/examination")
		{
			examinationGroup.POST("/up", authSvc.Guest, examinationUp)
		}
		fitActivityGroup := group.Group("/fit")
		{
			fitActivityGroup.GET("/user_info", authSvc.User, fitUserInfo)
			fitActivityGroup.GET("/plan_card_list", authSvc.Guest, getPlanCardList)
			fitActivityGroup.GET("/plan_card_detail", authSvc.Guest, getPlanCardDetail)
			fitActivityGroup.GET("/hot_tags_list", authSvc.Guest, getHotTagsList)
			fitActivityGroup.GET("/hot_videos_list", authSvc.Guest, getHotVideosByTag)
		}
		if service.Cpc100Svr != nil {
			cpc100YearGroup := group.Group("/cpc100")
			{
				cpc100YearGroup.GET("/info", authSvc.Guest, cpc100Info)
				cpc100YearGroup.GET("/reset", authSvc.User, cpc100Reset)
				cpc100YearGroup.POST("/unlock", authSvc.User, cpc100Unlock)
				cpc100YearGroup.GET("/count", cpc100Pv)
				cpc100YearGroup.GET("/total", cpc100Total)
			}
		}
		addExternalRewardsRouter(group)
		addExternalVoteRouter(group)
		addExternalBindRouter(group)
		// 暑期夏令营活动
		summerCampGroup := group.Group("/summer_camp")
		{
			summerCampGroup.GET("/user_info", authSvc.User, summerCampUserInfo)
			summerCampGroup.GET("/course_list", authSvc.Guest, getCourseList)
			summerCampGroup.GET("/user_course_info", authSvc.User, getUserCourseInfo)
			summerCampGroup.GET("/today_videos", authSvc.User, getUserTodayVideos)
			summerCampGroup.GET("/one_day_videos", authSvc.User, getOneDayVideos)
			summerCampGroup.POST("/start_plan", authSvc.User, userStartPlan)
			summerCampGroup.POST("/join_course", authSvc.User, joinCourse)
			summerCampGroup.POST("/exchange_prize", authSvc.User, exchangePrize)
			summerCampGroup.GET("/user_point_history", authSvc.User, userPointHistory)
			summerCampGroup.GET("/lottery_gift_list", summerLotteryGift)
			summerCampGroup.POST("/draw_lottery", authSvc.User, summerLottery)
			summerCampGroup.GET("/exchange_award_list", authSvc.User, exchangeAwardList)
		}
		addExternalMissionRouter(group)
	}
}

func internalRouter(e *bm.Engine) {
	group := e.Group("/x/internal/activity", bm.CSRF())
	{
		group.POST("/stockserver/sync/orders", syncOrders)
		group.GET("/is_vip_ticket", authSvc.User, hasVipTicket)
		group.GET("/bws/game/playable", authSvc.User, bwsGamePlayable2020)
		group.POST("/bws/game/play", authSvc.User, bwsGamePlay2020)
		group.GET("/bws/member", authSvc.User, bwsMember2020)
		group.POST("/bws/del_member_rank", bwsDelMemberRank2020)
		group.POST("/bws/add_member_rank", bwsAddMemberRank2020)
		group.GET("/bws/2020/user", user2020Internal)
		group.POST("/2020/add_heart", bws2020InternalAddHeart)
		group.GET("handwrite/personal_internal", handwritePersonalInternal)
		group.POST("/send_point", sendPoint)

		group.GET("/v2/guess/list", authSvc.Guest, v2GuessList)
		s10Group := group.Group("/s10")
		{
			s10Group.POST("/redelivery", redeliveryGift)
			s10Group.GET("/points", pointCache)
			s10Group.GET("/point/cost/detail", userCostPointsDetail)
			s10Group.GET("/goods/stock", delGoodsStock)
			s10Group.GET("/goods/round/stock", delRoundGoodsStock)
			s10Group.GET("/user/static", delUserStatic)
			s10Group.GET("/user/round/static", delRoundUserStatic)
			s10Group.GET("/lottery/infos", userLotteryInfo)
			s10Group.GET("/points/cost/static", userCostStatic)

			s10Group.GET("/sign2", signed2)
			s10Group.GET("/tasks2", tasksProgress2)
			s10Group.GET("/total2", totalPoints2)
			s10Group.GET("/stage2", matchesStage2)
			s10Group.GET("/lottery2", userMatchesLotteryInfo2)
			s10Group.GET("/stage/lottery/state2", updateUserLooteryState2)
			s10Group.GET("/stage/lottery2", stageLottery2)
			s10Group.GET("/goods2", actGoods2)
			s10Group.GET("/goods/exchange2", exchangeGoods2)
		}
		group.GET("/subject", vfySvc.Verify, subject)
		group.GET("/sub/info", vfySvc.Verify, subInfo)
		group.GET("/sub/infos", vfySvc.Verify, subInfos)
		group.POST("/vote", vfySvc.Verify, vote)
		group.GET("/ltime", vfySvc.Verify, ltime)
		group.GET("/reddot", vfySvc.Verify, redDot)
		group.GET("/reddot/clear", vfySvc.Verify, authSvc.Guest, clearRedDot)
		group.GET("/object/stat/set", vfySvc.Verify, setSubjectStat)
		group.GET("/view/rank/set", vfySvc.Verify, setViewRank)
		group.GET("/like/content/set", vfySvc.Verify, setLikeContent)
		group.GET("/likeact/add", vfySvc.Verify, addLikeAct)
		group.GET("/likeact/cache", vfySvc.Verify, likeActCache)
		group.GET("/likeact/state", vfySvc.Verify, likeActState)
		group.GET("/oids/info", vfySvc.Verify, likeOidsInfo)
		group.GET("/list", vfySvc.Verify, actListInfo)
		group.GET("/up/privilege", vfySvc.Verify, canCreateUpReserve)
		group.POST("/reserve/incr", vfySvc.Verify, reserveIncr)
		group.GET("/iir", vfySvc.Verify, resourceIir)
		group.GET("/up/check", vfySvc.Verify, upCheck)
		group.POST("/reserve/del_cache", delReserve)
		group.GET("/ugc/url", ugcURL)
		mcGroup := group.Group("/clear", vfySvc.Verify)
		{
			mcGroup.GET("/subject/up", subjectUp)
			mcGroup.GET("/like/up", likeUp)
			mcGroup.GET("/like/reload", actSetReload)
			mcGroup.GET("/like/ctime", likeCtimeCache)
			mcGroup.GET("/like/del/ctime", delLikeCtimeCache)

		}
		spGroup := group.Group("/sports")
		{
			spGroup.GET("/qq", vfySvc.Verify, qq)
			spGroup.GET("/news", vfySvc.Verify, news)
		}
		mactchGroup := group.Group("/match")
		{
			mactchGroup.GET("", matchs)
			mactchGroup.GET("/unstart", vfySvc.Verify, unStart)
			mactchGroup.POST("/cache/clear", clearCache)
			guGroup := mactchGroup.Group("/guess")
			{
				guGroup.GET("", vfySvc.Verify, guess)
				guGroup.GET("/list", vfySvc.Verify, listGuess)
				guGroup.POST("/add", vfySvc.Verify, addGuess)
			}
			foGroup := mactchGroup.Group("/follow")
			{
				foGroup.GET("", vfySvc.Verify, follow)
				foGroup.POST("/add", vfySvc.Verify, addFollow)
			}
		}
		initGroup := group.Group("/init")
		{
			initGroup.GET("/subject", vfySvc.Verify, subjectInit)
			initGroup.GET("/like", vfySvc.Verify, likeInit)
			initGroup.GET("/likeact", vfySvc.Verify, likeActCountInit)
			initGroup.GET("/subject/list", vfySvc.Verify, subjectLikeListInit)
		}
		tmGroup := group.Group("/timemachine")
		{
			tmGroup.GET("/start", startTmProc)
			tmGroup.GET("/stop", stopTmProc)
			tmGroup.GET("/2020/cache", userReport2020Cache)
			tmGroup.GET("/2020/filter", userReport2020Filter)
		}
		kfcIGroup := group.Group("/kfc")
		{
			kfcIGroup.POST("/deliver", vfySvc.Verify, deliverKfc)
		}
		preGroup := group.Group("/prediction")
		{
			preGroup.GET("/item/up", vfySvc.Verify, preItemUp)
			preGroup.GET("/up", vfySvc.Verify, preUp)
			preGroup.GET("/set/item", vfySvc.Verify, preSetItem)
			preGroup.GET("/set", vfySvc.Verify, preSet)
		}

		group.GET("/bnj2019/time/del", delTime)
		group.POST("/task/do", vfySvc.Verify, internalDoTask)
		group.POST("/task/add/award", vfySvc.Verify, addAwardTask)
		group.POST("/bws/achieve/add", vfySvc.Verify, addAchieve)
		group.GET("/bws/recharge/award", vfySvc.Verify, rechargeAward)
		group.POST("/lottery/addtimes", vfySvc.Verify, internalAddLotteryTimes)
		group.POST("/poll/vote", vfySvc.Verify, pollVote)
		group.POST("/up/list/his/add", vfySvc.Verify, addUpListHis)
		group.POST("/award/subject", vfySvc.Verify, awardSubject)
		group.POST("/bws/online/send/special/piece", vfySvc.Verify, bwsOnlineSendPiece)
		group.GET("/bws/online/tab/entrance", bwsOnlineTabEntrance)
		group.GET("/bws/online/reserve/award", bwsOnlineReserveAward)
		group.GET("/bws/online/import/offline/heart", bwsOnlineImportOfflineHeart)
		group.GET("/bws/online/import/offline/award", bwsOnlineImportOfflineAward)
		group.GET("/contribution/rank", totalRank)
		group.GET("/contribution/money", haveMoney)
		group.GET("/selection/one", selectionOne)
		group.GET("/selection/two", selectionTwo)
		group.GET("/productrole/reset", prReSet)
		group.GET("/productrole/maxvote", prMaxVote)
		group.GET("/productrole/notvote", prNotVote)
		group.GET("/up/vote/addtimes", upVoteAddTimes)
		group.GET("/up/vote/addtime", upAddTime)
		group.GET("/winter/finish", upProgress)
		group.GET("/winter/progress", winterInnerProgress)
		group.GET("/tag/convert", tagConvert)

		systemGroup := group.Group("/system")
		{
			systemGroup.GET("/user/info/v1", internalGetUserInfoByUID)
			systemGroup.GET("/users/info/v1", internalGetUsersInfoByUIDs)
			systemGroup.GET("/user/info/v2", internalGetUserInfoByCookie)

			systemGroup.GET("/user/add/v", internalAddV)                // 手动录入外包同学
			systemGroup.POST("/user/prize/notify", internalPrizeNotify) // 手动录入外包同学
		}

		group.GET("/memory/data", memoryData)
		group.GET("/cache/data", cacheData)
		group.GET("/add/data", addData)

		group.POST("/up_act_reserve/audit", vfySvc.Verify, upActReserveAudit)
	}

	newyear2021Group := group.Group("/bnj2021")
	{
		newyear2021Group.GET("/config", Bnj2021GetConfig)
		newyear2021Group.POST("/config", Bnj2021UpdateConfig)
		newyear2021Group.POST("/delete_config", Bnj2021DeleteConfig)

		newyear2021Group.GET("/configuration", ARConfiguration)
		newyear2021Group.GET("/live/draw/reissue", LiveDrawReIssue)
	}
	knowGroup := group.Group("/knowledge/badge")
	{
		knowGroup.POST("/config", badgeConfig)
	}

}

func ping(c *bm.Context) {
	if err := service.LikeSvc.Ping(c); err != nil {
		log.Error("activity interface ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(nil, nil)
}
