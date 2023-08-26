package builder

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type VideoID struct{}

func (bu VideoID) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	viCfg, ok := cfg.(*config.VideoID)
	if !ok {
		logCfgAssertionError(config.VideoID{})
		return nil
	}
	offset, hasMore, items := bu.buildModuleItems(viCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeVideo.String(),
		ModuleId:      viCfg.ModuleBase().ModuleID,
		ModuleColor:   buildModuleColorOfVideo(&viCfg.VideoCommon),
		ModuleSetting: &api.Setting{DisplayTitle: !viCfg.HideTitle, AutoPlay: viCfg.AutoPlay},
		ModuleItems:   items,
		ModuleUkey:    viCfg.ModuleBase().Ukey,
	}
	if viCfg.DisplayViewMore && hasMore {
		module.HasMore = true
		subpageParams := subpageParamsOfVideo(module.ModuleId, 0, offset, "")
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfVideo(module.ModuleId, viCfg.PageID, offset, "",
				func() *api.SubpageData {
					return buildSubpageData(viCfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}

func (bu VideoID) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu VideoID) buildModuleItems(cfg *config.VideoID, material *kernel.Material, ss *kernel.Session) (int64, bool, []*api.ModuleItem) {
	mixRly, ok := material.MixExtsRlys[cfg.MixExtsReqID]
	if !ok || len(mixRly.List) == 0 {
		return 0, false, nil
	}
	items := make([]*api.ModuleItem, 0, cfg.Ps)
	offset := ss.Offset
	for _, ext := range mixRly.List {
		if int64(len(items)) >= cfg.Ps {
			break
		}
		offset++
		var cd *api.VideoCard
		switch ext.MType {
		case natpagegrpc.MixAvidType:
			arcPlayer, ok := material.ArcsPlayer[ext.ForeignID]
			if !ok || arcPlayer.GetArc() == nil || !arcPlayer.GetArc().IsNormal() {
				continue
			}
			cd = bu.buildArchive(arcPlayer, ss)
		case natpagegrpc.MixEpidType:
			ep, ok := material.Episodes[ext.ForeignID]
			if !ok {
				continue
			}
			cd = bu.buildEpisode(ep)
		default:
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeVideo.String(),
			CardId:     strconv.FormatInt(ext.ForeignID, 10),
			CardDetail: &api.ModuleItem_VideoCard{VideoCard: cd},
		})
	}
	if len(items) == 0 {
		return 0, false, nil
	}
	hasMore := mixRly.HasMore == 1
	if !hasMore && offset < mixRly.Offset {
		hasMore = true
	}
	items = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	return offset, hasMore, items
}

func (bu VideoID) buildArchive(arcPlayer *arcgrpc.ArcPlayer, ss *kernel.Session) *api.VideoCard {
	arc := arcPlayer.GetArc()
	cd := &api.VideoCard{
		Title:          arc.GetTitle(),
		CoverImageUri:  arc.GetPic(),
		CoverLeftText1: appcardmdl.DurationString(arc.GetDuration()),
		CoverLeftText2: appcardmdl.StatString(arc.GetStat().View, "观看"),
		CoverLeftText3: appcardmdl.StatString(arc.GetStat().Danmaku, "弹幕"),
		ReportDic:      &api.ReportDic{Aid: arc.GetAid(), Cid: arc.GetFirstCid(), AuthorName: arc.GetAuthor().Name},
		ResourceType:   model.ResourceTypeUGC,
	}
	playerInfo := arcPlayer.GetPlayerInfo()[arcPlayer.GetDefaultPlayerCid()]
	cd.Uri = appcardmdl.FillURI(appcardmdl.GotoAv, ss.RawDevice().Plat(), int(ss.RawDevice().Build), strconv.FormatInt(arc.GetAid(), 10),
		appcardmdl.ArcPlayHandler(arc, playerInfo, ss.TraceId(), nil, int(ss.RawDevice().Build), ss.RawDevice().RawMobiApp, true))
	cd.Rights = &api.VideoRights{
		UgcPay:        arc.GetRights().UGCPay == arcgrpc.AttrYes,
		IsCooperation: arc.GetRights().IsCooperation == arcgrpc.AttrYes,
		IsPgc:         arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes,
	}
	if dimension := playerInfo.GetPlayerExtra().GetDimension(); dimension != nil {
		cd.Dimension = &api.PlayerDimension{
			Width:  dimension.GetWidth(),
			Height: dimension.GetHeight(),
			Rotate: dimension.GetRotate() == int64(arcgrpc.AttrYes),
		}
	}
	cd.ReportDic = &api.ReportDic{Aid: arc.GetAid(), Cid: arc.GetFirstCid(), AuthorName: arc.GetAuthor().Name}
	return cd
}

func (bu VideoID) buildEpisode(ep *model.EpPlayer) *api.VideoCard {
	cd := &api.VideoCard{
		Title:          ep.ShowTitle,
		CoverImageUri:  ep.Cover,
		CoverLeftText1: appcardmdl.DurationString(ep.Duration),
		CoverLeftText2: appcardmdl.Stat64String(ep.Stat.Play, "观看"),
		CoverLeftText3: appcardmdl.Stat64String(ep.Stat.Danmaku, "弹幕"),
		Uri:            ep.Uri,
		Badge:          &api.Badge{Text: ep.Season.TypeName},
		ResourceType:   model.ResourceTypeOGV,
	}
	cd.ReportDic = &api.ReportDic{
		BizType:    model.ReportBizTypePGC,
		SeasonType: strconv.FormatInt(ep.Season.Type, 10),
		Aid:        ep.AID,
		Cid:        ep.CID,
		EpId:       ep.EpID,
		IsPreview:  int32(ep.IsPreview),
		SeasonId:   ep.Season.SeasonID,
	}
	return cd
}

func buildModuleColorOfVideo(cfg *config.VideoCommon) *api.Color {
	return &api.Color{
		BgColor:            cfg.BgColor,
		TitleColor:         cfg.TitleColor,
		CardTitleFontColor: cfg.CardTitleFontColor,
		ViewMoreFontColor:  cfg.ViewMoreFontColor,
		ViewMoreBgColor:    cfg.ViewMoreBgColor,
	}
}

func buildMoreCardOfVideo(moduleID, pageID, offset int64, topicOffset string, newSpData func() *api.SubpageData) *api.ModuleItem {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("dy_offset", topicOffset)
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	uri := fmt.Sprintf("bilibili://following/activity_detail/%d?%s", moduleID, params.Encode())
	var subpageData *api.SubpageData
	if newSpData != nil {
		subpageData = newSpData()
	}
	return card.NewVideoMore("查看更多", uri, subpageData).Build()
}

func subpageParamsOfVideo(moduleID, sort, offset int64, topicOffset string) string {
	return passthrough.Marshal(&api.VideoParams{Offset: offset, TopicOffset: topicOffset, ModuleId: moduleID, SortType: sort})
}
