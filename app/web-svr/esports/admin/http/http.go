package http

import (
	"strconv"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/admin/service"
)

var (
	esSvc *service.Service
	//idfSvc  *identify.Identify
	permitSvc *permit.Permit
)

// Init init http sever instance.
func Init(c *conf.Config, s *service.Service) {
	esSvc = s
	permitSvc = permit.New2(nil)
	engine := bm.DefaultServer(c.BM)
	authRouter(engine)
	// init internal server
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func authRouter(e *bm.Engine) {
	e.Ping(ping)
	e.POST("/x/admin/esports/migration/poster", renewPosterFromAllTeam)
	e.POST("/x/admin/esports/migration/poster/list", renewPosterByTeamIDList)
	baseGroupNoAuth := e.Group("/x/admin/esports")
	selfHelp := baseGroupNoAuth.Group("/selfHelp")
	{
		selfHelp.GET("guess/seasons", guessSeasons)
		selfHelp.GET("guess/joinList", guessContests)
	}
	group := e.Group("/x/admin/esports", permitSvc.Permit2("ESPORTS_ADMIN"))
	{
		contestSeries := group.Group("/contest_series")
		{
			contestSeries.GET("/score_rules", getScoreRules)
			contestSeries.POST("/score_rules/save", saveScoreRules)
			contestSeries.GET("", fetchContestSeries)
			contestSeries.GET("/list", contestSeriesList)
			contestSeries.POST("", createContestSeries)
			contestSeries.PUT("", updateContestSeries)
			contestSeries.DELETE("", deleteContestSeries)
			contestSeriesPointMatch := contestSeries.Group("/point_match")
			{
				contestSeriesPointMatch.GET("get", getContestSeriesPointMatchConfig)
				contestSeriesPointMatch.POST("add", addContestSeriesPointMatchConfig)
				contestSeriesPointMatch.POST("update", updateContestSeriesPointMatchConfig)

				contestSeriesPointMatch.POST("table/preview", previewPointMatchInfo)
				contestSeriesPointMatch.POST("table/refresh", refreshPointMatchInfo)
				contestSeriesPointMatch.GET("table/get", getPointMatchInfo)
			}
			contestSeriesKnockoutMatch := contestSeries.Group("/knockout_match")
			{
				contestSeriesKnockoutMatch.GET("get", getContestSeriesKnockoutMatchConfig)
				contestSeriesKnockoutMatch.POST("add", addContestSeriesKnockoutMatchConfig)
				contestSeriesKnockoutMatch.POST("update", updateContestSeriesKnockoutMatchConfig)

				contestSeriesKnockoutMatch.POST("tree/preview", previewKnockoutMatchInfo)
				contestSeriesKnockoutMatch.POST("tree/refresh", refreshKnockoutMatchInfo)
				contestSeriesKnockoutMatch.GET("tree/get", getKnockoutMatchInfo)
			}

		}
		matchGroup := group.Group("/matchs")
		{
			matchGroup.GET("/info", matchInfo)
			matchGroup.GET("/list", matchList)
			matchGroup.POST("/add", addMatch)
			matchGroup.POST("/save", editMatch)
			matchGroup.POST("/forbid", forbidMatch)
		}
		seasonGroup := group.Group("/seasons")
		{
			seasonGroup.GET("/info", seasonInfo)
			seasonGroup.POST("/big/fix", bigFix)
			seasonGroup.GET("/list", seasonList)
			seasonGroup.POST("/add", addSeason)
			seasonGroup.POST("/save", editSeason)
			seasonGroup.POST("/forbid", forbidSeason)
			seasonTeamGroup := seasonGroup.Group("/team")
			{
				seasonTeamGroup.GET("/list", ListTeamInSeason)
				seasonTeamGroup.POST("/add", AddTeamToSeason)
				seasonTeamGroup.POST("/delete", RemoveTeamFromSeason)
				seasonTeamGroup.POST("/save", UpdateTeamInSeason)
				seasonTeamGroup.POST("/rebuild", RebuildTeamInSeason)

			}
			seasonRankGroup := seasonGroup.Group("/rank")
			{
				seasonRankGroup.GET("/info", rankInfo)
				seasonRankGroup.GET("/list", rankList)
				seasonRankGroup.POST("/add", addSeasonRank)
				seasonRankGroup.POST("/save", editSeasonRank)
				seasonRankGroup.POST("/forbid", forbidRankSeason)
			}
		}
		contestGroup := group.Group("/contest")
		{
			contestGroup.GET("/info", contestInfo)
			contestGroup.GET("/list", contestList)
			contestGroup.POST("/add", addContest)
			contestGroup.POST("/save", editContest)
			contestGroup.POST("/forbid", forbidContest)
			contestGroup.POST("/match/fix", matchFix)
			contestGroup.GET("/teams", contestTeams)
			contestGroup.POST("/teams/save", contestTeamsSave)
			contestGroup.POST("/teams/update", contestTeamsUpdate)
			contestGroup.GET("/teams/check", contestTeamsCheck)
			contestGroup.GET("/scores", contestTeamScores)
			contestGroup.POST("/scores/save", contestTeamScoresSave)
		}
		gameGroup := group.Group("/games")
		{
			gameGroup.GET("/info", gameInfo)
			gameGroup.GET("/list", gameList)
			gameGroup.POST("/add", addGame)
			gameGroup.POST("/save", editGame)
			gameGroup.POST("/forbid", forbidGame)
			gameGroup.GET("/types", types)
			gameGroup.GET("/team", gameTeams)
			gameGroup.GET("/season", gameSeasons)
		}
		teamGroup := group.Group("/teams")
		{
			teamGroup.GET("/info", teamInfo)
			teamGroup.GET("/list", teamList)
			teamGroup.POST("/add", addTeam)
			teamGroup.POST("/save", editTeam)
			teamGroup.POST("/forbid", forbidTeam)
		}
		tagGroup := group.Group("/tags")
		{
			tagGroup.GET("/info", tagInfo)
			tagGroup.GET("/list", tagList)
			tagGroup.POST("/add", addTag)
			tagGroup.POST("/save", editTag)
			tagGroup.POST("/forbid", forbidTag)
		}
		arcGroup := group.Group("/arcs")
		{
			arcGroup.GET("/list", arcList)
			arcGroup.POST("/edit", editArc)
			arcGroup.POST("/batch/add", batchAddArc)
			arcGroup.POST("/batch/edit", batchEditArc)
			arcGroup.POST("/batch/del", batchDelArc)
			arcGroup.POST("/import/csv", arcImportCSV)
			arcGroup.POST("/batch/pass", batchPassArc)
			arcGroup.POST("/batch/nopass", batchNopassArc)
		}
		whiteGroup := group.Group("/white")
		{
			whiteGroup.GET("/info", whiteInfo)
			whiteGroup.GET("/list", whiteList)
			whiteGroup.POST("/add", addWhite)
			whiteGroup.POST("/edit", editWhite)
			whiteGroup.POST("/batch/del", delWhite)
			whiteGroup.POST("/import/csv", importWhite)
		}
		autoTagGroup := group.Group("/autotag")
		{
			autoTagGroup.GET("/info", autotagInfo)
			autoTagGroup.GET("/list", autotagList)
			autoTagGroup.POST("/add", addAutotag)
			autoTagGroup.POST("/edit", editAutotag)
			autoTagGroup.POST("/batch/del", delAutotag)
			autoTagGroup.POST("/import/csv", importAutotag)
		}
		keyGroup := group.Group("/keyword")
		{
			keyGroup.GET("/info", keywordInfo)
			keyGroup.GET("/list", keywordList)
			keyGroup.POST("/add", addKeyword)
			keyGroup.POST("/edit", editKeyword)
			keyGroup.POST("/batch/del", delKeyword)
			keyGroup.POST("/import/csv", importKeyword)
		}
		actGroup := group.Group("/active")
		{
			actGroup.GET("", listAct)
			actGroup.POST("/add", addAct)
			actGroup.POST("/edit", editAct)
			actGroup.POST("/forbid", forbidAct)
			dGroup := actGroup.Group("/detail")
			{
				dGroup.GET("/list", listDetail)
				dGroup.POST("/add", addDetail)
				dGroup.POST("/edit", editDetail)
				dGroup.POST("/forbid", forbidDetail)
				dGroup.POST("/online", onLine)
			}
			tGroup := actGroup.Group("/tree")
			{
				tGroup.GET("/list", listTree)
				tGroup.POST("/add", addTree)
				tGroup.POST("/edit", editTree)
				tGroup.POST("/del", delTree)
			}
		}
		guessGroup := group.Group("/guess")
		{
			guessGroup.POST("/add", addGuess)
			guessGroup.POST("/del", delGuess)
			guessGroup.GET("/list", listGuess)
			guessGroup.POST("/result", resultGuess)
		}
		searchGroup := group.Group("/intervene")
		{
			searchGroup.GET("/info", infoIntervene)
			searchGroup.GET("/list", listIntervene)
			searchGroup.POST("/add", addIntervene)
			searchGroup.POST("/edit", editIntervene)
			searchGroup.POST("/forbid", forbidIntervene)
		}
		posterGroup := group.Group("/poster")
		{
			posterGroup.POST("/create", createPoster)
			posterGroup.POST("/edit", editPoster)
			posterGroup.POST("/toggle", togglePoster)
			posterGroup.POST("/center", centerPoster)
			posterGroup.POST("/delete", deletePoster)
			posterGroup.GET("/list", getPosterList)
			posterGroup.GET("/effectiveList", getEffectivePosterList)
		}
		s10Group := group.Group("/s10")
		{
			s10Group.POST("/rank/data/intervention", rankDataInterventionSave)
			s10Group.GET("/rank/data/intervention", rankDataInterventionGet)
		}
		topicGroup := group.Group("/topic")
		{
			topicGroup.POST("/video/add", addTopicVideoList)
			topicGroup.POST("/video/update", editTopicVideoList)
			topicGroup.POST("/video/forbid", forbidTopicVideoList)
			topicGroup.GET("/video/info", infoTopicVideoList)
			topicGroup.GET("/video/list", topicVideoLists)
			topicGroup.POST("/archive/check", checkArchive)
			topicGroup.GET("/video/filter", videoFilter)
		}
		wallGroup := group.Group("/reply/wall")
		{
			wallGroup.GET("/list", wallList)
			wallGroup.POST("/save", wallSave)
		}
	}
}

func ping(c *bm.Context) {

}

func userInfo(c *bm.Context) (rs *model.BaseInfo) {
	rs = new(model.BaseInfo)
	if nameInter, ok := c.Get("username"); ok {
		rs.Name = nameInter.(string)
	}
	if uidInter, ok := c.Get("uid"); ok {
		rs.ID = uidInter.(int64)
	}
	if rs.Name == "" {
		ck, err := c.Request.Cookie("username")
		if err != nil {
			log.Error("userInfo get cookie error (%v)", err)
			return
		}
		rs.Name = ck.Value
	}
	if rs.ID == 0 {
		ck, err := c.Request.Cookie("uid")
		if err != nil {
			log.Error("userInfo get cookie error (%v)", err)
			return
		}
		uidInt, _ := strconv.Atoi(ck.Value)
		rs.ID = int64(uidInt)
	}
	return
}
