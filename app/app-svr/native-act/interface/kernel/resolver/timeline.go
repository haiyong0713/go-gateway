package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Timeline struct{}

func (r Timeline) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	colors := natModule.ColorsUnmarshal()
	cfg := &config.Timeline{
		BaseCfgManager:   config.NewBaseCfg(natModule),
		ImageTitle:       natModule.Meta,
		TextTitle:        natModule.Caption,
		NodeType:         confSort.Axis,
		TimePrecision:    confSort.TimeSort,
		BgColor:          natModule.BgColor,
		CardBgColor:      colors.TitleBgColor,
		TimelineColor:    colors.TimelineColor,
		ShowNum:          natModule.Num,
		ViewMoreType:     confSort.MoreSort,
		ViewMoreText:     natModule.Remark,
		SupernatantTitle: natModule.Title,
	}
	if model.IsFromIndex(ss.ReqFrom) {
		cfg.Ps = cfg.ShowNum
		if confSort.MoreSort == model.TimelineMoreByExpand {
			cfg.Ps = 50
		}
	} else {
		var ps int64 = 10
		const _psMax = 100
		if ss.Offset == 0 && ss.Index > 0 { //第一刷&有偏移量
			ps += ss.Index
			if ps > _psMax {
				ps = _psMax
			}
		}
		cfg.Ps = ps
	}
	r.setBaseCfg(cfg, ss)
	return cfg
}

func (r Timeline) setBaseCfg(cfg *config.Timeline, ss *kernel.Session) {
	cfg.MixExtsReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtsRly, &kernel.ModuleMixExtsReq{
		Req:         &natpagegrpc.ModuleMixExtsReq{ModuleID: cfg.ModuleBase().ModuleID, Ps: cfg.Ps + 6, Offset: ss.Offset},
		NeedMultiML: true,
		ArcType:     model.MaterialArchive,
	})
}
