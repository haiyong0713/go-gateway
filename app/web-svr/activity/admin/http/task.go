package http

import (
	bm "go-common/library/net/http/blademaster"
	reqbinding "go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/activity/admin/model/task"
)

func taskList(c *bm.Context) {
	v := new(struct {
		BusinessID int64 `form:"business_id"`
		ForeignID  int64 `form:"foreign_id"`
		Pn         int64 `form:"pn" default:"1" validate:"min=1"`
		Ps         int64 `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := taskSrv.TaskList(c, v.BusinessID, v.ForeignID, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func addTask(c *bm.Context) {
	v := new(task.AddArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, taskSrv.AddTask(c, v))
}

func addTaskV2(c *bm.Context) {
	v := &task.AddArgV2{}
	if err := c.BindWith(v, reqbinding.JSON); err != nil {
		return
	}
	c.JSON(nil, taskSrv.AddTaskV2(c, v))
}

func saveTask(c *bm.Context) {
	v := new(task.SaveArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, taskSrv.SaveTask(c, v))
}

func addAward(c *bm.Context) {
	v := new(task.AddAward)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, taskSrv.AddAward(c, v))
}
