package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func schoolList(c *bm.Context) {
	c.JSON(service.LikeSvc.BdfSchoolList(c))
}

func schoolArcs(c *bm.Context) {
	v := new(struct {
		Lids []int64 `form:"lids,split" validate:"min=1,max=15,dive,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.BdfSchoolArcs(c, v.Lids))
}
