package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/space/interface/model"
)

func setTopDynamic(c *bm.Context) {
	c.JSON(nil, ecode.RequestErr)
	//v := new(struct {
	//	DyID int64 `form:"dy_id" validate:"min=1"`
	//})
	//if err := c.Bind(v); err != nil {
	//	return
	//}
	//midStr, _ := c.Get("mid")
	//mid := midStr.(int64)
	//c.JSON(nil, spcSvc.SetTopDynamic(c, mid, v.DyID))
}

func cancelTopDynamic(c *bm.Context) {
	c.JSON(nil, ecode.RequestErr)
	//midStr, _ := c.Get("mid")
	//mid := midStr.(int64)
	//c.JSON(nil, spcSvc.CancelTopDynamic(c, mid, time.Now()))
}

func dynamicList(c *bm.Context) {
	v := new(model.DyListArg)
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		v.Mid = midInter.(int64)
	}
	c.JSON(spcSvc.DynamicList(c, v))
}

func behaviorList(c *bm.Context) {
	var mid int64
	v := new(struct {
		Vmid     int64 `form:"mid" validate:"min=1"`
		LastTime int64 `form:"last_time"`
		Ps       int   `form:"ps" default:"20" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data := spcSvc.BehaviorList(c, mid, v.Vmid, v.LastTime, v.Ps)
	if len(data) == 0 {
		data = make([]*model.DyItem, 0)
	}
	c.JSON(data, nil)
}

func dynamicSearch(c *bm.Context) {
	v := &model.DynamicSearchArg{}
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	searchRes, err := spcSvc.DynamicSearch(c, mid, v)
	if err != nil {
		if !ecode.EqualError(ecode.NothingFound, err) {
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(searchRes, nil)
}
