package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
)

func regionIndex(ctx *bm.Context) {
	v := new(struct {
		Platform    string `form:"platform" validate:"required"`
		Lang        string `form:"lang"`
		TinyMode    int    `form:"tiny_mode"`    // 极小包模式【暂时使用单个字段，更清晰，后续如需要，改为位控制】
		TeenageMode int    `form:"teenage_mode"` // 极小包青少年模式【暂时使用单个字段，更清晰，后续如需要，改为位控制】
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var plat int
	switch v.Platform {
	case "html5":
		plat = model.PlatH5
	case "xcx":
		plat = model.PlatXcx
	default:
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	ctx.JSON(webSvc.RegionIndex(ctx, plat, v.Lang, ip, v.TinyMode, v.TeenageMode))
}
