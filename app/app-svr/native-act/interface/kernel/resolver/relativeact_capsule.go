package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type RelativeactCapsule struct{}

func (r RelativeactCapsule) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.RelativeactCapsule{
		BaseCfgManager: config.NewBaseCfg(natModule),
		TextTitle:      natModule.Caption,
		BgColor:        natModule.BgColor,
		PageIDs:        r.pageIDs(module.ActPage, natPage, ss),
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r RelativeactCapsule) pageIDs(actPage *natpagegrpc.ActPage, page *natpagegrpc.NativePage, ss *kernel.Session) []int64 {
	if actPage == nil || len(actPage.List) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(actPage.List))
	for _, item := range actPage.List {
		if item.PageID == page.ID {
			continue
		}
		ids = append(ids, item.PageID)
	}
	if ss.TabFrom == model.TabFromUserSpace && page.IsUpTopicAct() {
		ids = append([]int64{page.ID}, ids...)
	}
	return ids
}

func (r RelativeactCapsule) setBaseCfg(cfg *config.RelativeactCapsule) {
	_, _ = cfg.AddMaterialParam(model.MaterialNativeCard, cfg.PageIDs)
	_, _ = cfg.AddMaterialParam(model.MaterialNativeAllPage, cfg.PageIDs)
}
