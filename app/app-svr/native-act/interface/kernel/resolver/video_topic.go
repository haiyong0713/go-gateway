package resolver

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type VideoTopic struct{}

func (r VideoTopic) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.VideoTopic{
		BaseCfgManager: config.NewBaseCfg(natModule),
		VideoCommon:    buildVideoCommon(natModule, ss),
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r VideoTopic) setBaseCfg(cfg *config.VideoTopic, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	cfg.BriefDynsReqID, _ = cfg.AddMaterialParam(model.MaterialBriefDynsRly, &kernel.BriefDynsReq{
		Req: &model.BriefDynsReq{
			TopicID: module.Fid,
			From:    model.DynFromNative,
			Offset:  ss.OffsetStr,
			Types:   strconv.FormatInt(model.DynTypeVideo, 10),
			Ps:      cfg.Ps,
			Mid:     ss.Mid(),
			SortBy:  0,
		},
		NeedMultiML: true,
		ArcType:     model.MaterialArcPlayer,
	})
}
