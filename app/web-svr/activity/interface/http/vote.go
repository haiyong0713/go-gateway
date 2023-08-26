package http

import (
	bm "go-common/library/net/http/blademaster"
	riskModel "go-gateway/app/web-svr/activity/interface/model/risk"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"go-gateway/app/web-svr/activity/interface/service"
)

func addExternalVoteRouter(group *bm.RouterGroup) {
	rewardsGroup := group.Group("/vote_new")
	{
		rewardsGroup.GET("/rank", authSvc.Guest, VotesRank)
		rewardsGroup.POST("/do", authSvc.User, VotesDo)
		rewardsGroup.POST("/undo", authSvc.User, VotesUndo)
		rewardsGroup.GET("/search", authSvc.Guest, VoteSearch)
	}
}

func VotesRank(ctx *bm.Context) {
	v := &model.RankExternalParams{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}

	ctx.JSON(service.VoteSvr.GetVoteActivityRankExternal(ctx, mid, v))
}

func VotesDo(ctx *bm.Context) {
	v := &model.DoVoteParams{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	riskBase := riskParseUA(ctx, mid, riskModel.ActionVoteNew)
	ctx.JSON(service.VoteSvr.DoVote(ctx, mid, riskBase, v))
}

func VotesUndo(ctx *bm.Context) {
	v := &model.UndoVoteParams{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.VoteSvr.UndoVote(ctx, mid, v))
}

func VoteSearch(ctx *bm.Context) {
	v := &model.RankSearchParams{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	midStr, ok := ctx.Get("mid")
	if ok {
		mid = midStr.(int64)
	}
	ctx.JSON(service.VoteSvr.Search(ctx, mid, v))
}
