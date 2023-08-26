package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func upArtLists(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.UpArticleLists(c, mid))
}

func addArtList(c *bm.Context) {
	v := new(struct {
		ListID int64 `form:"list_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.AddArtList(c, mid, v.ListID))
}

func giantArticleList(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	list, leftTimes, winTimes, err := service.LikeSvc.ArticleGiantV4List(c, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{
		"list":       list,
		"left_times": leftTimes,
		"win_times":  winTimes,
	}, nil)
}

func giantArticleChoose(c *bm.Context) {
	v := new(struct {
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	scores, err := service.LikeSvc.GiantChoose(c, mid, v.Lid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(scores, nil)
}

func awardSubjectStateByID(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	state, err := service.LikeSvc.AwardSubjectStateByID(c, v.ID, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{
		"state": state,
	}, nil)
}

func rewardSubjectByID(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.AwardSubjectRewardByID(c, v.ID, mid))
}
