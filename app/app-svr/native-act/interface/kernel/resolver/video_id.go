package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type VideoID struct{}

func (r VideoID) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.VideoID{
		BaseCfgManager: config.NewBaseCfg(natModule),
		VideoCommon:    buildVideoCommon(natModule, ss),
	}
	r.setBaseCfg(cfg, ss)
	return cfg
}

func (r VideoID) setBaseCfg(cfg *config.VideoID, ss *kernel.Session) {
	cfg.MixExtsReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtsRly, &kernel.ModuleMixExtsReq{
		Req: &natpagegrpc.ModuleMixExtsReq{
			ModuleID: cfg.ModuleBase().ModuleID,
			Ps:       cfg.Ps + 6,
			Offset:   ss.Offset,
		},
		NeedMultiML: true,
		ArcType:     model.MaterialArcPlayer,
	})
}

func buildVideoCommon(module *natpagegrpc.NativeModule, ss *kernel.Session) config.VideoCommon {
	colors := module.ColorsUnmarshal()
	var ps = module.Num
	if ss.ReqFrom == model.ReqFromSubPage {
		ps = 10
	}
	return config.VideoCommon{
		ImageTitle:         module.Meta,
		TextTitle:          module.Caption,
		AutoPlay:           module.IsAttrAutoPlay() == natpagegrpc.AttrModuleYes,
		HideTitle:          module.IsAttrHideTitle() == natpagegrpc.AttrModuleYes,
		DisplayViewMore:    module.IsAttrHideMore() != natpagegrpc.AttrModuleYes,
		BgColor:            module.BgColor,
		TitleColor:         colors.DisplayColor,
		CardTitleFontColor: module.TitleColor,
		SubpageTitle:       module.Title,
		ViewMoreFontColor:  module.FontColor,
		ViewMoreBgColor:    module.MoreColor,
		Ps:                 ps,
		PageID:             module.NativeID,
	}
}
