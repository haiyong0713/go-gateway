package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Ogv struct{}

func (r Ogv) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Ogv{
		BaseCfgManager: config.NewBaseCfg(natModule),
		OgvCommon:      buildOgvCommon(natModule, ss),
	}
	r.setBaseCfg(cfg, ss)
	return cfg
}

func (r Ogv) setBaseCfg(cfg *config.Ogv, ss *kernel.Session) {
	cfg.MixExtsReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtsRly, &kernel.ModuleMixExtsReq{
		Req:         &natpagegrpc.ModuleMixExtsReq{ModuleID: cfg.ModuleBase().ModuleID, Ps: cfg.Ps + 6, Offset: ss.Offset},
		NeedMultiML: true,
	})
}

func buildOgvCommon(module *natpagegrpc.NativeModule, ss *kernel.Session) config.OgvCommon {
	colors := module.ColorsUnmarshal()
	var ps = module.Num
	if ss.ReqFrom == model.ReqFromSubPage {
		ps = 10
		const _psMax = 100
		if ss.Offset == 0 && ss.Index > 0 { //第一刷&有偏移量
			ps += ss.Index
			if ps > _psMax {
				ps = _psMax
			}
		}
	}
	return config.OgvCommon{
		ImageTitle: module.Meta,
		TextTitle:  module.Caption,
		Color: &config.OgvColor{
			BgColor:           module.BgColor,
			CardBgColor:       colors.TitleBgColor,
			ViewMoreFontColor: module.FontColor,
			ViewMoreBgColor:   module.MoreColor,
			RcmdFontColor:     colors.SubtitleColor,
			TitleColor:        module.TitleColor,
			SubtitleFontColor: colors.SubtitleColor,
		},
		IsThreeCard:      module.IsCardThree(),
		Ps:               ps,
		ViewMoreText:     module.Remark,
		SupernatantTitle: module.Title,
		DisplayPayBadge:  module.IsAttrDisplayPgcIcon() == natpagegrpc.AttrModuleYes,
		DisplayScore:     module.IsAttrDisplayNum() == natpagegrpc.AttrModuleYes,
		DisplayRcmd:      module.IsAttrDisplayRecommend() == natpagegrpc.AttrModuleYes,
		DisplaySubtitle:  module.IsAttrDisplayDesc() == natpagegrpc.AttrModuleYes,
		DisplayMore:      module.IsAttrHideMore() != natpagegrpc.AttrModuleYes,
	}
}
