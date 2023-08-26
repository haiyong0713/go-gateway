package http

import (
	"encoding/json"

	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/dance-taiko/interface/ecode"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

func gameStat(c *bm.Context) {
	v := new(struct {
		GameID int64  `form:"game_id" validate:"required"`
		MID    int64  `form:"mid"`
		Stats  string `form:"stats" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}

	header := c.Request.Header
	buvid := header.Get("Buvid")

	sts := make([]*model.Stat, 0)
	if err := json.Unmarshal([]byte(v.Stats), &sts); err != nil {
		c.JSON(nil, ecode.JsonFormatErr)
		return
	}
	c.JSON(nil, svc.GameStat(c, buvid, v.GameID, v.MID, sts))
}

func arcTopRank(c *bm.Context) {
	v := new(struct {
		AID    int64 `form:"aid" validate:"required"`
		Number int64 `form:"number" default:"20"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.ArcTopRanks(c, v.AID, v.Number))
}
