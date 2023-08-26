package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

// 页面初始化 返回三个值
func pageInfo(ctx *bm.Context) {
	ctx.JSON(service.FunnySvc.PageInfo(ctx))
}

// 获取当前点赞数量
func getLikeCount(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.FunnySvc.Likes(ctx, midI))
}

// 增加一次抽奖次数
func incrDrawTime(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.FunnySvc.AddLotteryTimes(ctx, midI))
}
