package http

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/service"
	innerSql "go-gateway/app/web-svr/esports/job/sql"
)

var epSrv *service.Service

// Init init
func Init(c *conf.Config, s *service.Service) {
	epSrv = s
	engine := bm.DefaultServer(c.BM)
	router(engine)
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func router(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)
	e.GET("/big/init", bigInit)
	e.GET("/big/info", infoInit)
	e.POST("/auto", auto)
	e.GET("/auto/history", autoHistory)
	e.GET("/auto/one", autoOne)
	e.GET("/auto/pass/all", autoPassAll)
	e.GET("/auto/pass/one", autoPassOne)
	e.GET("/live/offline", liveOffline)
	e.GET("/live/imageMap", imageMap)
	e.GET("/live/battleList", liveBattleList)
	e.GET("/live/battleInfo", liveBattleInfo)
	e.GET("/live/up", liveUpImage)
	e.GET("/archives/score/sync", archivesScoreSync)
	rankingData := e.Group("/s10/rankingData")
	{
		rankingData.GET("/sync", s10RankingDataSync)
		//rankingData.GET("/intervention", s10RankingDataIntervention)
	}
	e.GET("/season/statusM", service.SeasonNotifyStatusM)
	e.GET("/tunnel/push", tunnelPush)
	e.GET("/fix/contest/status/history", fixContestStatus)
}

func ping(c *bm.Context) {
	if err := innerSql.Ping(); err != nil {
		c.JSON(nil, err)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)
}

func bigInit(c *bm.Context) {
	p := new(struct {
		Tp  int64 `form:"tp"`
		Sid int64 `form:"sid"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, epSrv.BigInit(p.Tp, p.Sid))
}

func infoInit(c *bm.Context) {
	p := new(struct {
		Tp      string `form:"tp"`
		MatchID int64  `form:"match_id"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, epSrv.InfoInit(p.Tp, p.MatchID))
}

func auto(c *bm.Context) {
	p := new(struct {
		Msg string `form:"msg" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, epSrv.Auto(p.Msg))
}

func autoHistory(c *bm.Context) {
	if err := epSrv.NewAutoCheck(); err != nil {
		res := map[string]interface{}{}
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	go func() {
		epSrv.NewAutoTagHistoryArc()
	}()
	c.JSON("ok", nil)
}

func autoOne(c *bm.Context) {
	p := new(struct {
		Aid int64 `form:"aid" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, epSrv.NewAutoOneArc(p.Aid))
}

func autoPassAll(c *bm.Context) {
	go func() {
		epSrv.NewAutoCheckPass()
	}()
	c.JSON("ok", nil)
}

func autoPassOne(c *bm.Context) {
	p := new(struct {
		Aid int64 `form:"aid" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, epSrv.NewAutoOnePass(p.Aid))
}

func liveOffline(c *bm.Context) {
	go epSrv.OffLineImage(context.Background())
	c.JSON("ok", nil)
}

func imageMap(c *bm.Context) {
	epSrv.LoadLiveImageMap()
	c.JSON("ok", nil)
}

func liveBattleList(c *bm.Context) {
	p := new(struct {
		MatchID string `form:"match_id" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	epSrv.BattleListTwo(c, p.MatchID)
	c.JSON("ok", nil)
}

func liveBattleInfo(c *bm.Context) {
	p := new(struct {
		BattleString string `form:"battleString" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	// 更新内存中的图片
	epSrv.StoreOffLineImage(c)
	liveOffLineImage := epSrv.LoadLiveOffLineImageMap()
	epSrv.BattleInfoThree(c, p.BattleString, liveOffLineImage)
	c.JSON("ok", nil)
}

func liveUpImage(c *bm.Context) {
	p := new(struct {
		URL string `form:"url" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(epSrv.BattleUpImage(c, p.URL))
}

func s10RankingDataSync(c *bm.Context) {
	p := new(struct {
		RoundID string `form:"round_id"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if p.RoundID != "" {
		c.JSON(epSrv.SyncS10RankingData(p.RoundID), nil)
	} else {
		c.JSON(epSrv.SyncS10RankingData(), nil)
	}
}

func s10RankingDataIntervention(c *bm.Context) {
	p := new(service.S10RankingInterventionData)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON("", epSrv.S10RankingDataIntervention(c, p))
}

func tunnelPush(c *bm.Context) {
	p := new(struct {
		ContestID int64 `form:"contest_id" validate:"required"`
		Mid       int64 `form:"mid" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON("ok", epSrv.TunnelPush(c, p.ContestID, p.Mid))
}

func archivesScoreSync(c *bm.Context) {
	c.JSON("ok", epSrv.ArchivesScoreSync(c))
}

func fixContestStatus(c *bm.Context) {
	go func() {
		epSrv.FixContestStatus(c)
	}()
	c.JSON("ok", nil)
}
