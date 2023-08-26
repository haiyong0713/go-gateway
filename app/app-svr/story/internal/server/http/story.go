package http

import (
	"net/http"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-card/interface/model/card/story"
	"go-gateway/app/app-svr/story/internal/model"
	gateecode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	_headerBuvid    = "Buvid"
	_headerDeviceID = "Device-ID"
)

func feedStory(ctx *bm.Context) {
	var (
		mid int64
	)
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	dvcid := header.Get(_headerDeviceID)
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &model.StoryParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	if param.TeenagersMode == 1 {
		ctx.JSON(nil, gateecode.StoryTeenagersModeErr)
		return
	}
	if param.From == 0 {
		param.From = 7
	}
	if param.FromSpmid == "" {
		param.FromSpmid = "tm.recommend.0.0"
	}
	if param.Spmid == "" {
		param.Spmid = "main.ugc-video-detail-vertical.0.0"
	}
	if param.Bvid != "" {
		// 优先使用bvid并转化aid
		aid, err := bvid.BvToAv(param.Bvid)
		if err != nil {
			ctx.JSON(nil, ecode.RequestErr)
			return
		}
		param.AID = aid
	}
	plat := model.Plat(param.MobiApp, param.Device)
	now := time.Now()
	data, ai, config, respCode := svc.StorySvc.Story(ctx, plat, buvid, mid, param, now)
	ctx.JSON(struct {
		Items  []*story.Item      `json:"items"`
		Config *story.StoryConfig `json:"config"`
	}{Items: data, Config: config}, nil)
	if ai != nil {
		svc.StorySvc.StoryInfoc(ctx, "/x/v2/feed/index/story", buvid, mid, plat, param.Pull, data, respCode, param.Build, dvcid, param.Network, param.TrackID, ai.UserFeature, param.DisplayID, param.AID, now, param)
	}
	if param.DisplayID == 1 { // 首次进入story播放页时上报
		svc.StorySvc.StoryClickInfoc(ctx, "/x/v2/feed/index/story", buvid, mid, param.AID, plat, param.Build,
			param.DisplayID, param.From, param.AutoPlay, param.TrackID, param.FromSpmid,
			param.Spmid, now)
	}
}

func spaceStory(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &model.SpaceStoryParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	param.Buvid = buvid
	param.Plat = model.Plat(param.MobiApp, param.Device)
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	ctx.JSON(svc.StorySvc.SpaceStory(ctx, param))
}

func spaceStoryCursor(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &model.SpaceStoryCursorParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	param.Buvid = buvid
	param.Plat = model.Plat(param.MobiApp, param.Device)
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	if param.Mid == 0 && param.Buvid == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	out, err := svc.StorySvc.SpaceStoryCursor(ctx, param)
	if err != nil {
		if svc.StorySvc.SLBRetry(err) {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			log.Error("Story cursor retry by: %+v", err)
			return
		}
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(out, nil)
}

func dynamicStory(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &story.DynamicStoryParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	if param.DisplayID == 1 {
		param.Pull = 0
	}
	param.Buvid = buvid
	param.Plat = model.Plat(param.MobiApp, param.Device)
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	ctx.JSON(svc.StorySvc.DynamicStory(ctx, param))
}

func storyCart(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &model.StoryCartParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	param.Buvid = buvid
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	ctx.JSON(svc.StorySvc.StoryCart(ctx, param))
}

func storyGameStatus(ctx *bm.Context) {
	param := &model.StoryGameParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	ctx.JSON(svc.StorySvc.StoryGameStatus(ctx, param))
}
