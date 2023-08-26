package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Progress struct{}

func (r Progress) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Progress{
		BaseCfgManager:     config.NewBaseCfg(natModule),
		Style:              natModule.AvSort,
		BgColor:            natModule.BgColor,
		SlotType:           natModule.MoreColor,
		BarType:            natModule.TitleColor,
		BarColor:           natModule.FontColor,
		TextureType:        natModule.Length,
		DisplayProgressNum: natModule.IsAttrDisplayNum() == natpagegrpc.AttrModuleYes,
		Sid:                natModule.Fid,
		GroupID:            natModule.Width,
		DisplayNodeNum:     natModule.IsAttrDisplayNodeNum() == natpagegrpc.AttrModuleYes,
		DisplayNodeDesc:    natModule.IsAttrDisplayDesc() == natpagegrpc.AttrModuleYes,
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r Progress) setBaseCfg(cfg *config.Progress) {
	if cfg.Sid <= 0 || cfg.GroupID <= 0 {
		return
	}
	_, _ = cfg.AddMaterialParam(model.MaterialActProgressGroup, cfg.Sid, []int64{cfg.GroupID})
}
