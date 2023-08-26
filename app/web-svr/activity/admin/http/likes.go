package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

func likesList(c *bm.Context) {
	arg := new(model.LikesParam)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.LikesList(c, arg))
}

func likes(c *bm.Context) {
	var args struct {
		Sid  int64   `form:"sid" validate:"min=1,required"`
		Lids []int64 `form:"lids,split" validate:"min=1,max=50,dive,min=1"`
	}
	if err := c.Bind(&args); err != nil {
		return
	}
	c.JSON(actSrv.Likes(c, args.Sid, args.Lids))
}

func addLike(c *bm.Context) {
	args := new(model.AddLikes)
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(actSrv.AddLike(c, args))
}

func upLike(c *bm.Context) {
	args := new(model.UpLike)
	if err := c.Bind(args); err != nil {
		return
	}
	username, _ := c.Get("username")
	c.JSON(actSrv.UpLike(c, args, username.(string)))
}

func upListContent(c *bm.Context) {
	args := new(model.UpReply)
	if err := c.Bind(args); err != nil {
		return
	}
	username, _ := c.Get("username")
	c.JSON(nil, actSrv.UpLikesState(c, args.IDs, args.State, args.Reply, username.(string)))
}

func upWid(c *bm.Context) {
	args := new(model.UpWid)
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(nil, actSrv.UpWid(c, args))
}

func addPic(c *bm.Context) {
	args := new(model.AddPic)
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(actSrv.AddPicContent(c, args))
}

func batchLikes(c *bm.Context) {
	args := new(model.BatchLike)
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(nil, actSrv.BatchLikes(c, args))
}

func likeExport(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1,required"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.LikeExport(c, arg.Sid))
}

func likeBatchEdit(ctx *bm.Context) {
	arg := new(struct {
		IDs    []int64 `form:"ids,split" validate:"max=20,dive,min=1"`
		Sid    int64   `form:"sid" validate:"min=1"`
		State  int64   `form:"state"`
		Reason string  `form:"reason"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	ctx.JSON(nil, actSrv.BatchEdit(ctx, arg.Sid, arg.State, arg.IDs, arg.Reason))
}
