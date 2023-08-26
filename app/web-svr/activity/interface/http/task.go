package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func taskList(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.ActTaskList(c, loginMid, v.Sid))
}

func doTask(c *bm.Context) {
	var mid int64
	v := new(struct {
		TaskID int64 `form:"task_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, service.LikeSvc.DoTask(c, mid, v.TaskID, true))
}

func internalDoTask(c *bm.Context) {
	v := new(struct {
		Mid    int64 `form:"mid" validate:"min=1"`
		TaskID int64 `form:"task_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.DoTask(c, v.Mid, v.TaskID, false))

}
func awardTask(c *bm.Context) {
	v := new(struct {
		TaskID int64 `form:"task_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.AwardTask(c, mid, v.TaskID))
}

func awardTaskSpecial(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.AwardTaskSpecial(c, mid, v.Sid))
}

func addAwardTask(c *bm.Context) {
	v := new(struct {
		TaskID     int64 `form:"task_id" validate:"min=1"`
		Mid        int64 `form:"mid"`
		AwardCount int64 `form:"award"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.AddAwardTask(c, v.Mid, v.TaskID, v.AwardCount))
}

func taskTokenDo(c *bm.Context) {
	v := new(struct {
		Sid   int64  `form:"sid" validate:"min=1"`
		Token string `form:"token" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.TaskTokenDo(c, v.Sid, mid, v.Token))
}

func cardNum(c *bm.Context) {
	v := new(struct {
		Sid string `form:"sid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.CardNum(c, v.Sid, mid))
}

func taskCheck(c *bm.Context) {
	v := new(struct {
		Sid   int64  `form:"sid" validate:"min=1"`
		Token string `form:"token" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.TaskCheck(c, v.Sid, v.Token))
}

func sendPoints(c *bm.Context) {
	v := new(struct {
		Business  string `form:"business"`
		Activity  string `form:"activity"`
		Timestamp int64  `form:"timestamp"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, service.TaskSvr.ActSend(c, mid, v.Business, v.Activity, v.Timestamp))
}

func taskResult(c *bm.Context) {
	v := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.TaskSvr.Result(c, mid, v.Activity))
}
