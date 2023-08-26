package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Tab struct{}

func (r Tab) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Tab{
		BaseCfgManager:      config.NewBaseCfg(natModule),
		Style:               natModule.AvSort,
		BgColor:             natModule.BgColor,
		SelectedFontColor:   natModule.MoreColor,
		UnselectedFontColor: natModule.FontColor,
		DisplayUnfoldButton: natModule.IsAttrDisplayButton() == natpagegrpc.AttrModuleYes,
		BgImage:             config.SizeImage{Image: natModule.Meta, Height: natModule.Length, Width: natModule.Width},
		Items:               r.buildTabItems(module.InlineTab),
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r Tab) buildTabItems(tabs *natpagegrpc.InlineTab) []*config.TabItem {
	if tabs == nil || len(tabs.List) == 0 {
		return nil
	}
	items := make([]*config.TabItem, 0, len(tabs.List))
	for _, tab := range tabs.List {
		if tab == nil || tab.MType != natpagegrpc.MixInlineType || tab.ForeignID == 0 || !tab.IsOnline() {
			continue
		}
		item := &config.TabItem{PageID: tab.ForeignID}
		if ext, err := model.UnmarshalTabExt(tab.Reason); err == nil {
			item.TabItemExt = *ext
		}
		items = append(items, item)
	}
	return items
}

func (r Tab) setBaseCfg(cfg *config.Tab) {
	pageIDs := make([]int64, 0, len(cfg.Items))
	for _, item := range cfg.Items {
		pageIDs = append(pageIDs, item.PageID)
	}
	_, _ = cfg.AddMaterialParam(model.MaterialNativePages, pageIDs)
}
