package http

import (
	"encoding/json"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/api"
)

func pollMeta(ctx *bm.Context) {
	req := &api.PollMetaReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(pollSvc.PollMeta(ctx, req))
}

func pollOptions(ctx *bm.Context) {
	req := &api.PollOptionsReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(pollSvc.PollOptions(ctx, req))
}

func pollVote(ctx *bm.Context) {
	req := &api.PollVoteReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(nil, pollSvc.PollVote(ctx, req))
}

func pollOptionStatTop(ctx *bm.Context) {
	req := &api.PollOptionStatTopReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(pollSvc.PollOptionStatTop(ctx, req))
}

func pollS9Vote(ctx *bm.Context) {
	req := &api.PollVoteReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	mid, ok := ctx.Get("mid")
	if !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	req.Mid = mid.(int64)
	vote := ctx.Request.Form.Get("vote")
	if err := json.Unmarshal([]byte(vote), &req.Vote); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, pollSvc.PollS9Vote(ctx, req))
}

func pollVoted(ctx *bm.Context) {
	req := &api.PollVotedReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	mid, ok := ctx.Get("mid")
	if !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	req.Mid = mid.(int64)
	ctx.JSON(pollSvc.PollVoted(ctx, req))
}

func pollMOptions(ctx *bm.Context) {
	req := &struct {
		PollID int64 `form:"poll_id" validate:"required"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(pollSvc.PollMOptions(ctx, req.PollID))
}

func pollMOptionsDelete(ctx *bm.Context) {
	req := &struct {
		PollOptionID int64 `form:"poll_option_id" validate:"required"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(nil, pollSvc.PollMOptionDelete(ctx, req.PollOptionID))
}

func pollMOptionsAdd(ctx *bm.Context) {
	req := &struct {
		PollID int64  `form:"poll_id" validate:"required"`
		Title  string `form:"title" validate:"required"`
		Image  string `form:"image" validate:"required"`
		Group  string `form:"group" validate:"required"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(nil, pollSvc.PollMOptionsAdd(ctx, req.PollID, req.Title, req.Image, req.Group))
}

func pollMOptionsUpdate(ctx *bm.Context) {
	req := &struct {
		PollOptionID int64  `form:"poll_option_id" validate:"required"`
		Title        string `form:"title" validate:"required"`
		Image        string `form:"image" validate:"required"`
		Group        string `form:"group" validate:"required"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(nil, pollSvc.PollMOptionsUpdate(ctx, req.PollOptionID, req.Title, req.Image, req.Group))
}

func pollM(ctx *bm.Context) {
	rmid, ok := ctx.Get("mid")
	if !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	mid := rmid.(int64)
	if !pollSvc.PollM(ctx, mid) {
		ctx.JSON(nil, ecode.AccessDenied)
		ctx.Abort()
	}
}
