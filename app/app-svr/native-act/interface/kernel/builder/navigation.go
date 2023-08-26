package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Navigation struct{}

func (bu Navigation) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	naviCfg, ok := cfg.(*config.Navigation)
	if !ok {
		logCfgAssertionError(config.Navigation{})
		return nil
	}
	module := &api.Module{
		ModuleType: model.ModuleTypeNavigation.String(),
		ModuleId:   naviCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{
			SelectedFontColor:     naviCfg.SelectedFontColor,
			SelectedBgColor:       naviCfg.SelectedBgColor,
			UnselectedFontColor:   naviCfg.UnselectedFontColor,
			UnselectedBgColor:     naviCfg.UnselectedBgColor,
			NtSelectedFontColor:   naviCfg.NtSelectedFontColor,
			NtSelectedBgColor:     naviCfg.NtSelectedBgColor,
			NtUnselectedFontColor: naviCfg.NtUnselectedFontColor,
			NtUnselectedBgColor:   naviCfg.NtUnselectedBgColor,
		},
		ModuleUkey: naviCfg.ModuleBase().Ukey,
	}
	ss.PageRlyContext.HasNavigation = true
	return module
}

func (bu Navigation) After(data *AfterContextData, current *api.Module) bool {
	if len(data.NaviItems) == 0 {
		return false
	}
	current.ModuleItems = []*api.ModuleItem{
		{
			CardType: model.CardTypeNavigation.String(),
			CardDetail: &api.ModuleItem_NavigationCard{
				NavigationCard: &api.NavigationCard{Items: data.NaviItems},
			},
		},
	}
	return true
}

func (bu Navigation) BuildNavigationItems(cfgs []config.BaseCfgManager) []*api.NavigationItem {
	items := make([]*api.NavigationItem, 0, len(cfgs))
	for _, cfg := range cfgs {
		if cfg.ModuleBase().Bar == "" {
			continue
		}
		items = append(items, &api.NavigationItem{
			ModuleId: cfg.ModuleBase().ModuleID,
			Title:    cfg.ModuleBase().Bar,
		})
	}
	return items
}
