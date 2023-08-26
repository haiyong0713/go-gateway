package http

import bm "go-common/library/net/http/blademaster"

func guessSeasons(c *bm.Context) {
	v := new(struct {
		Date string `form:"date"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	res, err := esSvc.GetValidSeasonsByDate(c, v.Date)
	if err != nil {
		return
	}
	dataMap := map[string]interface{}{
		"status": 0,
		"msg":    "",
		"data":   res,
	}
	c.JSONMap(dataMap, nil)
}

func guessContests(c *bm.Context) {
	v := new(struct {
		Date     string `form:"date"`
		SeasonId int64  `form:"season_id" validate:"min=1"`
		Mid      int64  `form:"mid"  validate:"min=1"`
		JustJoin bool   `form:"just_join"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	res, err := esSvc.GetSeasonDateContestGuess(c, v.Date, v.SeasonId, v.Mid, v.JustJoin)
	if err != nil {
		return
	}
	dataMap := map[string]interface{}{
		"status": 0,
		"msg":    "",
		"data":   res,
	}
	c.JSONMap(dataMap, nil)
}
