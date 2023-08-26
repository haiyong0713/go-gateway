package http

import (
	"time"

	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/esports/interface/model"
)

// guessMoreShow guess more show
func guessMoreShow(c *bm.Context) {
	var (
		list *model.MoreShow
		err  error
	)
	p := new(struct {
		HomeID int64 `form:"home_id" validate:"required"`
		AwayID int64 `form:"away_id" validate:"required"`
		SID    int64 `form:"sid"`
		CID    int64 `form:"cid" validate:"required"`
	})
	if err = c.Bind(p); err != nil {
		return
	}
	if list, err = eSvc.GuessMoreMatch(c, p.HomeID, p.AwayID, p.SID, p.CID); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(list, nil)
}

func userSeasonGuessList(ctx *bm.Context) {
	var (
		mid int64
	)
	req := new(model.UserSeasonGuessReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}

	req.MID = mid
	ctx.JSON(eSvc.FetchUserSeasonGuessList(ctx, req))
}

func userSeasonGuessSummary(ctx *bm.Context) {
	var (
		mid int64
	)
	p := new(struct {
		SeasonID int64 `form:"season_id"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}

	req := new(model.GuessParams4V2)
	{
		req.MID = mid
		req.SeasonID = p.SeasonID
	}

	ctx.JSON(eSvc.FetchSeasonGuessSummary(ctx, req))
}

// guessDetail guess act detail
func guessDetail(c *bm.Context) {
	var (
		mid int64
	)
	p := new(struct {
		Cid int64 `form:"cid" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(eSvc.GuessDetail(c, p.Cid, mid))
}

func guessList(c *bm.Context) {
	var (
		mid int64
	)
	p := new(struct {
		Cid int64 `form:"cid" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(eSvc.GuessListByContestID(c, p.Cid, mid))
}

// guessDetailCoin .
func guessDetailCoin(c *bm.Context) {
	var (
		mid int64
	)
	p := new(struct {
		MainID int64 `form:"main_id" validate:"required"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(eSvc.GuessDetailCoin(c, mid, p.MainID))
}

// guessDetailAdd .
func guessDetailAdd(c *bm.Context) {
	var (
		mid int64
	)
	p := &model.AddGuessParam{}
	if err := c.Bind(p); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	p.MID = mid
	c.JSON(nil, eSvc.AddGuessDetail(c, p))
}

// guessCollCal .
func guessCollCal(c *bm.Context) {
	p := &model.ParamContest{}
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.GuessCollCalendar(c, p))
}

// guessCollGS .
func guessCollGS(c *bm.Context) {
	p := &struct {
		Gid int64 `form:"gid" validate:"gte=0"`
	}{}
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.GuessCollGS(c, p.Gid))
}

// guessCollQes .
func guessCollQes(c *bm.Context) {
	var (
		err   error
		total int
		list  []*model.GuessCollQues
		mid   int64
	)
	p := &model.ParamContest{}
	if err = c.Bind(p); err != nil {
		return
	}
	if p.Stime != "" {
		if _, err = time.Parse("2006-01-02 15:04:05", p.Stime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if p.Etime != "" {
		if _, err = time.Parse("2006-01-02 15:04:05", p.Etime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if list, total, err = eSvc.GuessCollQues(c, p, mid); err != nil {
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

// GuessCollStatis .
func GuessCollStatis(c *bm.Context) {
	var (
		mid int64
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(eSvc.GuessCollStatis(c, mid))
}

// guessCollRecord .
func guessCollRecord(c *bm.Context) {
	var (
		mid int64
	)
	p := &model.GuessCollRecoParam{}
	if err := c.Bind(p); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	p.Mid = mid
	c.JSON(eSvc.GuessCollRecord(c, p))
}

// guessTeamRecent .
func guessTeamRecent(c *bm.Context) {
	var (
		list []*model.Contest
		err  error
	)
	p := &model.ParamEsGuess{}
	if err = c.Bind(p); err != nil {
		return
	}
	if list, err = eSvc.GuessTeamRecent(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(list, nil)
}

// guessMatchRecord .
func guessMatchRecord(c *bm.Context) {
	var (
		mid int64
	)
	p := &struct {
		CID int64 `form:"cid" validate:"required"`
	}{}
	if err := c.Bind(p); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(eSvc.GuessMatchRecord(c, mid, p.CID))
}

// S9Result .
func S9Result(c *bm.Context) {
	var (
		mid int64
	)
	p := &struct {
		SID int64 `form:"sid" validate:"required"`
	}{}
	if err := c.Bind(p); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(eSvc.S9Result(c, mid, p.SID))
}

// S9Record .
func S9Record(c *bm.Context) {
	var (
		mid int64
	)
	p := &struct {
		SID int64 `form:"sid" validate:"required"`
		Pn  int64 `form:"pn" default:"1" validate:"min=1"`
		Ps  int64 `form:"ps" default:"50" validate:"gt=0,lte=50"`
	}{}
	if err := c.Bind(p); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(eSvc.S9Record(c, mid, p.SID, p.Pn, p.Ps))
}
