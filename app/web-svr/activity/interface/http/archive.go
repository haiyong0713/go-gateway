package http

import (
	bm "go-common/library/net/http/blademaster"
	xbinding "go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/activity/interface/service"
)

func getArchiveByBvid(ctx *bm.Context) {
	v := new(struct {
		Bvids []string `json:"bvids" validate:"required"`
	})
	if err := ctx.BindWith(v, xbinding.JSON); err != nil {
		return
	}
	bvidMap := service.KnowledgeSvr.GetBvidMap()
	bvidList := make([]string, 0)
	for _, v := range v.Bvids {
		if _, ok := bvidMap[v]; ok {
			bvidList = append(bvidList, v)
		}
	}
	ctx.JSON(service.ArcSvc.GetArchiveByBvid(ctx, bvidList))
}
