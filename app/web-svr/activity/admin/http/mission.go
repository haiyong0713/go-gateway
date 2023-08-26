package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/client"
	"go-gateway/app/web-svr/activity/interface/api"
	"time"
)

func addMissionActivityRouter(group *bm.RouterGroup) {
	group.SetMethodConfig(&bm.MethodConfig{Timeout: xtime.Duration(time.Second * 5)})
	missionGroup := group.Group("/mission")
	{
		missionGroup.GET("/list", getList)
		missionGroup.GET("/info", getActivityInfo)
		missionGroup.POST("/status", changeStatus)
		missionGroup.POST("/save", saveActivity)
		missionGroup.GET("/tasks", getActivityTasks)
		missionGroup.POST("/task/del", delTask)
		missionGroup.POST("/task/save", saveMissionTask)
		missionGroup.GET("/task/info", taskInfo)
		missionGroup.GET("/stock/check", stockCheck)
	}
}

func getList(ctx *bm.Context) {
	v := new(struct {
		Pn int64 `form:"pn" validate:"required,min=1"`
		Ps int64 `form:"ps" validate:"required,max=100"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.GetMissionActivityList(ctx, &api.GetMissionActivityListReq{
		Pn: v.Pn,
		Ps: v.Ps,
	}))
}

func getActivityInfo(ctx *bm.Context) {
	v := new(struct {
		ActId int64 `form:"act_id" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.GetMissionActivityInfo(ctx, &api.GetMissionActivityInfoReq{
		ActId:     v.ActId,
		SkipCache: 1,
	}))
}

func changeStatus(ctx *bm.Context) {
	v := new(struct {
		ActId  int64 `json:"act_id" form:"act_id" validate:"required,min=1"`
		Status int64 `json:"status" form:"status"`
	})
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.ChangeMissionActivityStatus(ctx, &api.ChangeMissionActivityStatusReq{
		ActId:  v.ActId,
		Status: v.Status,
	}))
}

func saveActivity(ctx *bm.Context) {
	v := new(api.MissionActivityDetail)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.SaveMissionActivity(ctx, v))
}

func getActivityTasks(ctx *bm.Context) {
	v := new(struct {
		Id int64 `form:"id" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.GetMissionTasks(ctx, &api.GetMissionTasksReq{
		Id: v.Id,
	}))
}

func saveTasks(ctx *bm.Context) {
	v := new(api.SaveMissionTasksReq)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.SaveMissionTasks(ctx, v))
}

func delTask(ctx *bm.Context) {
	v := new(struct {
		ActId  int64 `json:"act_id" form:"act_id" validate:"required,min=1"`
		TaskId int64 `json:"task_id" form:"task_id" validate:"required,min=1"`
	})
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.DelMissionTask(ctx, &api.DelMissionTaskReq{
		ActId:  v.ActId,
		TaskId: v.TaskId,
	}))
}

func saveMissionTask(ctx *bm.Context) {
	v := new(api.MissionTaskDetail)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.SaveMissionTask(ctx, v))
}

func taskInfo(ctx *bm.Context) {
	v := new(struct {
		ActId  int64 `form:"act_id" validate:"required,min=1"`
		TaskId int64 `form:"task_id" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.GetMissionTaskInfo(ctx, &api.GetMissionTaskInfoReq{
		ActId:  v.ActId,
		TaskId: v.TaskId,
	}))
}

func stockCheck(ctx *bm.Context) {
	v := new(api.MissionCheckStockReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	resp, err := client.ActivityClient.MissionCheckStock(ctx, v)
	res := make(map[string]interface{})
	res["data"] = false
	if err == nil {
		res["data"] = resp.Status
	}
	ctx.JSONMap(res, err)
}
