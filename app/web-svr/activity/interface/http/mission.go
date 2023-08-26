package http

import (
	"encoding/json"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/mission"
	"go-gateway/app/web-svr/activity/interface/service"
	"go-gateway/app/web-svr/activity/interface/tool"
	"strconv"
	"time"
)

func addExternalMissionRouter(group *bm.RouterGroup) {
	rewardsGroup := group.Group("/mission")
	{
		rewardsGroup.GET("/info", missionBaseInfo)
		rewardsGroup.GET("/tasks", authSvc.Guest, missionTask)
		rewardsGroup.POST("/task/reward/receive", authSvc.User, rewardReceive)
		rewardsGroup.GET("/liveRoom", authSvc.Guest, MissionGetLiveRooms)
		rewardsGroup.GET("/videos", MissionGetVideos)
		rewardsGroup.GET("/articles", MissionGetVideos)
	}
	group.GET("/rewards/callback/tencent", TencentAwardCallback)
}

func missionBaseInfo(ctx *bm.Context) {
	v := new(struct {
		ActId int64 `form:"act_id"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.MissionActivitySvr.GetMissionActivityInfo(ctx, &api.GetMissionActivityInfoReq{
		ActId:     v.ActId,
		SkipCache: 0,
	}))
}

func missionTask(ctx *bm.Context) {
	v := new(struct {
		ActId int64 `form:"act_id"`
	})
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.MissionActivitySvr.GetUserTasks(ctx, v.ActId, mid))
}

func rewardReceive(ctx *bm.Context) {
	v := new(struct {
		ActId     int64 `form:"act_id" validate:"min=1"`
		TaskId    int64 `form:"task_id" validate:"min=1"`
		GroupId   int64 `form:"group_id"`
		ReceiveId int64 `form:"receive_id" validate:"min=1"`
	})
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.MissionActivitySvr.ReceiveTaskAward(ctx, mid, v.ActId, v.TaskId, v.ReceiveId))
}

func MissionGetLiveRooms(ctx *bm.Context) {
	v := new(struct {
		ActId        int64  `form:"id"`
		Platform     string `form:"platform"`
		Cursor       int64  `form:"cursor"`
		NetworkState int64  `form:"networkstate"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.MissionActivitySvr.FetchLiveRoomByOperId(ctx, v.ActId, v.NetworkState, ctx.RoutePath))
}

func MissionGetVideos(ctx *bm.Context) {
	v := new(struct {
		ActId    int64  `form:"id"`
		Platform string `form:"platform"`
		Cursor   int64  `form:"cursor"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.MissionActivitySvr.FetchVideoByOperId(ctx, v.ActId))
}

func TencentAwardCallback(ctx *bm.Context) {
	params := new(mission.CommonOriginRequest)
	if err := ctx.Bind(params); err != nil {
		return
	}

	data := map[string]interface{}{
		"isPass": false,
	}
	secretStruct := conf.Conf.AppSecrets[params.AppKey]
	if secretStruct == nil || secretStruct.Secret == "" {
		err := xecode.Error(xecode.RequestErr, "公钥不合法")
		ctx.JSONMap(data, err)
		return
	}
	now := time.Now().Unix() //校验时间戳, 只接受近5分钟的请求
	if params.Timestamp > now+300 || params.Timestamp < now-300 {
		err := xecode.Error(xecode.RequestErr, "时间戳不合法")
		ctx.JSONMap(data, err)
		return
	}

	checkParams := make(map[string]string)
	checkParams["app_key"] = params.AppKey
	checkParams["timestamp"] = strconv.FormatInt(params.Timestamp, 10)
	checkParams["request_id"] = params.RequestId
	checkParams["version"] = params.Version
	checkParams["params"] = params.Params

	err := tool.TencentSignCheck(ctx, params.AppKey, secretStruct.Secret, params.Sign, checkParams)
	if err != nil {
		err = xecode.Error(xecode.RequestErr, "签名不合法")
		ctx.JSONMap(data, err)
		return
	}
	innerParams := new(mission.TencentAwardCallBackInnerParams)
	err = json.Unmarshal([]byte(params.Params), innerParams)
	if err != nil {
		err = xecode.Error(xecode.RequestErr, "参数不合法")
		ctx.JSONMap(data, err)
		return
	}
	taskId, err := strconv.ParseInt(innerParams.TaskId, 10, 64)
	if err != nil {
		err = xecode.Error(xecode.RequestErr, "参数不合法")
		ctx.JSONMap(data, err)
		return
	}
	pass, err := service.MissionActivitySvr.CheckTencentGameAward(ctx, taskId, innerParams.UserId, innerParams.SerialNum)
	if err != nil {
		log.Errorc(ctx, "service.MissionActivitySvr.CheckTencentGameAward error: %+v", err)
	}
	data["isPass"] = pass
	ctx.JSONMap(data, err)
}
