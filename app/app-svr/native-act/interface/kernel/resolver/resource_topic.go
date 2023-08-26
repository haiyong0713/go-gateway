package resolver

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type ResourceTopic struct{}

func (r ResourceTopic) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.ResourceTopic{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ResourceCommon: buildResourceCommon(natModule, ss),
	}
	r.setBaseCfg(cfg, natModule, module.Dynamic, ss)
	return cfg
}

func (r ResourceTopic) setBaseCfg(cfg *config.ResourceTopic, module *natpagegrpc.NativeModule, dynamic *natpagegrpc.Dynamic, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	var types = int64(model.DynTypeVideo)
	if dynamic != nil && len(dynamic.SelectList) > 0 {
		types = dynamic.SelectList[0].SelectType
	}
	cfg.BriefDynsReqID, _ = cfg.AddMaterialParam(model.MaterialBriefDynsRly, &kernel.BriefDynsReq{
		Req: &model.BriefDynsReq{
			TopicID: module.Fid,
			From:    model.DynFromNative,
			Offset:  ss.OffsetStr,
			Types:   strconv.FormatInt(types, 10),
			Ps:      cfg.Ps,
			Mid:     ss.Mid(),
			SortBy:  0,
		},
		NeedMultiML: true,
		ArcType:     model.MaterialArchive,
	})
}
