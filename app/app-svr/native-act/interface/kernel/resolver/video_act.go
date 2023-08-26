package resolver

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type VideoAct struct{}

func (r VideoAct) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.VideoAct{
		BaseCfgManager: config.NewBaseCfg(natModule),
		VideoCommon:    buildVideoCommon(natModule, ss),
		SortType:       actSortType(module, ss),
		SortList:       actSortList(module),
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r VideoAct) setBaseCfg(cfg *config.VideoAct, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	cfg.ActLikesReqID, _ = cfg.AddMaterialParam(model.MaterialActLikesRly, &kernel.ActLikesReq{
		Req: &activitygrpc.ActLikesReq{
			Sid:      module.Fid,
			Mid:      ss.Mid(),
			SortType: int32(cfg.SortType),
			Ps:       int32(cfg.Ps),
			Offset:   ss.Offset,
		},
		NeedMultiML: true,
		ArcType:     model.MaterialArcPlayer,
	})
}
