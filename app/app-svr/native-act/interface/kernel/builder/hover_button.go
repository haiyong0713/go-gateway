package builder

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type HoverButton struct{}

func (bu HoverButton) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	hbCfg, ok := cfg.(*config.HoverButton)
	if !ok {
		logCfgAssertionError(config.HoverButton{})
		return nil
	}
	items := bu.buildModuleItems(hbCfg, material)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeHoverButton.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleItems: items,
		ModuleUkey:  cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu HoverButton) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu HoverButton) buildModuleItems(cfg *config.HoverButton, material *kernel.Material) []*api.ModuleItem {
	item := &api.ClickItem{}
	ckBu := Click{}
	switch cfg.Item.Type {
	case model.ClickTypeReserve:
		ckBu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeReserve, &cfg.Item, func() bool {
			if rly, ok := material.ActRsvFollows[cfg.Item.Id]; ok && rly.IsFollow {
				return true
			}
			return false
		})
	case model.ClickTypeActivity:
		ckBu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeActivity, &cfg.Item, func() bool {
			if info, ok := material.ActRelationInfos[cfg.Item.Id]; ok && info.ReserveItems != nil && info.ReserveItems.State == 1 {
				return true
			}
			return false
		})
	case model.ClickTypeBtnRedirect:
		item.Action = api.ClickItem_ActRedirect
		item.ActionDetail = &api.ClickItem_RedirectAct{
			RedirectAct: &api.ClickActRedirect{Url: cfg.Item.Url, Image: cfg.Item.Image},
		}
	default:
		log.Warn("unknown click_type=%+v of hover_button", cfg.Item.Type)
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType: model.CardTypeHoverButton.String(),
		CardDetail: &api.ModuleItem_HoverButtonCard{
			HoverButtonCard: &api.HoverButtonCard{Item: item, MutexUkeys: cfg.MutexUkeys},
		},
	}
	return []*api.ModuleItem{moduleItem}
}
