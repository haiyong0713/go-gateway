package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Reserve struct{}

func (r Reserve) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Reserve{
		BaseCfgManager:    config.NewBaseCfg(natModule),
		ImageTitle:        natModule.Meta,
		TextTitle:         natModule.Caption,
		BgColor:           natModule.BgColor,
		FontColor:         natModule.TitleColor,
		CardBgColor:       natModule.ColorsUnmarshal().TitleBgColor,
		DisplayUpFaceName: natModule.IsAttrIsDisplayUpIcon() == natpagegrpc.AttrModuleYes,
		UpRsvIDs:          r.rsvIDs(module.Reserve),
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r Reserve) rsvIDs(reserve *natpagegrpc.Reserve) []int64 {
	if reserve == nil {
		return nil
	}
	var rsvIDs []int64
	for _, ext := range reserve.List {
		if ext == nil || ext.MType != natpagegrpc.MixUpReserve || ext.ForeignID <= 0 {
			continue
		}
		rsvIDs = append(rsvIDs, ext.ForeignID)
	}
	return rsvIDs
}

func (r Reserve) setBaseCfg(cfg *config.Reserve) {
	cfg.UpRsvIDsReqID, _ = cfg.AddMaterialParam(model.MaterialUpRsvInfo, &kernel.UpRsvIDsReq{
		IDs:         cfg.UpRsvIDs,
		NeedMultiML: true,
		NeedAccount: cfg.DisplayUpFaceName,
	})
}
