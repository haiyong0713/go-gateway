package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	_defaultTopIcon    = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/q~1vlO6h25.png"
	_defaultBottomIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/rYGIuJ~Ii4.png"
)

type Editor struct{}

func (bu Editor) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	editorCfg, ok := cfg.(*config.Editor)
	if !ok {
		logCfgAssertionError(config.Editor{})
		return nil
	}
	mixRly, ok := material.MixExtsRlys[editorCfg.MixExtsReqID]
	if !ok || len(mixRly.List) == 0 {
		return nil
	}
	items := bu.buildModuleItems(editorCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeEditor.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleColor:   bu.buildModuleColor(editorCfg, mixRly.List[0]),
		ModuleSetting: bu.buildModuleSetting(editorCfg),
		ModuleItems:   items,
		ModuleUkey:    cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Editor) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Editor) buildModuleItems(cfg *config.Editor, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	viewedArcs := buildViewedArcs(material.GetHisRlys[cfg.GetHisReqID])
	mixExts := material.MixExtsRlys[cfg.MixExtsReqID].List
	items := make([]*api.ModuleItem, 0, len(mixExts))
	for _, ext := range mixExts {
		mixEditor, _ := model.UnmarshalMixExtEditor(ext.Reason)
		var cd *api.EditorRecommendCard
		switch ext.MType {
		case natpagegrpc.MixAvidType, natpagegrpc.MixFolder:
			arc, ok := material.Arcs[ext.ForeignID]
			if !ok || !arc.IsNormal() {
				continue
			}
			var folder *favmdl.Folder
			if ext.MType == natpagegrpc.MixFolder {
				if folder, ok = material.Folders[favmdl.TypeVideo][mixEditor.Fid]; !ok {
					continue
				}
			}
			cd = bu.buildArchiveFolder(mixEditor, &cfg.Position, ss, arc, folder, viewedArcs)
		case natpagegrpc.MixEpidType:
			ep, ok := material.Episodes[ext.ForeignID]
			if !ok {
				continue
			}
			cd = bu.buildEpisode(mixEditor, &cfg.Position, ep)
		case natpagegrpc.MixCvidType:
			art, ok := material.Articles[ext.ForeignID]
			if !ok {
				continue
			}
			cd = bu.buildArticle(mixEditor, &cfg.Position, art)
		default:
			log.Warn("unknown material_type=%d", ext.MType)
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeEditor.String(),
			CardId:     strconv.FormatInt(ext.ForeignID, 10),
			CardDetail: &api.ModuleItem_EditorRecommendCard{EditorRecommendCard: cd},
		})
	}
	return items
}

func (bu Editor) buildModuleColor(cfg *config.Editor, mixExt *natpagegrpc.NativeMixtureExt) *api.Color {
	mixEditor, _ := model.UnmarshalMixExtEditor(mixExt.Reason)
	return &api.Color{
		BgColor:         cfg.BgColor,
		TopFontColor:    mixEditor.RcmdContent.TopFontColor,
		BottomFontColor: mixEditor.RcmdContent.BottomFontColor,
	}
}

func (bu Editor) buildModuleSetting(cfg *config.Editor) *api.Setting {
	return &api.Setting{DisplayMoreButton: cfg.DisplayMoreButton}
}

func (bu Editor) buildArchiveFolder(mixEditor *model.MixExtEditor, pos *config.Position, ss *kernel.Session, arc *arcgrpc.Arc, folder *favmdl.Folder, viewedArcs map[int64]struct{}) *api.EditorRecommendCard {
	cd := &api.EditorRecommendCard{}
	cd.Position1 = bu.transArcPosition(pos.Position1, arc, viewedArcs)
	cd.Position2 = bu.transArcPosition(pos.Position2, arc, viewedArcs)
	cd.Position3 = bu.transArcPosition(pos.Position3, arc, viewedArcs)
	cd.Position4 = bu.transArcPosition(pos.Position4, arc, viewedArcs)
	cd.Position5 = bu.transArcPosition(pos.Position5, arc, viewedArcs)
	setCommonEditorCard(cd, mixEditor)
	cd.Title = arc.GetTitle()
	cd.CoverImageUri = arc.GetPic()
	switch {
	case ss.IsIOS(), ss.IsAndroid(), ss.IsIPad():
		if folder == nil || ss.IsIPad() {
			cd.Uri = appcardmdl.FillURI(appcardmdl.GotoAv, ss.RawDevice().Plat(), int(ss.RawDevice().Build), strconv.FormatInt(arc.GetAid(), 10),
				appcardmdl.ArcPlayHandler(arc, nil, ss.TraceId(), nil, int(ss.RawDevice().Build), ss.RawDevice().RawMobiApp, false))
		} else {
			cd.Uri = appcardmdl.FillURI(appcardmdl.GotoPlaylist, ss.RawDevice().Plat(), int(ss.RawDevice().Build),
				fmt.Sprintf("%d?avid=%d&oid=%d&page_type=4", folder.Mlid, arc.GetAid(), arc.GetAid()), nil)
		}
	case ss.IsH5():
		bvID, _ := bvid.AvToBv(arc.GetAid())
		if folder == nil {
			cd.Uri = fmt.Sprintf("https://m.bilibili.com/video/%s", bvID)
		} else {
			cd.Uri = fmt.Sprintf("https://m.bilibili.com/playlist/pl%d?bvid=%s&oid=%d", folder.Mlid, bvID, arc.GetAid())
		}
	case ss.IsWeb():
		bvID, _ := bvid.AvToBv(arc.GetAid())
		if folder == nil {
			cd.Uri = fmt.Sprintf("https://www.bilibili.com/video/%s", bvID)
		} else {
			cd.Uri = fmt.Sprintf("https://www.bilibili.com/medialist/play/ml%d/%s", folder.Mlid, bvID)
		}
	}
	cd.Share = &api.Share{
		DisplayLater: true,
		Oid:          arc.GetAid(),
		ShareOrigin:  model.ShareOriginUGC,
		ShareType:    model.ShareTypeActivity,
	}
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeUGC}
	cd.ResourceType = model.ResourceTypeUGC
	return cd
}

func (bu Editor) buildEpisode(mixEditor *model.MixExtEditor, pos *config.Position, ep *model.EpPlayer) *api.EditorRecommendCard {
	cd := &api.EditorRecommendCard{}
	cd.Position1 = bu.transEpPosition(pos.Position1, ep)
	cd.Position2 = bu.transEpPosition(pos.Position2, ep)
	cd.Position3 = bu.transEpPosition(pos.Position3, ep)
	cd.Position4 = bu.transEpPosition(pos.Position4, ep)
	cd.Position5 = bu.transEpPosition(pos.Position5, ep)
	setCommonEditorCard(cd, mixEditor)
	// PGC不下发三点操作
	cd.Setting = &api.Setting{DisplayMoreButton: false}
	cd.Title = ep.ShowTitle
	cd.CoverImageUri = ep.Cover
	cd.Uri = ep.Uri
	cd.Share = &api.Share{ShareType: model.ShareTypeActivity}
	cd.ReportDic = &api.ReportDic{
		BizType:    model.ReportBizTypePGC,
		SeasonType: strconv.FormatInt(ep.Season.Type, 10),
	}
	cd.ResourceType = model.ResourceTypeOGV
	return cd
}

func (bu Editor) buildArticle(mixEditor *model.MixExtEditor, pos *config.Position, art *articlemdl.Meta) *api.EditorRecommendCard {
	cd := &api.EditorRecommendCard{}
	cd.Position1 = bu.transArticlePosition(pos.Position1, art)
	cd.Position2 = bu.transArticlePosition(pos.Position2, art)
	cd.Position3 = bu.transArticlePosition(pos.Position3, art)
	cd.Position4 = bu.transArticlePosition(pos.Position4, art)
	cd.Position5 = bu.transArticlePosition(pos.Position5, art)
	setCommonEditorCard(cd, mixEditor)
	cd.Title = art.Title
	if len(art.ImageURLs) > 0 {
		cd.CoverImageUri = art.ImageURLs[0]
	}
	cd.Uri = fmt.Sprintf("https://www.bilibili.com/read/cv%d", art.ID)
	cd.Share = &api.Share{
		Oid:         art.ID,
		ShareOrigin: model.ShareOriginArticle,
		ShareType:   model.ShareTypeActivity,
	}
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeArticle}
	cd.ResourceType = model.ResourceTypeArticle
	return cd
}

func (bu Editor) transArcPosition(position string, arc *arcgrpc.Arc, viewedArcs map[int64]struct{}) string {
	switch position {
	case model.PositionUp:
		return arc.GetAuthor().Name
	case model.PositionView:
		return appcardmdl.StatString(arc.GetStat().View, "观看")
	case model.PositionPubTime:
		return appcardmdl.PubDataString(arc.GetPubDate().Time())
	case model.PositionLike:
		return appcardmdl.StatString(arc.GetStat().Like, "点赞")
	case model.PositionDanmaku:
		return appcardmdl.StatString(arc.GetStat().Danmaku, "弹幕")
	case model.PositionViewStat:
		if _, ok := viewedArcs[arc.GetAid()]; ok {
			return "已观看"
		}
	}
	return ""
}

func (bu Editor) transEpPosition(position string, ep *model.EpPlayer) string {
	switch position {
	case model.PositionDuration:
		return appcardmdl.DurationString(ep.Duration)
	case model.PositionView:
		return appcardmdl.Stat64String(ep.Stat.Play, "观看")
	case model.PositionFollow:
		return appcardmdl.Stat64String(ep.Stat.Follow, "追剧")
	}
	return ""
}

func (bu Editor) transArticlePosition(position string, art *articlemdl.Meta) string {
	switch position {
	case model.PositionUp:
		if art.Author == nil {
			return ""
		}
		return art.Author.Name
	case model.PositionView:
		if art.Stats == nil {
			return ""
		}
		return appcardmdl.Stat64String(art.Stats.View, "观看")
	case model.PositionLike:
		if art.Stats == nil {
			return ""
		}
		return appcardmdl.Stat64String(art.Stats.Like, "点赞")
	case model.PositionPubTime:
		return appcardmdl.PubDataString(art.PublishTime.Time())
	}
	return ""
}

func setCommonEditorCard(cd *api.EditorRecommendCard, mixEditor *model.MixExtEditor) {
	if mixEditor == nil {
		return
	}
	if mixEditor.RcmdContent.TopContent != "" {
		cd.TopIcon = _defaultTopIcon
		cd.TopContent = mixEditor.RcmdContent.TopContent
	}
	if mixEditor.RcmdContent.BottomContent != "" {
		cd.BottomIcon = _defaultBottomIcon
		cd.BottomContent = mixEditor.RcmdContent.BottomContent
	}
	cd.MiddleIcon = mixEditor.RcmdContent.MiddleIcon
}

func buildViewedArcs(rly *actplatv2grpc.GetHistoryResp) map[int64]struct{} {
	if rly == nil {
		return map[int64]struct{}{}
	}
	viewedArcs := make(map[int64]struct{}, len(rly.GetHistory()))
	for _, his := range rly.GetHistory() {
		if his == nil || his.Source == "" {
			continue
		}
		var source struct {
			Aid int64 `json:"aid"`
		}
		if err := json.Unmarshal([]byte(his.Source), &source); err != nil {
			log.Error("Fail to unmarshal HistoryContent.Source, source=%s error=%+v", his.Source, err)
			continue
		}
		viewedArcs[source.Aid] = struct{}{}
	}
	return viewedArcs
}
