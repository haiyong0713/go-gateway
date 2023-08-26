package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/interface/model"
)

func liveMatchs(c *bm.Context) {
	p := new(struct {
		Mid  int64
		Cids []int64 `form:"cids,split" validate:"required,dive,gt=0"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.LiveMatchs(c, p.Mid, p.Cids))
}

func liveMatchsAct(c *bm.Context) {
	p := new(model.MatchLive)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.LiveMatchsAct(c, p))
}

func battleList(c *bm.Context) {
	p := new(struct {
		MatchID string `form:"match_id"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.ScoreBattleList(c, p.MatchID))
}

func battleInfo(c *bm.Context) {
	p := new(struct {
		BattleString string `form:"battleString" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.ScoreBattleInfo(c, p.BattleString))
}
