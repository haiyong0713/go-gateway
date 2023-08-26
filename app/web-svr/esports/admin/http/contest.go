package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"
)

const _special = 1

func contestInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.ContestInfo(c, v.ID))
}

func contestList(c *bm.Context) {
	var (
		list []*model.ContestInfo
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn        int64 `form:"pn" validate:"min=0"`
		Ps        int64 `form:"ps" validate:"min=0,max=50"`
		Mid       int64 `form:"mid"`
		Sid       int64 `form:"sid"`
		Sort      int64 `form:"sort"`
		Teamid    int64 `form:"teamid"`
		GuessType int64 `form:"guess_type"  default:"-1"`
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
	if list, cnt, err = esSvc.ContestList(c, v.Mid, v.Sid, v.Pn, v.Ps, v.Sort, v.Teamid, v.GuessType); err != nil {
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

func addContest(c *bm.Context) {
	var (
		err           error
		tmpGids, gids []int64
	)
	v := new(model.Contest)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Special == model.ContestSpecial && v.SpecialName == "" {
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
	if v.DataType > 0 && v.MatchID == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if err = esSvc.SaveContestByGrpc(c, v, gids); err != nil {
		res := map[string]interface{}{}
		res["message"] = "添加赛程失败 " + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editContest(c *bm.Context) {
	var (
		err           error
		tmpGids, gids []int64
	)
	v := new(model.Contest)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.ID <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if v.Special == model.ContestSpecial && v.SpecialName == "" {
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
	if v.DataType > 0 && v.MatchID == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if err = esSvc.UpdateCheck(c, v); err != nil {
		c.JSON(nil, err)
		return
	}
	if err = esSvc.SaveContestByGrpc(c, v, gids); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修改赛程失败 " + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func forbidContest(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if err := esSvc.FreezeCheck(c, v.State, v.ID); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, esSvc.ForbidContest(c, v.ID, v.State))
}

func matchFix(c *bm.Context) {
	var err error
	v := new(struct {
		MatchID int64 `form:"match_id" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if err = esSvc.MatchFix(c, v.MatchID); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修复失败 " + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	c.JSON("ok", nil)
}
