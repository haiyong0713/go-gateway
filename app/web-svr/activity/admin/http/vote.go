package http

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/client"
	actEcode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"time"
)

func addInternalVoteRouter(group *bm.RouterGroup) {
	group.SetMethodConfig(&bm.MethodConfig{Timeout: xtime.Duration(time.Second * 5)})
	voteGroup := group.Group("/vote", authSrv.Permit2("ACTIVITY_VOTE_USER"))
	{
		voteGroup.POST("/activity/add", authSrv.Permit2("ACTVITY_VOTE_ADMIN"), VoteAddActivity)
		voteGroup.POST("/activity/update", authSrv.Permit2("ACTVITY_VOTE_ADMIN"), VoteUpdateActivity)
		voteGroup.POST("/activity/del", VoteDelActivity)
		voteGroup.GET("/activity/list", VoteListActivity)

		voteGroup.POST("/activity/rule/update", authSrv.Permit2("ACTVITY_VOTE_ADMIN"), VoteUpdateActivityRule)

		voteGroup.POST("/activity/datasource/add", VoteAddActivityDS)
		voteGroup.POST("/activity/datasource/del", VoteDelActivityDS)
		voteGroup.POST("/activity/datasource/update", VoteUpdateActivityDS)
		voteGroup.GET("/activity/datasource/list", VoteListActivityDS)

		voteGroup.POST("/intervene/update", VoteUpdateInterveneVoteCount)
		voteGroup.POST("/blacklist/add", VoteAddActivityBlackList)
		voteGroup.POST("/blacklist/del", VoteDelActivityBlackList)
		voteGroup.GET("/activity/rank", VoteGetActivityInternalRank)
		voteGroup.POST("/activity/rank/refresh", VoteRefreshActivityRank)
		voteGroup.POST("/activity/record/export", VoteExportActivityRecord)
		voteGroup.POST("/activity/rank/export", VoteExportActivityRank)

	}
}

func VoteAddActivity(ctx *bm.Context) {
	v := &api.AddVoteActivityReq{}
	if usernameCtx, ok := ctx.Get("username"); ok {
		v.Creator = usernameCtx.(string)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.AddVoteActivity(ctx, v))
}

func VoteDelActivity(ctx *bm.Context) {
	v := &api.DelVoteActivityReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.DelVoteActivity(ctx, v))
}

func VoteUpdateActivity(ctx *bm.Context) {
	v := &api.UpdateVoteActivityReq{}
	if usernameCtx, ok := ctx.Get("username"); ok {
		v.Editor = usernameCtx.(string)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.UpdateVoteActivity(ctx, v))
}

func VoteListActivity(ctx *bm.Context) {
	v := &api.ListVoteActivityReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.ListVoteActivity(ctx, v))
}

func VoteUpdateActivityRule(ctx *bm.Context) {
	v := &api.UpdateVoteActivityRuleReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.UpdateVoteActivityRule(ctx, v))
}

func VoteAddActivityDS(ctx *bm.Context) {
	v := &api.AddVoteActivityDataSourceGroupReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.AddVoteActivityDataSourceGroup(ctx, v))
}

func VoteDelActivityDS(ctx *bm.Context) {
	v := &api.DelVoteActivityDataSourceGroupReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.DelVoteActivityDataSourceGroup(ctx, v))
}

func VoteUpdateActivityDS(ctx *bm.Context) {
	v := &api.UpdateVoteActivityDataSourceGroupReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.UpdateVoteActivityDataSourceGroup(ctx, v))
}

func VoteListActivityDS(ctx *bm.Context) {
	v := &api.ListVoteActivityDataSourceGroupsReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.ListVoteActivityDataSourceGroups(ctx, v))
}

func VoteGetActivityInternalRank(ctx *bm.Context) {
	v := &api.GetVoteActivityRankInternalReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	res, err := client.ActivityClient.GetVoteActivityRankInternal(ctx, v)

	if ecode.EqualError(actEcode.ActivityVoteRankExpired, err) ||
		ecode.EqualError(actEcode.ActivityVoteDSGNotFound, err) {
		err = ecode.Error(ecode.ServerErr, "未找到数据, 若是新创建的活动/数据组请等待5分钟后查看. 如有问题可联系 @jinyang")

	}
	ctx.JSON(res, err)
}

func VoteUpdateInterveneVoteCount(ctx *bm.Context) {
	v := &api.UpdateVoteActivityInterveneVoteCountReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.UpdateVoteActivityInterveneVoteCount(ctx, v))
}

func VoteAddActivityBlackList(ctx *bm.Context) {
	v := &api.AddVoteActivityBlackListReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.AddVoteActivityBlackList(ctx, v))
}

func VoteDelActivityBlackList(ctx *bm.Context) {
	v := &api.DelVoteActivityBlackListReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.DelVoteActivityBlackList(ctx, v))
}

func VoteRefreshActivityRank(ctx *bm.Context) {
	v := &api.RefreshVoteActivityRankZsetReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	err := innerRefreshActivityRank(metadata.WithContext(ctx), v.ActivityId)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, "刷新排行失败, 重试三次后仍有问题可联系 @jinyang")
	}
	ctx.JSON(nil, err)
}

func innerRefreshActivityRank(ctx context.Context, id int64) (err error) {
	_, err = client.ActivityClient.RefreshVoteActivityRankZset(ctx, &api.RefreshVoteActivityRankZsetReq{ActivityId: id})
	if err != nil {
		log.Errorc(ctx, "VoteRefreshRank RefreshVoteActivityRankZset for activityId: %v error: %v", id, err)
		return
	} else {
		log.Errorc(ctx, "VoteRefreshRank RefreshVoteActivityRankZset for activityId: %v success", id)
	}
	var eg errgroup.Group
	eg.Go(func(ctx context.Context) (err error) {
		_, err = client.ActivityClient.RefreshVoteActivityRankExternal(ctx, &api.RefreshVoteActivityRankExternalReq{ActivityId: id})
		if err != nil {
			log.Errorc(ctx, "VoteRefreshRank RefreshVoteActivityRankExternal for activityId: %v error: %v", id, err)
		} else {
			log.Errorc(ctx, "VoteRefreshRank RefreshVoteActivityRankExternal for activityId: %v success", id)
		}
		return
	})

	eg.Go(func(ctx context.Context) (err error) {
		_, err = client.ActivityClient.RefreshVoteActivityRankInternal(ctx, &api.RefreshVoteActivityRankInternalReq{ActivityId: id})
		if err != nil {
			log.Errorc(ctx, "VoteRefreshRank RefreshVoteActivityRankInternal for activityId: %v error: %v", id, err)
		} else {
			log.Errorc(ctx, "VoteRefreshRank RefreshVoteActivityRankInternal for activityId: %v success", id)
		}
		return
	})

	return eg.Wait()
}

func VoteExportActivityRecord(ctx *bm.Context) {
	v := new(struct {
		ActivityId int64 `form:"activity_id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		ctx.JSON(nil, ecode.NoLogin)
		return
	}
	err := _fanOut.Do(ctx, func(ctx context.Context) {
		actSrv.ExportNewVoteDetail(ctx, userName, v.ActivityId)
	})

	ctx.JSON(nil, err)
}

func VoteExportActivityRank(ctx *bm.Context) {
	v := new(struct {
		GroupId int64 `form:"group_id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		ctx.JSON(nil, ecode.NoLogin)
		return
	}
	err := _fanOut.Do(ctx, func(ctx context.Context) {
		actSrv.ExportNewVoteRank(context.Background(), userName, v.GroupId)
	})
	ctx.JSON(nil, err)
}
