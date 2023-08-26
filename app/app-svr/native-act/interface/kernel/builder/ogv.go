package builder

import (
	"context"
	"fmt"
	"strconv"

	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Ogv struct{}

func (bu Ogv) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	ogvCfg, ok := cfg.(*config.Ogv)
	if !ok {
		logCfgAssertionError(config.Ogv{})
		return nil
	}
	hasMore, offset, items := bu.buildModuleItems(ogvCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeOgv.String(),
		ModuleId:    ogvCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfOgv(ogvCfg.Color, ogvCfg.IsThreeCard),
		ModuleItems: items,
		ModuleUkey:  ogvCfg.ModuleBase().Ukey,
		HasMore:     hasMore,
	}
	if hasMore && ogvCfg.DisplayMore {
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildOgvMore(&ogvCfg.OgvCommon, offset, ogvCfg.ModuleBase().ModuleID))
		} else {
			module.SubpageParams = ogvSupernatantParams(offset, -1, ogvCfg.ModuleBase().ModuleID)
		}
	}
	return module
}

func (bu Ogv) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func buildModuleColorOfOgv(cfg *config.OgvColor, isThreeCard bool) *api.Color {
	color := &api.Color{
		BgColor:           cfg.BgColor,
		ViewMoreFontColor: cfg.ViewMoreFontColor,
		ViewMoreBgColor:   cfg.ViewMoreBgColor,
	}
	if isThreeCard {
		color.TitleColor = cfg.TitleColor
		color.SubtitleFontColor = cfg.SubtitleFontColor
	} else {
		color.CardBgColor = cfg.CardBgColor
		color.RcmdFontColor = cfg.RcmdFontColor
	}
	return color
}

func (bu Ogv) buildModuleItems(cfg *config.Ogv, material *kernel.Material, ss *kernel.Session) (bool, int64, []*api.ModuleItem) {
	mixRly, ok := material.MixExtsRlys[cfg.MixExtsReqID]
	if !ok || len(mixRly.List) == 0 {
		return false, 0, nil
	}
	items := make([]*api.ModuleItem, 0, len(mixRly.List))
	offset := ss.Offset
	for _, ext := range mixRly.List {
		if int64(len(items)) >= cfg.Ps {
			break
		}
		offset++
		if ext == nil || ext.ForeignID == 0 || ext.MType != natpagegrpc.MixOgvSsid {
			continue
		}
		sc, ok := material.SeasonCards[int32(ext.ForeignID)]
		if !ok {
			continue
		}
		if cfg.IsThreeCard {
			items = append(items, bu.buildOgvThree(&cfg.OgvCommon, sc, ext.RemarkUnmarshal().Title))
		} else {
			items = append(items, bu.buildOgvOne(&cfg.OgvCommon, sc, ext.RemarkUnmarshal().Title))
		}
	}
	if len(items) == 0 {
		return false, 0, nil
	}
	hasMore := mixRly.HasMore == 1
	if !hasMore && offset < mixRly.Offset {
		hasMore = true
	}
	items = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	return hasMore, offset, items
}

func (bu Ogv) buildOgvThree(cfg *config.OgvCommon, sc *pgcappgrpc.SeasonCardInfoProto, subtitle string) *api.ModuleItem {
	cd := &api.OgvThreeCard{
		CoverLeftText1: "-",
		Image:          sc.Cover,
		Title:          sc.Title,
		ReportDic: &api.ReportDic{
			BizType:    model.ReportBizTypePGC,
			SeasonType: strconv.FormatInt(int64(sc.SeasonType), 10),
			SeasonId:   int64(sc.SeasonId),
		},
		Url:          sc.Url,
		ResourceType: model.ResourceTypeOGV,
	}
	if sc.Stat != nil {
		cd.CoverLeftText1 = sc.Stat.FollowView
	}
	if cfg.DisplaySubtitle {
		cd.Subtitle = subtitle
		if cd.Subtitle == "" {
			cd.Subtitle = sc.RecommendView
		}
	}
	if cfg.DisplayPayBadge && sc.BadgeInfo != nil {
		cd.Badge = &api.Badge{
			Text:         sc.BadgeInfo.Text,
			BgColor:      sc.BadgeInfo.BgColor,
			BgColorNight: sc.BadgeInfo.BgColorNight,
		}
	}
	if sc.FollowInfo != nil {
		cd.FollowButton = &api.OgvFollowButton{
			IsFollowed:   sc.FollowInfo.IsFollow == 1,
			FollowText:   sc.FollowInfo.FollowText,
			FollowIcon:   sc.FollowInfo.FollowIcon,
			UnfollowText: sc.FollowInfo.UnfollowText,
			UnfollowIcon: sc.FollowInfo.UnfollowIcon,
		}
		cd.FollowButton.FollowParams = followOgvParams(sc.SeasonId, cd.FollowButton.IsFollowed)
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeOgvThree.String(),
		CardId:     strconv.FormatInt(int64(sc.SeasonId), 10),
		CardDetail: &api.ModuleItem_OgvThreeCard{OgvThreeCard: cd},
	}
}

func (bu Ogv) buildOgvOne(cfg *config.OgvCommon, sc *pgcappgrpc.SeasonCardInfoProto, rcmdContent string) *api.ModuleItem {
	cd := &api.OgvOneCard{
		Position1: "-观看",
		Image:     sc.Cover,
		Title:     sc.Title,
		ReportDic: &api.ReportDic{
			BizType:    model.ReportBizTypePGC,
			SeasonType: strconv.FormatInt(int64(sc.SeasonType), 10),
			SeasonId:   int64(sc.SeasonId),
		},
		Url:          sc.Url,
		ResourceType: model.ResourceTypeOGV,
	}
	if sc.Stat != nil {
		cd.Position1 = appcardmdl.Stat64String(sc.Stat.View, "观看")
		cd.Position2 = sc.Stat.FollowView
	}
	if sc.NewEp != nil {
		cd.Position3 = sc.NewEp.IndexShow
	}
	if cfg.DisplayScore && sc.Rating != nil {
		cd.CoverRightText1 = fmt.Sprintf("%0.1f", sc.Rating.Score)
		cd.CoverRightText2 = "分"
	}
	if cfg.DisplayRcmd {
		cd.RcmdContent = rcmdContent
		if cd.RcmdContent == "" {
			cd.RcmdContent = sc.Subtitle
		}
		if cd.RcmdContent != "" {
			cd.RcmdIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/rYGIuJ~Ii4.png"
		}
	}
	if cfg.DisplayPayBadge && sc.BadgeInfo != nil {
		cd.Badge = &api.Badge{
			Text:         sc.BadgeInfo.Text,
			BgColor:      sc.BadgeInfo.BgColor,
			BgColorNight: sc.BadgeInfo.BgColorNight,
		}
	}
	if sc.FollowInfo != nil {
		cd.FollowButton = &api.OgvFollowButton{
			IsFollowed:   sc.FollowInfo.IsFollow == 1,
			FollowText:   sc.FollowInfo.FollowText,
			FollowIcon:   sc.FollowInfo.FollowIcon,
			UnfollowText: sc.FollowInfo.UnfollowText,
			UnfollowIcon: sc.FollowInfo.UnfollowIcon,
		}
		cd.FollowButton.FollowParams = followOgvParams(sc.SeasonId, cd.FollowButton.IsFollowed)
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeOgvOne.String(),
		CardId:     strconv.FormatInt(int64(sc.SeasonId), 10),
		CardDetail: &api.ModuleItem_OgvOneCard{OgvOneCard: cd},
	}
}

func buildOgvMore(cfg *config.OgvCommon, lastIndex, moduleID int64) *api.ModuleItem {
	moreText := cfg.ViewMoreText
	if moreText == "" {
		moreText = "查看更多"
	}
	params := ogvSupernatantParams(0, lastIndex, moduleID)
	return card.NewOgvMore(moreText, cfg.SupernatantTitle, params).Build()
}

func ogvSupernatantParams(offset, lastIndex, moduleID int64) string {
	return passthrough.Marshal(&api.OgvSupernatantParams{LastIndex: lastIndex, Offset: offset, ModuleId: moduleID})
}

func followOgvParams(seasonID int32, isFollowed bool) string {
	action := api.ActionType_Do
	if isFollowed {
		action = api.ActionType_Undo
	}
	return passthrough.Marshal(&api.FollowOgvParams{Action: action, SeasonId: seasonID})
}
