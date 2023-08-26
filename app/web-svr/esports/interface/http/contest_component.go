package http

import (
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
)

func timeContests(c *bm.Context) {
	var (
		mid int64
	)
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(eSvc.TimeContests(c, mid, v.Sid))
}

func allContests(c *bm.Context) {
	var (
		mid   int64
		list  []*pb.ContestCardComponent
		total int
		err   error
	)
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamAllContest)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.AllContests(c, mid, p); err != nil {
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

func allFold(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamAllFold)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.AllFoldContests(c, mid, p))
}

func abstract(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamAbstract)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.AbstractContests(c, mid, p))
}

func seasonContests(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamSeasonContests)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.SeasonContests(c, mid, p))
}

func battleContests(c *bm.Context) {
	var (
		mid   int64
		list  []*model.ContestBattleCardComponent
		total int
		err   error
	)
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamContestBattle)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.BattleContests(c, mid, p); err != nil {
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

func battleContestTeams(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamBattleTeams)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.BattleContestTeams(c, mid, p))
}

func teamContests(c *bm.Context) {
	var (
		mid   int64
		list  []*pb.ContestCardComponent
		total int
		err   error
	)
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamTeamContest)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.TeamContests(c, mid, p); err != nil {
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

func teamContestsV2(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamV2TeamContest)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.TeamContestsV2(c, mid, p))
}

func howeAwayContests(c *bm.Context) {
	param := &model.ParamEsGuess{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.HomeAwayContest(c, param))
}

func seasonTeamsComponent(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(eSvc.SeasonTeamsComponent(c, v.Sid))
}

func videoListComponent(c *bm.Context) {
	param := &model.ParamVideoList{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.TopicVideoListComponent(c, param))
}

func contestReplyWall(c *bm.Context) {
	param := &model.ParamWall{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.ContestReplyWall(c, param))
}
