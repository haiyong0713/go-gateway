package resolver

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Select struct{}

func (r Select) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	ryColors := natModule.ColorsUnmarshal()
	cfg := &config.Select{
		BaseCfgManager:         config.NewBaseCfg(natModule),
		Items:                  r.buildSelectItems(module.Select),
		Title:                  natModule.Title,
		BgColor:                natModule.BgColor,
		TopFontColor:           ryColors.SelectColor,
		PanelSelectColor:       ryColors.NotSelectColor,
		PanelBgColor:           ryColors.PanelBgColor,
		PanelSelectFontColor:   ryColors.PanelSelectColor,
		PanelNtSelectFontColor: ryColors.PanelNotSelectColor,
		PrimaryPageID:          natPage.ID,
		ShareInfo: &config.ShareInfo{
			Image:   natPage.ShareImage,
			Title:   natPage.ShareTitle,
			Caption: natPage.ShareCaption,
		},
	}
	r.setBaseCfg(cfg, ss)
	return cfg
}

func (r Select) buildSelectItems(ses *natpagegrpc.Select) []*config.SelectItem {
	if ses == nil || len(ses.List) == 0 {
		return nil
	}
	items := make([]*config.SelectItem, 0, len(ses.List))
	for _, v := range ses.List {
		if v == nil || v.MType != natpagegrpc.MixInlineType || v.ForeignID == 0 || !v.IsOnline() {
			continue
		}
		item := &config.SelectItem{PageID: v.ForeignID}
		if ext, err := config.UnmarshalSelectedExt(v.Reason); err == nil {
			item.SelectExt = *ext
		}
		items = append(items, item)
	}
	return items
}

func (r Select) setBaseCfg(cfg *config.Select, ss *kernel.Session) {
	pageIDs := make([]int64, 0, len(cfg.Items))
	weekIDs := make([]int64, 0, len(cfg.Items))
	for _, item := range cfg.Items {
		if item == nil {
			continue
		}
		pageIDs = append(pageIDs, item.PageID)
		if model.IsFromIndex(ss.ReqFrom) && item.Type == natpagegrpc.SelectWeek && item.LocationKey != "" {
			weekid, _ := strconv.ParseInt(item.LocationKey, 10, 64)
			if weekid > 0 {
				weekIDs = append(weekIDs, weekid)
			}
		}
	}
	_, _ = cfg.AddMaterialParam(model.MaterialNativePages, pageIDs)
	if len(weekIDs) > 0 {
		_, _ = cfg.AddMaterialParam(model.MaterialWeeks, weekIDs)
	}
}
