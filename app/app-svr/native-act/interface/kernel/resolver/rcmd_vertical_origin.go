package resolver

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type RcmdVerticalOrigin struct{}

func (r RcmdVerticalOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	cfg := &config.RcmdVerticalOrigin{
		BaseCfgManager: config.NewBaseCfg(natModule),
		RcmdCommon:     buildRcmdCommon(natModule),
		SourceType:     confSort.SourceType,
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r RcmdVerticalOrigin) setBaseCfg(cfg *config.RcmdVerticalOrigin, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	confSort := module.ConfUnmarshal()
	switch confSort.SourceType {
	case model.SourceTypeActUp:
		sortType := module.ConfUnmarshal().SortType
		if sortType == "" {
			sortType = model.SortTypeCtime
		}
		cfg.UpListReqID, _ = cfg.AddMaterialParam(model.MaterialUpListRly, &kernel.UpListReq{
			Req:         &activitygrpc.UpListReq{Sid: module.Fid, Type: sortType, Pn: 1, Ps: 40, Mid: ss.Mid()},
			NeedMultiML: true,
		})
	default:
		log.Warn("unknown sourceType of RcmdVerticalOrigin, sourceType=%s module=%+v", confSort.SourceType, module)
		return
	}
}
