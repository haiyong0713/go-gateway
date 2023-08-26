package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type RcmdVerticalOrigin struct{}

func (bu RcmdVerticalOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rvoCfg, ok := cfg.(*config.RcmdVerticalOrigin)
	if !ok {
		logCfgAssertionError(config.RcmdVerticalOrigin{})
		return nil
	}
	items := bu.buildModuleItems(rvoCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeRcmdVertical.String(),
		ModuleId:    rvoCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfRcmd(&rvoCfg.RcmdCommon),
		ModuleItems: items,
		ModuleUkey:  rvoCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu RcmdVerticalOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu RcmdVerticalOrigin) buildModuleItems(cfg *config.RcmdVerticalOrigin, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	var items []*api.ModuleItem
	switch cfg.SourceType {
	case model.SourceTypeActUp:
		items = bu.buildModuleItemsOfActUp(cfg, material)
	}
	if len(items) == 0 {
		return nil
	}
	return unshiftTitleCard(items, cfg.ImageTitle, "", ss.ReqFrom)
}

func (bu RcmdVerticalOrigin) buildModuleItemsOfActUp(cfg *config.RcmdVerticalOrigin, material *kernel.Material) []*api.ModuleItem {
	upList, ok := material.UpListRlys[cfg.UpListReqID]
	if !ok || len(upList.List) == 0 {
		return nil
	}
	rcmdBuilder := Rcmd{}
	items := make([]*api.RcmdCard, 0, len(upList.List))
	for _, upItem := range upList.List {
		if upItem == nil || upItem.Account == nil {
			continue
		}
		acc, ok := material.AccountCards[upItem.Account.Mid]
		if !ok {
			continue
		}
		item := rcmdBuilder.buildRcmdCard(&config.RcmdUser{Mid: acc.Mid}, acc, material.Relations[acc.Mid])
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType: model.CardTypeRcmdVertical.String(),
		CardDetail: &api.ModuleItem_RecommendVerticalCard{
			RecommendVerticalCard: &api.RcmdVerticalCard{Items: items},
		},
	}
	return []*api.ModuleItem{moduleItem}
}
