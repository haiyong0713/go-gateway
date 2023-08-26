package builder

import (
	"context"
	"fmt"
	"strconv"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Rcmd struct{}

func (bu Rcmd) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rcmdCfg, ok := cfg.(*config.Rcmd)
	if !ok {
		logCfgAssertionError(config.Rcmd{})
		return nil
	}
	items := bu.buildModuleItems(rcmdCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeRcmd.String(),
		ModuleId:    rcmdCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfRcmd(&rcmdCfg.RcmdCommon),
		ModuleItems: items,
		ModuleUkey:  rcmdCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Rcmd) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Rcmd) buildModuleItems(cfg *config.Rcmd, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	items := make([]*api.ModuleItem, 0, len(cfg.RcmdUsers))
	for _, user := range cfg.RcmdUsers {
		acc, ok := material.AccountCards[user.Mid]
		if !ok {
			continue
		}
		cd := bu.buildRcmdCard(user, acc, material.Relations[user.Mid])
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeRcmd.String(),
			CardId:     strconv.FormatInt(user.Mid, 10),
			CardDetail: &api.ModuleItem_RecommendCard{RecommendCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	return unshiftTitleCard(items, cfg.ImageTitle, "", ss.ReqFrom)
}

func (bu Rcmd) buildRcmdCard(user *config.RcmdUser, acc *accountgrpc.Card, rel *relationgrpc.FollowingReply) *api.RcmdCard {
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
	return cd
}

func officialInfo2Native(official *accountgrpc.OfficialInfo) *api.OfficialInfo {
	return &api.OfficialInfo{
		Role:  official.GetRole(),
		Title: official.GetTitle(),
		Desc:  official.GetDesc(),
		Type:  official.GetType(),
	}
}

func vipInfo2Native(vip *accountgrpc.VipInfo) *api.VipInfo {
	return &api.VipInfo{
		Type:       vip.GetType(),
		Status:     vip.GetStatus(),
		DueDate:    vip.GetDueDate(),
		VipPayType: vip.GetVipPayType(),
		ThemeType:  vip.GetThemeType(),
		Label: &api.VipLabel{
			Path:        vip.GetLabel().Path,
			Text:        vip.GetLabel().Text,
			LabelTheme:  vip.GetLabel().LabelTheme,
			TextColor:   vip.GetLabel().TextColor,
			BgStyle:     vip.GetLabel().BgStyle,
			BgColor:     vip.GetLabel().BgColor,
			BorderColor: vip.GetLabel().BorderColor,
		},
		AvatarSubscript:    vip.GetAvatarSubscript(),
		NicknameColor:      vip.GetNicknameColor(),
		Role:               vip.GetRole(),
		AvatarSubscriptUrl: vip.GetAvatarSubscriptUrl(),
	}
}

func buildModuleColorOfRcmd(cfg *config.RcmdCommon) *api.Color {
	return &api.Color{
		BgColor:            cfg.BgColor,
		CardTitleFontColor: cfg.CardTitleFontColor,
	}
}
