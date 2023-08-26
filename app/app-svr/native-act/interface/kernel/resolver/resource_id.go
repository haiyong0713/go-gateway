package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type ResourceID struct{}

func (r ResourceID) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.ResourceID{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ResourceCommon: buildResourceCommon(natModule, ss),
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r ResourceID) setBaseCfg(cfg *config.ResourceID, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	var isLive int64
	if module.IsAttrDisplayNodeNum() == natpagegrpc.AttrModuleYes {
		isLive = 1
	}
	// 多取几条以满足展示数量
	cfg.MixExtsReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtsRly, &kernel.ModuleMixExtsReq{
		Req:         &natpagegrpc.ModuleMixExtsReq{ModuleID: cfg.ModuleBase().ModuleID, Ps: cfg.Ps + 6, Offset: ss.Offset},
		NeedMultiML: true,
		IsLive:      isLive,
		ArcType:     model.MaterialArchive,
	})
}

func buildResourceCommon(module *natpagegrpc.NativeModule, ss *kernel.Session) config.ResourceCommon {
	colors := module.ColorsUnmarshal()
	var ps = module.Num
	if ss.ReqFrom == model.ReqFromSubPage {
		ps = 10
	}
	return config.ResourceCommon{
		ImageTitle:          module.Meta,
		TextTitle:           module.Caption,
		SubpageTitle:        module.Title,
		DisplayUGCBadge:     module.IsAttrDisplayVideoIcon() == natpagegrpc.AttrModuleYes,
		DisplayPGCBadge:     module.IsAttrDisplayPgcIcon() == natpagegrpc.AttrModuleYes,
		DisplayArticleBadge: module.IsAttrDisplayArticleIcon() == natpagegrpc.AttrModuleYes,
		DisplayViewMore:     module.IsAttrHideMore() != natpagegrpc.AttrModuleYes,
		DisplayOnlyLive:     module.IsAttrDisplayNodeNum() == natpagegrpc.AttrModuleYes,
		BgColor:             module.BgColor,
		TitleColor:          colors.DisplayColor,
		CardTitleFontColor:  module.TitleColor,
		CardTitleBgColor:    colors.TitleBgColor,
		ViewMoreFontColor:   module.FontColor,
		ViewMoreBgColor:     module.MoreColor,
		Ps:                  ps,
		PageID:              module.NativeID,
	}
}
