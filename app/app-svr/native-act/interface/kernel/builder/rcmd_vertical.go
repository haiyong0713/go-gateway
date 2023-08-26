package builder

import (
	"context"
	"fmt"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type RcmdVertical struct{}

func (bu RcmdVertical) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rvCfg, ok := cfg.(*config.RcmdVertical)
	if !ok {
		logCfgAssertionError(config.RcmdVertical{})
		return nil
	}
	items := bu.buildModuleItems(rvCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeRcmdVertical.String(),
		ModuleId:    rvCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfRcmd(&rvCfg.RcmdCommon),
		ModuleItems: items,
		ModuleUkey:  rvCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu RcmdVertical) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu RcmdVertical) buildModuleItems(cfg *config.RcmdVertical, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	items := make([]*api.RcmdCard, 0, len(cfg.RcmdUsers))
	for _, user := range cfg.RcmdUsers {
		acc, ok := material.AccountCards[user.Mid]
		if !ok {
			continue
		}
		item := bu.buildRcmdVerticalItem(user, acc, material.Relations[user.Mid])
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
	return unshiftTitleCard([]*api.ModuleItem{moduleItem}, cfg.ImageTitle, "", ss.ReqFrom)
}

func (bu RcmdVertical) buildRcmdVerticalItem(user *config.RcmdUser, acc *accountgrpc.Card, rel *relationgrpc.FollowingReply) *api.RcmdCard {
	cd := &api.RcmdCard{
		Mid:          acc.GetMid(),
		Name:         acc.GetName(),
		Face:         acc.GetFace(),
		Uri:          fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", acc.GetMid()),
		Reason:       user.Reason,
		Official:     officialInfo2Native(&acc.Official),
		Vip:          vipInfo2Native(&acc.Vip),
		RedirectType: api.RedirectType_RtTypeSpace,
	}
	if rel != nil && (rel.Attribute == model.RelationFollow || rel.Attribute == model.RelationFriend) {
		cd.IsFollowed = true
	}
	if cd.Official.Role == model.RoleVertical {
		cd.Official.Role = model.RoleUp
	}
	if user.Uri != "" {
		cd.RedirectType = api.RedirectType_RtTypeUri
		cd.Uri = user.Uri
	}
	return cd
}
