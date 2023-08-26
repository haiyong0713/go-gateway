package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Icon struct{}

func (bu Icon) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	icCfg, ok := cfg.(*config.Icon)
	if !ok {
		logCfgAssertionError(config.Icon{})
		return nil
	}
	items := bu.buildModuleItems(icCfg)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeIcon.String(),
		ModuleId:    icCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: icCfg.BgColor, FontColor: icCfg.FontColor},
		ModuleItems: items,
		ModuleUkey:  icCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Icon) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Icon) buildModuleItems(cfg *config.Icon) []*api.ModuleItem {
	iconItems := make([]*api.IconItem, 0, len(cfg.Items))
	for _, icon := range cfg.Items {
		iconItems = append(iconItems, &api.IconItem{
			Title: icon.Title,
			Image: icon.Image,
			Uri:   icon.Uri,
		})
	}
	if len(iconItems) == 0 {
		return nil
	}
	item := &api.ModuleItem{
		CardType: model.CardTypeIcon.String(),
		CardDetail: &api.ModuleItem_IconCard{
			IconCard: &api.IconCard{Items: iconItems},
		},
	}
	return []*api.ModuleItem{item}
}
