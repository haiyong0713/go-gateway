package http

import (
	bm "go-common/library/net/http/blademaster"
)

func ListTeamBySeason(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(eSvc.GetTeamsInSeason(c, v.Sid))
}
