package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/model"
)

func gameInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.GameInfo(c, v.ID))
}

func gameList(c *bm.Context) {
	var (
		list []*model.Game
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn    int64  `form:"pn" validate:"min=0"`
		Ps    int64  `form:"ps" validate:"min=0,max=30"`
		Name  string `form:"name"`
		State int64  `form:"state"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Pn == 0 {
		v.Pn = 1
	}
	if v.Ps == 0 {
		v.Ps = 20
	}
	if list, cnt, err = esSvc.GameList(c, v.Pn, v.Ps, v.Name); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func addGame(c *bm.Context) {
	v := new(model.Game)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.Plat != 0 {
		if _, ok := model.PlatMap[v.Plat]; !ok {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if v.Type != 0 {
		if _, ok := model.TypeMap[v.Type]; !ok {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	c.JSON(nil, esSvc.AddGame(c, v))
}

func editGame(c *bm.Context) {
	v := new(model.Game)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.ID <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if v.Plat != 0 {
		if _, ok := model.PlatMap[v.Plat]; !ok {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if v.Type != 0 {
		if _, ok := model.TypeMap[v.Type]; !ok {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	c.JSON(nil, esSvc.EditGame(c, v))
}

func forbidGame(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.ForbidGame(c, v.ID, v.State))
}

func types(c *bm.Context) {
	c.JSON(esSvc.Types(c))
}

func gameTeams(c *bm.Context) {
	var (
		list []*model.Team
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn    int64  `form:"pn" validate:"min=0" default:"1"`
		Ps    int64  `form:"ps" validate:"min=1" default:"30"`
		ID    int64  `form:"id" validate:"min=1"`
		Title string `form:"title"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.GameTeams(c, v.Pn, v.Ps, v.ID, v.Title); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"count": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func gameSeasons(c *bm.Context) {
	var (
		list []*model.Season
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn    int64  `form:"pn" validate:"min=0" default:"1"`
		Ps    int64  `form:"ps" validate:"min=1" default:"30"`
		ID    int64  `form:"id" validate:"min=0"`
		Title string `form:"title"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.GameSeasons(c, v.Pn, v.Ps, v.ID, v.Title); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"count": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}
