package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/interface/model"
)

func game(c *bm.Context) {
	p := new(model.ParamGame)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Game(c, p))
}
func types(c *bm.Context) {
	c.JSON(eSvc.Types(c))
}
func roles(c *bm.Context) {
	p := new(struct {
		Tp string `form:"tp"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Roles(c, p.Tp))
}

func items(c *bm.Context) {
	p := new(model.ParamLeidas)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Items(c, p))
}

func heroes(c *bm.Context) {
	p := new(model.ParamLeidas)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Heroes(c, p))
}

func abilities(c *bm.Context) {
	p := new(model.ParamLeidas)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Abilities(c, p))
}
func players(c *bm.Context) {
	p := new(model.ParamLeidas)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Players(c, p))
}

func teams(c *bm.Context) {
	p := new(model.ParamLeidas)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.Teams(c, p))
}

func seasons(c *bm.Context) {
	var (
		total int
		list  []*model.Season
		err   error
	)
	p := new(struct {
		Tp     int64 `form:"tp"`
		TeamID int64 `form:"team_id"`
		Pn     int   `form:"pn" default:"1"  validate:"min=1"`
		Ps     int   `form:"ps" default:"50" validate:"gt=0,lte=50"`
	})
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.LeidaSeasons(c, p.Tp, p.TeamID, p.Pn, p.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func bigPlayers(c *bm.Context) {
	var (
		err          error
		list, header interface{}
		total        int
	)
	p := new(model.StatsBig)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, header, total, err = eSvc.BigPlayers(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 3)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	data["header"] = header
	c.JSON(data, nil)
}

func bigTeams(c *bm.Context) {
	var (
		err          error
		list, header interface{}
		total        int
	)
	p := new(model.StatsBig)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, header, total, err = eSvc.BigTeams(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 3)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	data["header"] = header
	c.JSON(data, nil)
}

func specialTeams(c *bm.Context) {
	var (
		err   error
		list  interface{}
		total int
	)
	p := new(model.ParamSpecTeams)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.SpecialTeams(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func specTeam(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	p := new(model.ParamSpecial)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.SpecTeam(c, mid, p))
}

func specPlayer(c *bm.Context) {
	p := new(model.ParamSpecial)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.SpecPlayer(c, p))
}

func playerRecent(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	p := new(model.ParamRecent)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.PlayerRecent(c, mid, p))
}

func mvpRank(ctx *bm.Context) {
	p := new(model.ParamMvpRank)
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(eSvc.PlayerMvpRank(ctx, p))
}

func kdaRank(ctx *bm.Context) {
	p := new(model.ParamKdaRank)
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(eSvc.PlayerKdaRank(ctx, p))
}

func hero2Rank(ctx *bm.Context) {
	p := new(model.ParamHero2Rank)
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(eSvc.Hero2Rank(ctx, p))
}
