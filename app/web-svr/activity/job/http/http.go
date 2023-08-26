package http

import (
	"context"
	"net/http"
	"time"

	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/bnj"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/service"
	"go-gateway/app/web-svr/activity/job/service/bnj2021"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/model/rewards"
)

var ajSrv *service.Service

// Init .
func Init(conf *conf.Config, srv *service.Service) {
	ajSrv = srv
	engine := bm.DefaultServer(conf.BM)
	outerRouter(engine)
	if err := engine.Start(); err != nil {
		log.Error("httpx.Serve(%v) error(%+v)", conf.BM, err)
		panic(err)
	}
}

func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.GET("/bnj/2021/lottery/grpcLimit", bnjBizLimitRule)
	e.GET("/match/finish", finishMatch)
	e.GET("/guess/back", backGuess)
	e.GET("/guess/finish", finishGuess)
	e.POST("/guess/compensation", guessCompensation)
	e.POST("/guess/repair", compensationRepair)
	e.POST("/guess/repair/mid", compensationRepair4MID)
	e.POST("/guess/overissue/repair", overIssueRepair)
	e.POST("/guess/user/repair", userGuessRepair)
	e.POST("/guess/main/repair", guessMainRepair)
	e.GET("/guess/rank", rankGuess)
	e.GET("/bws/lottery", bwsLottery)
	e.GET("/bws/spec/lottery", bwsSpecLottery)
	e.POST("/bnj2020/msg/send", bnj2020MsgSend)
	e.GET("/image/rank/set", imageRankSet)
	e.GET("/faction/set", factionLikeSet)
	e.GET("/image/day/set", imageDayRankSet)
	e.GET("/stupid/set", stupidListSet)
	e.GET("/type/count/reset", resetTypeCount)
	e.GET("/handwrite/rank", handwriteRank)
	e.GET("/handwrite/data", handwriteData)
	e.GET("/handwrite/favsync", favSync)
	e.GET("/useraction/syncmid", userActionSyncMid)
	e.GET("/useraction/syncaid", userActionSyncAid)
	e.GET("/useraction/synconce", userActionSyncOnce)
	e.GET("/useraction/syncfull", userActionSyncFull)
	e.GET("/useraction/syncinfo", userActionSyncInfo)
	e.GET("/remix/rank", remixRank)
	e.GET("/remix/data", remixData)
	e.GET("/newstar/arc", newstarArc)
	e.GET("/newstar/finish", newstarFinish)
	e.GET("/newstar/identity", newstarIdentity)
	e.GET("/gameholiday/sync", gameHolidaySync)
	e.GET("/article/day/finish", articleDayFinish)
	e.GET("/article/day/risk", articleDayRisk)
	e.GET("/reserve/LoadNotifySubjectInfo", loadNotifySubjectInfo)
	e.GET("/college/rank", collegeRank)
	e.GET("/college/update_version", collegeUpdateVersion)
	e.GET("/share/share_url", shareUpdateURL)
	e.GET("/dubbing/rank", dubbingRank)
	e.GET("/college/video_bonus", collegeVideoBonus)
	e.GET("/college/member_score", collegeMemberScore)
	e.GET("/contribution/calc", crontributionCalc)
	e.GET("/contribution/user_db", crontributionDB)
	e.GET("/contribution/rank", crontributionRank)
	e.GET("/contribution/bcut_likes", crontributionBcut)
	e.GET("/lottery/addtimes", lotteryAddTimes)
	e.GET("/dubbing/data", dubbingData)
	//e.GET("/funny/sync", funnySync)
	e.GET("/funny/caculate_part_one", funnyCaculatePartOne)
	e.GET("/funny/caculate_part_two", funnyCaculatePartTwo)
	e.GET("/column/export", columnDataExport)
	e.GET("/acg/2020/UpdateTaskState", acg2020UpdateTaskState)
	e.GET("/college/college_auto", collegeScoreAuto)
	e.GET("/answer/hour", answerHourCalc)
	e.GET("/answer/week", answerWeekCalc)
	e.GET("/rank/do", rankDo)
	e.GET("/rank/cron_map", rankCronMap)
	e.GET("/rank/cron_new", rankJobNew)
	e.GET("/productrole/assistance", prAssistance)
	e.GET("/productrole/dayreport", dayVoteReport)
	e.GET("/productrole/votereport", voteReport)
	e.GET("/productrole/midreport", midReport)
	e.GET("/tunnel/group/all", tunnelGroupAll)
	e.GET("/yellow/green/yingyuan", ygYingYuanVote)
	e.GET("/knowledge", Knowledge)
	e.GET("/handwrite2021/rank", Handwrite2021)
	e.POST("/rewards/send", sendAwardDirectly)
	e.GET("/handwrite2021/data", Handwrite2021Data)
	e.POST("/bnj2021/reserve/live/repair", resendReserveLiveAward)
	e.POST("/bnj2021/live/coupon/reissue", reissueLiveARCoupon)
	e.GET("/bnj2021/live/coupon/check", checkLiveARCoupon)
	e.GET("/rank/rank_log", rankLog)
	e.GET("/knowledge/history/delete", knowDelHistory)
	e.GET("/knowledge/task/calc", knowTaskCalc)
	// fit健身打卡相关
	e.POST("/fit/flush_view_data", flushViewData)
	e.POST("/fit/send_push_card", sendPushCard)
	e.POST("/fit/set_member_into_rw", setMemberIntToRW)
	e.GET("/fit/export_user", exportUser)
	e.GET("/cache/data", cacheData)
	e.GET("/gaokao/lateX2png", lateX2png)
	e.GET("/gaokao/transLatexInfo", uploadPng2Bos)
}

func exportUser(ctx *bm.Context) {
	ctx.JSON(ajSrv.FitSvr.ExportUser(context.Background()))
}

func setMemberIntToRW(ctx *bm.Context) {
	go ajSrv.FitSvr.SetMemberIntToRWHttp(context.Background())
	ctx.JSON(nil, nil)
}

func flushViewData(ctx *bm.Context) {
	go ajSrv.FitSvr.FlushPlanCardData(context.Background())
	ctx.JSON(nil, nil)
}

func sendPushCard(ctx *bm.Context) {
	go ajSrv.FitSvr.SendTianMaCardHttp(context.Background())
	ctx.JSON(nil, nil)
}

func lateX2png(ctx *bm.Context) {
	v := new(struct {
		Latex string `form:"latex" json:"latex"  validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(ajSrv.LaTeX2PNG(ctx, v.Latex))
}

func uploadPng2Bos(ctx *bm.Context) {
	v := new(struct {
		Url string `form:"url" json:"url"  validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(ajSrv.UploadPng2Bos(ctx, v.Url))
}

func rankLog(ctx *bm.Context) {
	ctx.JSON(nil, ajSrv.SetRankLog())
}

func checkLiveARCoupon(ctx *bm.Context) {
	v := new(struct {
		MID int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(bnj.FetchBnjLiveUserRecordList(ctx, v.MID))
}

func reissueLiveARCoupon(ctx *bm.Context) {
	v := new(struct {
		MID    int64 `form:"mid" validate:"min=1"`
		Coupon int64 `form:"coupon" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, bnj2021.ReIssueLiveARCouponByMidAndCoupon(v.MID, v.Coupon))
}

func resendReserveLiveAward(ctx *bm.Context) {
	v := new(struct {
		MID     int64 `form:"mid" validate:"min=1"`
		AwardID int64 `form:"awardID" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, bnj2021.ReSendReserveLiveAward(ctx, v.MID, v.AwardID))
}

func ping(c *bm.Context) {
	if err := ajSrv.Ping(c); err != nil {
		log.Error("activity-job ping error")
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func lotteryAddTimes(c *bm.Context) {
	v := new(struct {
		ActionType int `form:"action_type" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(ajSrv.LotteryAddTimes(c, v.ActionType))
}

func resetTypeCount(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ajSrv.ResetLikeTypeCount(c, v.Sid))
}

func bnjBizLimitRule(c *bm.Context) {
	c.JSON(bnj2021.BnjBizLimitRule(), nil)
}

func finishMatch(c *bm.Context) {
	v := new(struct {
		MoID int64 `form:"mo_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ajSrv.FinishMatch(c, v.MoID))
}

func backGuess(c *bm.Context) {
	v := new(struct {
		MainID   int64  `form:"main_id" validate:"min=1"`
		Business int64  `form:"business" validate:"min=1"`
		Oid      int64  `form:"oid" validate:"min=1"`
		Title    string `form:"title" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.BackGuess(v.MainID, v.Business, v.Oid, v.Title)
}

func finishGuess(c *bm.Context) {
	v := new(struct {
		MainID   int64 `form:"main_id" validate:"min=1"`
		ResultID int64 `form:"result_id" validate:"min=1"`
		Business int64 `form:"business" validate:"min=1"`
		Oid      int64 `form:"oid" validate:"min=1"`
		Debug    bool  `form:"debug"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.FinishGuess(v.MainID, v.ResultID, v.Business, v.Oid, v.Debug)
}

func compensationRepair(c *bm.Context) {
	v := new(service.CompensationRepair)
	if err := c.Bind(v); err != nil {
		return
	}
	err := ajSrv.CompensationRepair(c, v)
	c.JSON(nil, err)
}

func compensationRepair4MID(c *bm.Context) {
	v := new(service.CompensationRepair4MID)
	if err := c.Bind(v); err != nil {
		return
	}
	m := ajSrv.CompensationRepair4MID(c, v)
	c.JSON(m, nil)
}

func userGuessRepair(c *bm.Context) {
	v := new(service.GuessRepair4MID)
	if err := c.Bind(v); err != nil {
		return
	}
	m, err := ajSrv.UserGuessRepair(v)
	c.JSON(m, err)
}

func guessMainRepair(c *bm.Context) {
	v := new(service.GuessRepair4MainID)
	if err := c.Bind(v); err != nil {
		return
	}
	m, err := ajSrv.GuessMainRepair(v)
	c.JSON(m, err)
}

func overIssueRepair(c *bm.Context) {
	v := new(service.OverIssueRepair)
	if err := c.Bind(v); err != nil {
		return
	}
	err := ajSrv.RepairSpecifiedUserGuess(v)
	c.JSON(nil, err)
}

func guessCompensation(c *bm.Context) {
	v := new(struct {
		MIDs     []int64 `form:"mids,split" validate:"required,dive,gt=0"`
		MainID   int64   `form:"main_id" validate:"min=1"`
		ResultID int64   `form:"result_id" validate:"min=1"`
		Business int64   `form:"business" validate:"min=1"`
		Oid      int64   `form:"oid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	d := ajSrv.GuessCompensation(c, v.MIDs, v.MainID, v.ResultID, v.Business, v.Oid)
	c.JSON(d, nil)
}

func rankGuess(c *bm.Context) {
	v := new(struct {
		Business int64 `form:"business" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.CalcRank(v.Business)
}

func bwsLottery(c *bm.Context) {
	go func() {
		ajSrv.CreateLotteryUsers()
	}()
	c.JSON(nil, nil)
}

func bwsSpecLottery(c *bm.Context) {
	go func() {
		ajSrv.CreateSpecLottery()
	}()
	c.JSON(nil, nil)
}

func bnj2020MsgSend(ctx *bm.Context) {
	v := new(struct {
		Mids []int64 `form:"mids,split" validate:"required,dive,min=1"`
		Send int64   `form:"send" default:"0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(ajSrv.Bnj2020MessageSend(ctx, v.Mids, v.Send))
}

func imageRankSet(c *bm.Context) {
	ajSrv.SetImageLikes()
	c.JSON(nil, nil)
}

func factionLikeSet(c *bm.Context) {
	ajSrv.SetFactionLikes()
	c.JSON(nil, nil)
}

func imageDayRankSet(c *bm.Context) {
	go func() {
		ajSrv.SetDayImage()
	}()
	c.JSON(nil, nil)
}

func stupidListSet(c *bm.Context) {
	ajSrv.SetStupidList()
	c.JSON(nil, nil)
}

func handwriteRank(c *bm.Context) {
	go func() {
		ajSrv.HandWriteMemberScore()
	}()
	c.JSON(nil, nil)
}

func handwriteData(c *bm.Context) {
	go func() {
		ajSrv.DataResult()
	}()
	c.JSON(nil, nil)
}

func favSync(c *bm.Context) {
	go func() {
		ajSrv.FavSyncCounterFilter()
	}()
	c.JSON(nil, nil)
}

func userActionSyncMid(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" json:"sid" validate:"required,min=1"`
		Mid int64 `form:"mid" json:"mid" validate:"required,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.SyncMid2Counter(c, &like.Reserve{
		Mid: v.Mid,
		Sid: v.Sid,
	})
	c.JSON(nil, nil)
}

func userActionSyncAid(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" json:"sid" validate:"required,min=1"`
		Aid int64 `form:"aid" json:"aid" validate:"required,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.SyncAvid2Counter(c, &like.Item{
		Wid: v.Aid,
		Sid: v.Sid,
	})
	c.JSON(nil, nil)
}

func userActionSyncFull(c *bm.Context) {
	v := new(like.SubjectRule)
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.SyncFullData2Counter(c, v)
	c.JSON(nil, nil)
}

func userActionSyncOnce(c *bm.Context) {
	ajSrv.UserActionStatSyncOnce()
	c.JSON(nil, nil)
}

func userActionSyncInfo(c *bm.Context) {
	c.JSON(ajSrv.UserActionSyncInfo(), nil)
}
func remixRank(c *bm.Context) {
	go func() {
		ajSrv.RemixEveryHour()
	}()
	c.JSON(nil, nil)

}

func remixData(c *bm.Context) {
	go func() {
		ajSrv.RemixDataResult()
	}()
	c.JSON(nil, nil)
}

func newstarArc(c *bm.Context) {
	go func() {
		ajSrv.NewstarArchiveTask()
	}()
	c.JSON(nil, nil)
}

func newstarFinish(c *bm.Context) {
	go func() {
		ajSrv.FinishNewstar()
	}()
	c.JSON(nil, nil)
}

func newstarIdentity(c *bm.Context) {
	go func() {
		ajSrv.IdentityChecking()
	}()
	c.JSON(nil, nil)
}

func gameHolidaySync(c *bm.Context) {
	go func() {
		ajSrv.GameHolidaySyncCounterFilter()
	}()
	c.JSON(nil, nil)
}

func articleDayFinish(c *bm.Context) {
	go func() {
		ajSrv.FinishArticleDay()
	}()
	c.JSON(nil, nil)
}

func articleDayRisk(c *bm.Context) {
	v := new(struct {
		Mid    int64 `form:"mid" json:"mid" validate:"required,min=1"`
		Status int64 `form:"status" json:"status" validate:"required,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ajSrv.ArticleDayRisk(c, v.Mid, v.Status)
	c.JSON(nil, nil)
}

func loadNotifySubjectInfo(c *bm.Context) {
	c.JSON(ajSrv.LoadNotifySubjectInfo(c))
}

func collegeRank(c *bm.Context) {
	go func() {
		ajSrv.CollegeRank()
	}()
	c.JSON(nil, nil)
}

func collegeUpdateVersion(c *bm.Context) {
	go func() {
		ajSrv.CollegeVersion()
	}()
	c.JSON(nil, nil)
}

func shareUpdateURL(c *bm.Context) {
	go func() {
		ajSrv.ShareURLUpdate()
	}()
	c.JSON(nil, nil)
}

func dubbingRank(c *bm.Context) {
	go func() {
		ajSrv.DubbingRank()
	}()
	c.JSON(nil, nil)
}

func collegeVideoBonus(c *bm.Context) {
	go func() {
		ajSrv.CollegeBonus()
	}()
	c.JSON(nil, nil)
}

func collegeMemberScore(c *bm.Context) {
	go func() {
		ajSrv.CollegeScore()
	}()
	c.JSON(nil, nil)
}

func crontributionCalc(c *bm.Context) {
	go func() {
		ajSrv.CalcUserContribution()
	}()
	c.JSON(nil, nil)
}

func crontributionDB(c *bm.Context) {
	go func() {
		ajSrv.UpUserContributionDB()
	}()
	c.JSON(nil, nil)
}

func crontributionRank(c *bm.Context) {
	go func() {
		ajSrv.ScoreTotalRank()
	}()
	c.JSON(nil, nil)
}
func crontributionBcut(c *bm.Context) {
	go func() {
		ajSrv.UpUserBcutLikes()
	}()
	c.JSON(nil, nil)
}

func dubbingData(c *bm.Context) {
	go func() {
		ajSrv.DubbingData()
	}()
	c.JSON(nil, nil)
}

//func funnySync(c *bm.Context) {
//	go func() {
//		ajSrv.FunnySyncVideoData()
//	}()
//	c.JSON(nil, nil)
//}

func columnDataExport(c *bm.Context) {
	go func() {
		ajSrv.ColumnDataExport()
	}()
	c.JSON(nil, nil)
}

func acg2020UpdateTaskState(c *bm.Context) {
	v := new(struct {
		Time time.Time `form:"time" json:"time" time_format:"2006-01-02 15:04:05"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	go func() {
		if v.Time.IsZero() {
			v.Time = time.Now()
		}
		ajSrv.AcgSrv.UpdateTaskState(context.Background(), v.Time)
	}()
	c.JSON(nil, nil)
}
func collegeScoreAuto(c *bm.Context) {
	go func() {
		ajSrv.CollegeScoreAuto()
	}()
	c.JSON(nil, nil)
}

func funnyCaculatePartOne(c *bm.Context) {
	go func() {
		ajSrv.CaculatePartOne()
	}()
	c.JSON(nil, nil)
}

func answerHourCalc(c *bm.Context) {
	go func() {
		ajSrv.AnswerHour()
	}()
	c.JSON(nil, nil)
}

func funnyCaculatePartTwo(c *bm.Context) {
	go func() {
		ajSrv.CaculatePartTwo()
	}()
	c.JSON(nil, nil)
}

func answerWeekCalc(c *bm.Context) {
	v := new(struct {
		StrWeek string `form:"week" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	go func() {
		ajSrv.AnswerHttpWeek(v.StrWeek)
	}()
	c.JSON(nil, nil)
}

func rankDo(c *bm.Context) {
	v := new(struct {
		ID            int64 `form:"id" validate:"required"`
		AttributeType uint  `form:"attribute"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	go func() {
		ajSrv.RankJob(v.ID, v.AttributeType)
	}()
	c.JSON(nil, nil)

}

func rankCronMap(c *bm.Context) {

	c.JSON(ajSrv.RankCronMap(), nil)
}

func rankJobNew(c *bm.Context) {
	go func() {
		ajSrv.StartRankJob()
	}()
	c.JSON(nil, nil)
}
func prAssistance(c *bm.Context) {
	go func() {
		ajSrv.SelAssistance()
	}()
	c.JSON(nil, nil)
}

func dayVoteReport(c *bm.Context) {
	go func() {
		ajSrv.DayVoteReport()
	}()
	c.JSON(nil, nil)
}

func voteReport(c *bm.Context) {
	go func() {
		ajSrv.VoteReport()
	}()
	c.JSON(nil, nil)
}

func midReport(c *bm.Context) {
	go func() {
		ajSrv.MidReport()
	}()
	c.JSON(nil, nil)
}

func tunnelGroupAll(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	go func() {
		ajSrv.TunnelGroupAllUser(v.Sid)
	}()
	c.JSON(nil, nil)
}

func ygYingYuanVote(c *bm.Context) {
	go func() {
		ajSrv.YingYuanVote()
	}()
	c.JSON(nil, nil)
}

func Knowledge(c *bm.Context) {
	go func() {
		ajSrv.ActKnowledge()
	}()
	c.JSON(nil, nil)
}

func Handwrite2021(c *bm.Context) {
	go func() {
		ajSrv.Handwrite2021()
	}()
	c.JSON(nil, nil)
}

func Handwrite2021Data(c *bm.Context) {
	go func() {
		ajSrv.Handwrite2021Data()
	}()
	c.JSON(nil, nil)
}

func sendAwardDirectly(c *bm.Context) {
	v := &rewards.AsyncSendingAwardInfo{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(ajSrv.SendAward(v))
}

func knowDelHistory(ctx *bm.Context) {
	v := new(struct {
		LogDate string `form:"log_date" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	// 得到上一天日期
	t := time.Now()
	newTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	beforeDate := newTime.AddDate(0, 0, -1).Format("20060102")
	if v.LogDate == beforeDate {
		err := xecode.Errorf(xecode.RequestErr, "不能删除前一天的数据")
		ctx.JSON(nil, err)
		return
	}
	go func() {
		ajSrv.DeleteKnowledgeCalculateDB(v.LogDate)
	}()
	ctx.JSON("ok", nil)
}

func knowTaskCalc(ctx *bm.Context) {
	v := new(struct {
		LogDate string `form:"log_date"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.LogDate == "" {
		// 得到上一天日期
		t := time.Now()
		newTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		v.LogDate = newTime.AddDate(0, 0, -1).Format("20060102")
	}
	go func() {
		ajSrv.UserKnowledgeTaskCalculate(v.LogDate)
	}()
	ctx.JSON("ok", nil)
}

func cacheData(c *bm.Context) {
	v := new(struct {
		Typ int64 `form:"type" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(ajSrv.CacheData(v.Typ), nil)
}
