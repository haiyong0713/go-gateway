package http

import (
	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"
)

func seasonInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.SeasonInfo(c, v.ID))
}

func bigFix(c *bm.Context) {
	var err error
	v := new(struct {
		Tp  int64 `form:"tp" validate:"min=1"`
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if err = esSvc.BigFix(c, v.Tp, v.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修复失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON("ok", nil)
}

func seasonList(c *bm.Context) {
	var (
		list []*model.SeasonInfo
		cnt  int64
		err  error
	)
	v := new(struct {
		Mid   int64  `form:"mid"`
		Gid   int64  `form:"gid"`
		Pn    int64  `form:"pn" validate:"min=0"`
		Ps    int64  `form:"ps" validate:"min=0,max=30"`
		Title string `form:"title"`
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
	if list, cnt, err = esSvc.SeasonList(c, v.Mid, v.Pn, v.Ps, v.Gid, v.Title); err != nil {
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

func addSeason(c *bm.Context) {
	var (
		err           error
		tmpGids, gids []int64
	)
	v := new(model.Season)
	if err = c.Bind(v); err != nil {
		return
	}
	gidStr := c.Request.Form.Get("gids")
	if gidStr == "" {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if tmpGids, err = xstr.SplitInts(gidStr); err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	for _, v := range tmpGids {
		if v > 0 {
			gids = append(gids, v)
		}
	}
	if len(gids) == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.AddSeason(c, v, gids))
}

func editSeason(c *bm.Context) {
	var (
		err           error
		tmpGids, gids []int64
	)
	v := new(model.Season)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.ID <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	gidStr := c.Request.Form.Get("gids")
	if gidStr == "" {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if tmpGids, err = xstr.SplitInts(gidStr); err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	for _, v := range tmpGids {
		if v > 0 {
			gids = append(gids, v)
		}
	}
	if len(gids) == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.EditSeason(c, v, gids))
}

func forbidSeason(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.ForbidSeason(c, v.ID, v.State))
}

func forbidRankSeason(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.ForbidRankSeason(c, v.ID, v.State))
}

func rankInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.RankInfo(c, v.ID))
}

func rankList(c *bm.Context) {
	var (
		list []*model.SeasonInfo
		cnt  int64
		err  error
	)
	v := new(struct {
		Gid int64 `form:"gid" validate:"min=0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.RankList(c, v.Gid); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"count": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func addSeasonRank(c *bm.Context) {
	var (
		err error
	)
	v := new(model.SeasonRank)
	if err = c.Bind(v); err != nil {
		return
	}
	if err := esSvc.AddSeasonRank(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "赛季优先级创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editSeasonRank(c *bm.Context) {
	v := new(model.SeasonRank)
	if err := c.Bind(v); err != nil {
		return
	}
	if err := esSvc.EditSeasonRank(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "赛季优先级修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func AddTeamToSeason(c *bm.Context) {
	v := new(model.TeamInSeasonParam)
	if err := c.Bind(v); err != nil {
		return
	}
	if err := esSvc.AddTeamToSeason(c, v.Tid, v.Sid, v.Rank); err != nil {
		res := map[string]interface{}{}
		res["message"] = "向赛季中添加战队失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func RemoveTeamFromSeason(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Tid int64 `form:"tid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if err := esSvc.RemoveTeamFromSeason(c, v.Tid, v.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "从赛季中删除战队失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func UpdateTeamInSeason(c *bm.Context) {
	v := new(model.TeamInSeasonParam)
	if err := c.Bind(v); err != nil {
		return
	}
	if err := esSvc.UpdateTeamInSeason(c, v.Tid, v.Sid, v.Rank); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修改赛季中的队伍失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func ListTeamInSeason(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.ListTeamInSeason(c, v.Sid))
}

func RebuildTeamInSeason(c *bm.Context) {
	c.JSON(esSvc.RebuildTeamInSeasonInBackground(c))
}
