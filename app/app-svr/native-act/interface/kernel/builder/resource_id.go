package builder

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"

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

type ResourceID struct{}

func (bu ResourceID) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	riCfg, ok := cfg.(*config.ResourceID)
	if !ok {
		logCfgAssertionError(config.ResourceID{})
		return nil
	}
	hasMore, offset, items := bu.buildModuleItems(riCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeResource.String(),
		ModuleId:    riCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfResource(&riCfg.ResourceCommon),
		ModuleItems: items,
		ModuleUkey:  riCfg.ModuleBase().Ukey,
	}
	if riCfg.DisplayViewMore && hasMore {
		module.HasMore = true
		subpageParams := subpageParamsOfResource(module.ModuleId, 0, offset, "")
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, riCfg.PageID, offset, "",
				func() *api.SubpageData {
					return buildSubpageData(riCfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}

func (bu ResourceID) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu ResourceID) buildModuleItems(cfg *config.ResourceID, material *kernel.Material, ss *kernel.Session) (bool, int64, []*api.ModuleItem) {
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
		mixFolder, _ := model.UnmarshalMixExtResourceFolder(ext.Reason)
		var cd *api.ResourceCard
		switch ext.MType {
		case natpagegrpc.MixAvidType, natpagegrpc.MixFolder:
			arc, ok := material.Arcs[ext.ForeignID]
			if !ok || !arc.IsNormal() {
				continue
			}
			var folder *favmdl.Folder
			if ext.MType == natpagegrpc.MixFolder {
				if folder, ok = material.Folders[favmdl.TypeVideo][mixFolder.Fid]; !ok {
					continue
				}
			}
			cd = bu.buildArchiveFolder(arc, folder, ss, cfg.DisplayUGCBadge)
		case natpagegrpc.MixEpidType:
			ep, ok := material.Episodes[ext.ForeignID]
			if !ok {
				continue
			}
			cd = bu.buildEpisode(ep, cfg.DisplayPGCBadge)
		case natpagegrpc.MixCvidType:
			art, ok := material.Articles[ext.ForeignID]
			if !ok || !art.IsNormal() {
				continue
			}
			cd = bu.buildArticle(art, cfg.DisplayArticleBadge)
		case natpagegrpc.MixLive:
			var isLive int64
			if cfg.DisplayOnlyLive {
				isLive = 1
			}
			if _, ok := material.LiveRooms[isLive]; !ok {
				continue
			}
			room, ok := material.LiveRooms[isLive][ext.ForeignID]
			if !ok {
				continue
			}
			cd = bu.buildLive(room, ss)
		default:
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(ext.ForeignID, 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
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

func (bu ResourceID) buildArchiveFolder(arc *arcgrpc.Arc, folder *favmdl.Folder, ss *kernel.Session, displayBadge bool) *api.ResourceCard {
	cd := &api.ResourceCard{
		Title:          arc.GetTitle(),
		CoverImageUri:  arc.GetPic(),
		CoverRightText: appcardmdl.DurationString(arc.GetDuration()),
		CoverLeftText1: appcardmdl.StatString(arc.GetStat().View, ""),
		CoverLeftIcon1: int64(appcardmdl.IconPlay),
		CoverLeftText2: appcardmdl.StatString(arc.GetStat().Danmaku, ""),
		CoverLeftIcon2: int64(appcardmdl.IconDanmaku),
		ReportDic:      &api.ReportDic{BizType: model.ReportBizTypeUGC},
		ResourceType:   model.ResourceTypeUGC,
	}
	if folder == nil {
		if arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && arc.GetRedirectURL() != "" {
			cd.Uri = arc.GetRedirectURL()
		} else {
			cd.Uri = appcardmdl.FillURI(appcardmdl.GotoAv, ss.RawDevice().Plat(), int(ss.RawDevice().Build), strconv.FormatInt(arc.GetAid(), 10),
				appcardmdl.ArcPlayHandler(arc, nil, ss.TraceId(), nil, int(ss.RawDevice().Build), ss.RawDevice().RawMobiApp, false))
		}
	} else {
		cd.Uri = appcardmdl.FillURI(appcardmdl.GotoPlaylist, ss.RawDevice().Plat(), int(ss.RawDevice().Build),
			fmt.Sprintf("%d?avid=%d&oid=%d&page_type=4", folder.Mlid, arc.GetAid(), arc.GetAid()), nil)
	}
	if displayBadge {
		cd.Badge = &api.Badge{Text: "视频"}
	}
	return cd
}

func (bu ResourceID) buildEpisode(ep *model.EpPlayer, displayBadge bool) *api.ResourceCard {
	cd := &api.ResourceCard{
		Title:          ep.ShowTitle,
		CoverImageUri:  ep.Cover,
		Uri:            ep.Uri,
		CoverRightText: appcardmdl.DurationString(ep.Duration),
		CoverLeftText1: appcardmdl.Stat64String(ep.Stat.Play, ""),
		CoverLeftIcon1: int64(appcardmdl.IconPlay),
		CoverLeftText2: appcardmdl.Stat64String(ep.Stat.Follow, ""),
		CoverLeftIcon2: int64(appcardmdl.IconFavorite),
		ReportDic:      &api.ReportDic{BizType: model.ReportBizTypePGC, SeasonType: strconv.FormatInt(ep.Season.Type, 10)},
		ResourceType:   model.ResourceTypeOGV,
	}
	if displayBadge {
		cd.Badge = &api.Badge{Text: ep.Season.TypeName}
	}
	return cd
}

func (bu ResourceID) buildArticle(art *articlemdl.Meta, displayBadge bool) *api.ResourceCard {
	cd := &api.ResourceCard{
		Title:          art.Title,
		Uri:            fmt.Sprintf("https://www.bilibili.com/read/cv%d", art.ID),
		CoverLeftIcon1: int64(appcardmdl.IconRead),
		CoverLeftIcon2: int64(appcardmdl.IconComment),
		ReportDic:      &api.ReportDic{BizType: model.ReportBizTypeArticle},
		ResourceType:   model.ResourceTypeArticle,
	}
	if len(art.ImageURLs) > 0 {
		cd.CoverImageUri = art.ImageURLs[0]
	}
	if art.Stats != nil {
		cd.CoverLeftText1 = appcardmdl.Stat64String(art.Stats.View, "")
		cd.CoverLeftText2 = appcardmdl.Stat64String(art.Stats.Reply, "")
	}
	if displayBadge {
		cd.Badge = &api.Badge{Text: "文章"}
	}
	return cd
}

func (bu ResourceID) buildLive(live *liveplaygrpc.RoomList, ss *kernel.Session) *api.ResourceCard {
	cd := &api.ResourceCard{
		Title:          live.Title,
		CoverImageUri:  live.Icon,
		Uri:            fmt.Sprintf("https://live.bilibili.com/%d", live.RoomId),
		CoverRightText: live.UserName,
		ReportDic:      &api.ReportDic{BizType: model.ReportBizTypeLive},
		ResourceType:   model.ResourceTypeLive,
	}
	if live.Pendant != "" {
		cd.Badge = &api.Badge{Text: live.Pendant}
	}
	cd.CoverLeftText1, cd.CoverLeftIcon1 = func() (string, int64) {
		if !enableLiveWatched(ss.RawDevice().MobiApp(), ss.RawDevice().Device, ss.RawDevice().Build) {
			return live.Online, int64(appcardmdl.IconOnline)
		}
		if ws := live.WatchedShow; ws != nil && ws.Switch {
			return ws.TextLarge, int64(appcardmdl.IconLiveWatched)
		}
		return live.Online, int64(appcardmdl.IconLiveOnline)
	}()
	return cd
}

func buildModuleColorOfResource(cfg *config.ResourceCommon) *api.Color {
	return &api.Color{
		BgColor:            cfg.BgColor,
		TitleColor:         cfg.TitleColor,
		CardTitleFontColor: cfg.CardTitleFontColor,
		CardTitleBgColor:   cfg.CardTitleBgColor,
		ViewMoreFontColor:  cfg.ViewMoreFontColor,
		ViewMoreBgColor:    cfg.ViewMoreBgColor,
	}
}

func buildMoreCardOfResource(moduleID, pageID, offset int64, topicOffset string, newSpData func() *api.SubpageData) *api.ModuleItem {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("dy_offset", topicOffset)
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	uri := fmt.Sprintf("bilibili://following/activity_detail/%d?%s", moduleID, params.Encode())
	var subpageData *api.SubpageData
	if newSpData != nil {
		subpageData = newSpData()
	}
	return card.NewResourceMore("查看更多", uri, subpageData).Build()
}

func subpageParamsOfResource(moduleID, sort, offset int64, topicOffset string) string {
	return passthrough.Marshal(&api.ResourceParams{Offset: offset, TopicOffset: topicOffset, ModuleId: moduleID, SortType: sort})
}

func enableLiveWatched(mobiApp, device string, build int64) bool {
	return (mobiApp == "android" && build >= 6610000) ||
		(mobiApp == "iphone" && device == "phone" && build >= 66100000) ||
		(mobiApp == "iphone" && device == "pad" && build >= 66200000) ||
		(mobiApp == "ipad" && build >= 33600000)
}
