package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/model"
)

func playURL(ctx *bm.Context) {
	params := &model.PlayURLParam{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	//codecid=2则为h265
	switch params.PreferCodecType {
	case model.CodeType_CODE264:
		params.PreferCodecID = model.CodeH264
	case model.CodeType_CODE265:
		params.PreferCodecID = model.CodeH265
	default:
		// 默认返回264,和视频云确认，所有dash类视频都会有264
		params.PreferCodecID = model.CodeH264
	}
	ctx.JSON(svc.PlayURL(ctx, params))
}

func dmSeg(ctx *bm.Context) {
	params := &model.DmSegParam{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(svc.DmSeg(ctx, params))
}
