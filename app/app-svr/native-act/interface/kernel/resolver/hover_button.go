package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type HoverButton struct{}

func (r HoverButton) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	ext, err := model.UnmarshalHoverButtonExt(natModule.ConfSort)
	if err != nil {
		return nil
	}
	cfg := &config.HoverButton{
		BaseCfgManager: config.NewBaseCfg(natModule),
	}
	switch ext.BtType {
	case model.HoverBtnReserve:
		r.setReserve(cfg, ext, natModule)
	case model.HoverBtnActivity:
		r.setActivity(cfg, ext, natModule)
	case model.HoverBtnRedirect:
		r.setRedirect(cfg, natModule)
	}
	return cfg
}

func (r HoverButton) setReserve(cfg *config.HoverButton, ext *model.HoverButtonExt, module *natpagegrpc.NativeModule) {
	cfg.Item.Type = model.ClickTypeReserve
	cfg.Item.Id = module.Fid
	cfg.Item.DoneImage = module.TitleColor
	cfg.Item.UndoneImage = module.FontColor
	cfg.Item.MsgBoxTip = ext.Hint
	cfg.MutexUkeys = ext.MUkeys
	if cfg.Item.Id > 0 {
		_, _ = cfg.AddMaterialParam(model.MaterialActReserveFollow, []int64{cfg.Item.Id})
	}
}

func (r HoverButton) setActivity(cfg *config.HoverButton, ext *model.HoverButtonExt, module *natpagegrpc.NativeModule) {
	cfg.Item.Type = model.ClickTypeActivity
	cfg.Item.Id = module.Fid
	cfg.Item.DoneImage = module.TitleColor
	cfg.Item.UndoneImage = module.FontColor
	cfg.Item.MsgBoxTip = ext.Hint
	cfg.MutexUkeys = ext.MUkeys
	if cfg.Item.Id > 0 {
		_, _ = cfg.AddMaterialParam(model.MaterialActRelationInfo, []int64{cfg.Item.Id})
	}
}

func (r HoverButton) setRedirect(cfg *config.HoverButton, module *natpagegrpc.NativeModule) {
	cfg.Item.Type = model.ClickTypeBtnRedirect
	cfg.Item.Image = module.MoreColor
	cfg.Item.Url = module.Colors
}
