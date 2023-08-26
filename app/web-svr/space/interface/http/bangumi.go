package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/interface/model"
)

func bangumiList(c *bm.Context) {
	var (
		mid   int64
		list  []*model.Bangumi
		count int
		err   error
	)
	v := new(struct {
		Vmid int64 `form:"vmid" validate:"min=1"`
		Pn   int   `form:"pn" default:"1" validate:"min=1"`
		Ps   int   `form:"ps" default:"15" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if list, count, err = spcSvc.BangumiList(c, mid, v.Vmid, v.Pn, v.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	type data struct {
		List []*model.Bangumi `json:"list"`
		*model.Page
	}
	c.JSON(&data{List: list, Page: &model.Page{Pn: v.Pn, Ps: v.Ps, Total: count}}, nil)
}

func followList(c *bm.Context) {
	var (
		mid   int64
		list  []*model.FollowCard
		count int32
		err   error
	)
	v := new(struct {
		Vmid         int64 `form:"vmid" validate:"min=1"`
		Type         int32 `form:"type" validate:"required"`
		Pn           int32 `form:"pn" default:"1" validate:"min=1"`
		Ps           int32 `form:"ps" default:"15" validate:"min=1,max=30"`
		FollowStatus int32 `form:"follow_status"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Type != model.FollowTypeAnime && v.Type != model.FollowTypeCinema {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.FollowStatus != 0 {
		if v.FollowStatus != model.FollowStatusWant && v.FollowStatus != model.FollowStatusIng && v.FollowStatus != model.FollowStatusFinish {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	list, count, err = spcSvc.FollowList(c, mid, v.Vmid, v.Type, v.Pn, v.Ps, v.FollowStatus)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	type page struct {
		Pn    int32 `json:"pn"`
		Ps    int32 `json:"ps"`
		Total int32 `json:"total"`
	}
	type data struct {
		List []*model.FollowCard `json:"list"`
		*page
	}
	c.JSON(&data{List: list, page: &page{Pn: v.Pn, Ps: v.Ps, Total: count}}, nil)
}

func bangumiConcern(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	v := new(struct {
		SeasonID int64 `form:"season_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.BangumiConcern(c, mid, v.SeasonID))
}

func bangumiUnConcern(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	v := new(struct {
		SeasonID int64 `form:"season_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.BangumiUnConcern(c, mid, v.SeasonID))
}
