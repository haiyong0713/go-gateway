package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/service"
)

var (
	authn  *auth.Auth
	vfySvr *verify.Verify
	eSvc   *service.Service
)

// Init init http server
func Init(c *conf.Config, s *service.Service) {
	authn = auth.New(c.Auth)
	vfySvr = verify.New(c.Verify)
	eSvc = s
	engine := bm.DefaultServer(c.BM)
	limiter := quota.New(c.Limiter)
	engine.Use(limiter.Handler())
	outerRouter(engine)
	internalRouter(engine)
	if err := engine.Start(); err != nil {
		log.Error("httpx.Serve error(%v)", err)
		panic(err)
	}
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Use(bm.CORS())
	e.Ping(ping)
	group := e.Group("/x/esports")
	{
		group.GET("/season", season)
		group.GET("/season/teams", ListTeamBySeason)
		group.GET("/season/series/point_match", getSeriesPointMatchInfo)
		group.GET("/season/series/knockout_match", authn.Guest, getSeriesKnockoutMatchInfo)
		group.GET("/app/season", appSeason)
		group.GET("/match/seasons/info", authn.Guest, matchSeasons)
		group.GET("/seasons/info", authn.Guest, seasonsInfo)
		group.GET("/season/teams/info", authn.Guest, seasonTeams)
		group.GET("/web/reply/wall", authn.Guest, webReplyWall)
		autoSub := group.Group("/auto_sub")
		{
			autoSub.GET("", authn.User, autoSubscribeStatus)
			autoSub.POST("", authn.User, autoSubscribe)
		}
		componentGroup := group.Group("/component")
		{
			componentGroup.GET("/contests/time", authn.Guest, timeContests)
			componentGroup.GET("/contests/all", authn.Guest, allContests)
			componentGroup.GET("/contests/all/fold", authn.Guest, allFold)
			componentGroup.GET("/contests/battleground", authn.Guest, battleContests)
			componentGroup.GET("/contests/abstract", authn.Guest, abstract)
			componentGroup.GET("/contests", authn.Guest, seasonContests)
			componentGroup.GET("/contests/battleground/teams", authn.Guest, battleContestTeams)
			componentGroup.GET("/teams/contests", authn.Guest, teamContests)
			componentGroup.GET("/v2/teams/contests", authn.Guest, teamContestsV2)
			componentGroup.GET("/howe_away/contests", howeAwayContests)
			componentGroup.GET("/season/teams", seasonTeamsComponent)
			componentGroup.GET("/video/list", videoListComponent)
			componentGroup.GET("/contests/reply/wall", contestReplyWall)
		}
		matGroup := group.Group("/matchs")
		{
			matGroup.GET("/filter", filterMatch)
			matGroup.GET("/list", authn.Guest, listContest)
			matGroup.GET("/app/list", authn.Guest, appContest)
			matGroup.GET("/calendar", calendar)
			matGroup.GET("/active", authn.Guest, active)
			matGroup.GET("/videos", authn.Guest, actVideos)
			matGroup.GET("/points", authn.Guest, actPoints)
			matGroup.GET("/top", authn.Guest, actTop)
			matGroup.GET("/knockout", authn.Guest, actKnockout)
			matGroup.GET("/info", authn.Guest, contest)
			matGroup.GET("/recent", authn.Guest, recent)
			matGroup.GET("/intervene", intervene)
			matGroup.GET("/infov2", authn.Guest, contestV2)
		}
		videoGroup := group.Group("/video")
		{
			videoGroup.GET("/filter", filterVideo)
			videoGroup.GET("/list", listVideo)
			videoGroup.GET("/search", authn.Guest, search)
		}
		favGroup := group.Group("/fav")
		{
			favGroup.GET("", authn.Guest, listFav)
			favGroup.GET("/season", authn.Guest, seasonFav)
			favGroup.GET("/stime", authn.Guest, stimeFav)
			favGroup.GET("/list", authn.Guest, appListFav)
			favGroup.POST("/add", authn.User, addFav)
			favGroup.POST("/del", authn.User, delFav)
			favGroup.GET("/batch", authn.User, batchQueryFav)
			favGroup.POST("/batch", authn.User, batchAddFav)
		}
		pointGroup := group.Group("/leida")
		{
			pointGroup.GET("/game", game)
			pointGroup.GET("/game/types", types)
			pointGroup.GET("/items", items)
			pointGroup.GET("/heroes", heroes)
			pointGroup.GET("/abilities", abilities)
			pointGroup.GET("/players", players)
			pointGroup.GET("/teams", teams)
			pointGroup.GET("/seasons", seasons)
			pointGroup.GET("/roles", roles)
		}
		bigGroup := group.Group("/stats")
		{
			bigGroup.GET("/player", bigPlayers)
			bigGroup.GET("/team", bigTeams)
			bigGroup.GET("/player/mvp/rank", mvpRank)
			bigGroup.GET("/player/kda/rank", kdaRank)
			bigGroup.GET("/hero2/rank", hero2Rank)
		}
		specialGroup := group.Group("/special")
		{
			specialGroup.GET("/teams", specialTeams)
			specialGroup.GET("/team", authn.Guest, specTeam)
			specialGroup.GET("/player", specPlayer)
			specialGroup.GET("/player/recent", authn.Guest, playerRecent)
		}
		guessGroup := group.Group("/guess")
		{
			guessGroup.GET("", authn.Guest, guessDetail)
			guessGroup.GET("/v2/season/summary", authn.User, userSeasonGuessSummary)
			guessGroup.GET("/v2/season", authn.User, userSeasonGuessList)
			guessGroup.GET("/list", authn.Guest, guessList)
			guessGroup.GET("/moreshow", guessMoreShow)
			guessGroup.GET("/coin", authn.User, guessDetailCoin)
			guessGroup.POST("add", authn.User, guessDetailAdd)
			guessGroup.GET("/collection/calendar", guessCollCal)
			guessGroup.GET("/collection/filter", guessCollGS)
			guessGroup.GET("/collection/question", authn.Guest, guessCollQes)
			guessGroup.GET("/collection/statis", authn.User, GuessCollStatis)
			guessGroup.GET("/collection/record", authn.User, guessCollRecord)
			guessGroup.GET("/match/recent", guessTeamRecent)
			guessGroup.GET("/match/record", authn.User, guessMatchRecord)
			guessGroup.GET("/act/result", authn.User, S9Result)
			guessGroup.GET("/act/record", authn.User, S9Record)
		}
		gsGroup := group.Group("/game")
		{
			gsGroup.GET("/rank", gameRank)
			gsGroup.GET("/season", gameSeason)
		}
	}
}

func internalRouter(e *bm.Engine) {
	group := e.Group("/x/internal/esports")
	{
		group.GET("/season", vfySvr.Verify, season)
		group.GET("/app/season", vfySvr.Verify, appSeason)
		matGroup := group.Group("/matchs")
		{
			matGroup.GET("/filter", vfySvr.Verify, filterMatch)
			matGroup.GET("/list", vfySvr.Verify, listContest)
			matGroup.GET("/app/list", vfySvr.Verify, appContest)
			matGroup.GET("/calendar", vfySvr.Verify, calendar)
			matGroup.GET("/active", vfySvr.Verify, active)
			matGroup.GET("/videos", vfySvr.Verify, actVideos)
			matGroup.GET("/points", vfySvr.Verify, actPoints)
			matGroup.GET("/top", vfySvr.Verify, actTop)
			matGroup.GET("/knockout", vfySvr.Verify, actKnockout)
			matGroup.GET("/info", vfySvr.Verify, contest)
			matGroup.GET("/recent", vfySvr.Verify, recent)
			matGroup.GET("/intervene", vfySvr.Verify, intervene)
		}
		livedataGroup := group.Group("/livedata")
		{
			livedataGroup.GET("/battle/list", battleList)
			livedataGroup.GET("/battle/info", battleInfo)
		}
		videoGroup := group.Group("/video")
		{
			videoGroup.GET("/filter", vfySvr.Verify, filterVideo)
			videoGroup.GET("/list", vfySvr.Verify, listVideo)
			videoGroup.GET("/search", vfySvr.Verify, search)
		}
		favGroup := group.Group("/fav")
		{
			favGroup.GET("", vfySvr.Verify, listFav)
			favGroup.GET("/season", vfySvr.Verify, seasonFav)
			favGroup.GET("/stime", vfySvr.Verify, stimeFav)
			favGroup.GET("/list", vfySvr.Verify, appListFav)
		}
		pointGroup := group.Group("/leida")
		{
			pointGroup.GET("/game", vfySvr.Verify, game)
			pointGroup.GET("/game/types", vfySvr.Verify, types)
			pointGroup.GET("/items", vfySvr.Verify, items)
			pointGroup.GET("/heroes", vfySvr.Verify, heroes)
			pointGroup.GET("/abilities", vfySvr.Verify, abilities)
			pointGroup.GET("/players", vfySvr.Verify, players)
			pointGroup.GET("/teams", vfySvr.Verify, teams)
			pointGroup.GET("/seasons", vfySvr.Verify, seasons)
			pointGroup.GET("/roles", vfySvr.Verify, roles)
		}
		bigGroup := group.Group("/stats")
		{
			bigGroup.GET("/player", vfySvr.Verify, bigPlayers)
			bigGroup.GET("/team", vfySvr.Verify, bigTeams)
		}
		guessGroup := group.Group("/guess")
		{
			guessGroup.GET("", vfySvr.Verify, guessDetail)
			guessGroup.GET("/moreshow", vfySvr.Verify, guessMoreShow)
			guessGroup.GET("/coin", vfySvr.Verify, guessDetailCoin)
			guessGroup.GET("/collection/calendar", vfySvr.Verify, guessCollCal)
			guessGroup.GET("/collection/filter", vfySvr.Verify, guessCollGS)
			guessGroup.GET("/collection/question", vfySvr.Verify, guessCollQes)
			guessGroup.GET("/collection/statis", vfySvr.Verify, GuessCollStatis)
			guessGroup.GET("/collection/record", vfySvr.Verify, guessCollRecord)
			guessGroup.GET("/match/recent", vfySvr.Verify, guessTeamRecent)
			guessGroup.GET("/match/record", vfySvr.Verify, guessMatchRecord)
			guessGroup.GET("/act/result", vfySvr.Verify, S9Result)
			guessGroup.GET("/act/record", vfySvr.Verify, S9Record)
		}
	}
}

func ping(c *bm.Context) {
	if err := eSvc.Ping(c); err != nil {
		log.Error("esports interface ping error")
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
