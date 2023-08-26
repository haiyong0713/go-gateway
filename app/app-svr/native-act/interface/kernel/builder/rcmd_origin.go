package builder

import (
	"context"
	"fmt"
	"strconv"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type RcmdOrigin struct{}

func (bu RcmdOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	roCfg, ok := cfg.(*config.RcmdOrigin)
	if !ok {
		logCfgAssertionError(config.RcmdOrigin{})
		return nil
	}
	items := bu.buildModuleItems(roCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeRcmd.String(),
		ModuleId:    roCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfRcmd(&roCfg.RcmdCommon),
		ModuleItems: items,
		ModuleUkey:  roCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu RcmdOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu RcmdOrigin) buildModuleItems(cfg *config.RcmdOrigin, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	var items []*api.ModuleItem
	switch cfg.SourceType {
	case model.SourceTypeActUp:
		items = bu.buildModuleItemsOfActUp(cfg, material)
	case model.SourceTypeRank:
		items = bu.buildModuleItemsOfRank(cfg, material)
	}
	if len(items) == 0 {
		return nil
	}
	return unshiftTitleCard(items, cfg.ImageTitle, "", ss.ReqFrom)
}

func (bu RcmdOrigin) buildModuleItemsOfActUp(cfg *config.RcmdOrigin, material *kernel.Material) []*api.ModuleItem {
	upList, ok := material.UpListRlys[cfg.UpListReqID]
	if !ok || len(upList.List) == 0 {
		return nil
	}
	rcmdBuilder := Rcmd{}
	items := make([]*api.ModuleItem, 0, len(upList.List))
	for _, upItem := range upList.List {
		if upItem == nil || upItem.Account == nil {
			continue
		}
		acc, ok := material.AccountCards[upItem.Account.Mid]
		if !ok {
			continue
		}
		cd := rcmdBuilder.buildRcmdCard(&config.RcmdUser{Mid: acc.Mid}, acc, material.Relations[acc.Mid])
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeRcmd.String(),
			CardId:     strconv.FormatInt(acc.Mid, 10),
			CardDetail: &api.ModuleItem_RecommendCard{RecommendCard: cd},
		})
	}
	return items
}

func (bu RcmdOrigin) buildModuleItemsOfRank(cfg *config.RcmdOrigin, material *kernel.Material) []*api.ModuleItem {
	rankRst, ok := material.RankRstRlys[cfg.RankRstReqID]
	if !ok || len(rankRst.List) == 0 {
		return nil
	}
	var i int64
	icons := bu.rankIcons(material.MixExtRlys[cfg.MixExtReqID])
	items := make([]*api.ModuleItem, 0, len(rankRst.List))
	for _, rank := range rankRst.List {
		if rank == nil || rank.Account == nil || rank.ObjectType != 1 {
			continue
		}
		cd := bu.buildCardOfRank(rank, material.Relations[rank.Account.MID], cfg.DisplayRankScore, icons[i])
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeRcmd.String(),
			CardId:     strconv.FormatInt(rank.Account.MID, 10),
			CardDetail: &api.ModuleItem_RecommendCard{RecommendCard: cd},
		})
		i++
	}
	return items
}

func (bu RcmdOrigin) rankIcons(mixExtRly *natpagegrpc.ModuleMixExtReply) map[int64]string {
	if mixExtRly == nil {
		return map[int64]string{}
	}
	var i int64
	icons := make(map[int64]string, len(mixExtRly.List))
	for _, v := range mixExtRly.List {
		if v == nil {
			continue
		}
		if remark := v.RemarkUnmarshal(); remark.Image != "" {
			icons[i] = remark.Image
			i++
		}
	}
	return icons
}

func (bu RcmdOrigin) buildCardOfRank(rank *activitygrpc.RankResult, rel *relationgrpc.FollowingReply, displayScore bool, rankIcon string) *api.RcmdCard {
	cd := &api.RcmdCard{
		RankIcon:     rankIcon,
		RedirectType: api.RedirectType_RtTypeSpace,
	}
	if acc := rank.Account; acc != nil {
		cd.Mid = acc.MID
		cd.Name = acc.Name
		cd.Face = acc.Face
		cd.Uri = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", acc.MID)
		cd.Official = bu.transOfficialInfo(&acc.Official)
		cd.Vip = bu.transVipInfo(&acc.Vip)
		if cd.Official.Role == model.RoleVertical {
			cd.Official.Role = model.RoleUp
		}
	}
	if rel != nil {
		if rel.Attribute == model.RelationFollow || rel.Attribute == model.RelationFriend {
			cd.IsFollowed = true
		}
	}
	if displayScore {
		cd.Reason = rank.ShowScore
	}
	return cd
}

func (bu RcmdOrigin) transOfficialInfo(official *activitygrpc.OfficialInfo) *api.OfficialInfo {
	return &api.OfficialInfo{
		Role:  official.Role,
		Title: official.Title,
		Desc:  official.Desc,
		Type:  official.Type,
	}
}

func (bu RcmdOrigin) transVipInfo(vip *activitygrpc.VipInfo) *api.VipInfo {
	return &api.VipInfo{
		Type:       vip.Type,
		Status:     vip.Status,
		DueDate:    vip.DueDate,
		VipPayType: vip.VipPayType,
		ThemeType:  vip.ThemeType,
		Label: &api.VipLabel{
			Path:        vip.Label.Path,
			Text:        vip.Label.Text,
			LabelTheme:  vip.Label.LabelTheme,
			TextColor:   vip.Label.TextColor,
			BgStyle:     vip.Label.BgStyle,
			BgColor:     vip.Label.BgColor,
			BorderColor: vip.Label.BorderColor,
		},
		AvatarSubscript:    vip.AvatarSubscript,
		NicknameColor:      vip.NicknameColor,
		Role:               vip.Role,
		AvatarSubscriptUrl: vip.AvatarSubscriptUrl,
	}
}
