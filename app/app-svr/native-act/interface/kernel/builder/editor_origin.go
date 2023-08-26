package builder

import (
	"context"
	"strconv"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/pkg/idsafe/bvid"
)

type EditorOrigin struct{}

func (bu EditorOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	editorCfg, ok := cfg.(*config.EditorOrigin)
	if !ok {
		logCfgAssertionError(config.EditorOrigin{})
		return nil
	}
	if editorCfg.IsFeed && (editorCfg.RdbType == model.RDBMustsee || editorCfg.RdbType == model.RDBGAT) {
		if model.IsFromIndex(ss.ReqFrom) {
			return bu.buildFromFeedIndex(editorCfg)
		}
	}
	switch editorCfg.RdbType {
	case model.RDBMustsee:
		return bu.buildModuleOfMustsee(editorCfg, material, ss)
	case model.RDBWeek:
		return bu.buildModuleOfWeek(editorCfg, material, ss)
	case model.RDBRank:
		return bu.buildModuleOfRank(editorCfg, material)
	case model.RDBGAT:
		return bu.buildModuleOfGAT(editorCfg, material, ss)
	}
	return nil
}

func (bu EditorOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu EditorOrigin) buildFromFeedIndex(cfg *config.EditorOrigin) *api.Module {
	module := bu.buildModuleBase(cfg)
	module.SubpageParams = bu.subpageParams(cfg, 0)
	return module
}

func (bu EditorOrigin) buildModuleBase(cfg *config.EditorOrigin) *api.Module {
	return &api.Module{
		ModuleType:    model.ModuleTypeEditor.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleColor:   bu.buildModuleColor(cfg),
		ModuleSetting: bu.buildModuleSetting(cfg),
		ModuleUkey:    cfg.ModuleBase().Ukey,
		IsFeed:        cfg.IsFeed,
	}
}

func (bu EditorOrigin) buildModuleColor(cfg *config.EditorOrigin) *api.Color {
	return &api.Color{BgColor: cfg.BgColor}
}

func (bu EditorOrigin) buildModuleSetting(cfg *config.EditorOrigin) *api.Setting {
	return &api.Setting{DisplayMoreButton: cfg.DisplayMoreButton}
}

func (bu EditorOrigin) buildModuleOfMustsee(cfg *config.EditorOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	pageArcs, ok := material.PageArcsRlys[cfg.PageArcsReqID]
	if !ok || len(pageArcs.List) == 0 {
		return nil
	}
	folder := material.Folders[favmdl.TypeVideo][pageArcs.GetMediaId()]
	viewedArcs := buildViewedArcs(material.GetHisRlys[cfg.GetHisReqID])
	edBuilder := Editor{}
	items := make([]*api.ModuleItem, 0, len(pageArcs.GetList()))
	for _, pageArc := range pageArcs.GetList() {
		arc, ok := material.Arcs[pageArc.GetAid()]
		if !ok || !arc.IsNormal() {
			continue
		}
		mixEditor := &model.MixExtEditor{RcmdContent: model.RcmdContent{TopContent: pageArc.Recommend}}
		cd := edBuilder.buildArchiveFolder(mixEditor, &cfg.Position, ss, arc, folder, viewedArcs)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeEditor.String(),
			CardId:     strconv.FormatInt(arc.GetAid(), 10),
			CardDetail: &api.ModuleItem_EditorRecommendCard{EditorRecommendCard: cd},
		})
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = items
	if pageArcs.GetPage() != nil && pageArcs.GetPage().GetHasMore() == 1 {
		module.HasMore = true
		module.SubpageParams = bu.subpageParams(cfg, pageArcs.GetPage().GetOffset())
	}
	return module
}

func (bu EditorOrigin) buildModuleOfWeek(cfg *config.EditorOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	selSerie, ok := material.SelSerieRlys[cfg.SelSerieReqID]
	if !ok || len(selSerie.List) == 0 {
		return nil
	}
	var folder *favmdl.Folder
	if selSerie.Config != nil {
		folder = material.Folders[favmdl.TypeVideo][selSerie.Config.MediaId]
	}
	viewedArcs := buildViewedArcs(material.GetHisRlys[cfg.GetHisReqID])
	edBuilder := Editor{}
	items := make([]*api.ModuleItem, 0, len(selSerie.List))
	for _, selRes := range selSerie.List {
		if selRes == nil || selRes.Rid <= 0 || selRes.Rtype != model.SelRtypeArchive {
			continue
		}
		arc, ok := material.Arcs[selRes.Rid]
		if !ok {
			continue
		}
		mixEditor := &model.MixExtEditor{RcmdContent: model.RcmdContent{BottomContent: selRes.RcmdReason}}
		cd := edBuilder.buildArchiveFolder(mixEditor, &cfg.Position, ss, arc, folder, viewedArcs)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeEditor.String(),
			CardId:     strconv.FormatInt(arc.GetAid(), 10),
			CardDetail: &api.ModuleItem_EditorRecommendCard{EditorRecommendCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = items
	return module
}

func (bu EditorOrigin) buildModuleOfRank(cfg *config.EditorOrigin, material *kernel.Material) *api.Module {
	rankRst, ok := material.RankRstRlys[cfg.RankRstReqID]
	if !ok || len(rankRst.List) == 0 {
		return nil
	}
	mixEditors := bu.buildMixEditors(material.MixExtRlys[cfg.MixExtReqID])
	items := make([]*api.ModuleItem, 0, len(rankRst.List))
	for k, v := range rankRst.List {
		if v == nil || v.ObjectType != 2 || len(v.Archive) == 0 || v.Archive[0] == nil {
			continue
		}
		arc := v.Archive[0]
		aid, _ := bvid.BvToAv(arc.BvID)
		cd := bu.buildRankArchive(mixEditors[k], &cfg.Position, v, arc)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeEditor.String(),
			CardId:     strconv.FormatInt(aid, 10),
			CardDetail: &api.ModuleItem_EditorRecommendCard{EditorRecommendCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = items
	return module
}

func (bu EditorOrigin) buildRankArchive(mixEditor *model.MixExtEditor, pos *config.Position, rankRst *activitygrpc.RankResult, arc *activitygrpc.ArchiveInfo) *api.EditorRecommendCard {
	cd := &api.EditorRecommendCard{}
	cd.Position1 = bu.transRankArcPosition(pos.Position1, rankRst, arc)
	cd.Position2 = bu.transRankArcPosition(pos.Position2, rankRst, arc)
	cd.Position3 = bu.transRankArcPosition(pos.Position3, rankRst, arc)
	cd.Position4 = bu.transRankArcPosition(pos.Position4, rankRst, arc)
	cd.Position5 = bu.transRankArcPosition(pos.Position5, rankRst, arc)
	setCommonEditorCard(cd, mixEditor)
	cd.Title = arc.Title
	cd.CoverImageUri = arc.Pic
	cd.Uri = arc.ShowLink
	aid, _ := bvid.BvToAv(arc.BvID)
	cd.Share = &api.Share{
		DisplayLater: true,
		Oid:          aid,
		ShareOrigin:  model.ShareOriginUGC,
		ShareType:    model.ShareTypeActivity,
	}
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeUGC}
	cd.ResourceType = model.ResourceTypeUGC
	return cd
}

func (bu EditorOrigin) transRankArcPosition(position string, rankRst *activitygrpc.RankResult, arc *activitygrpc.ArchiveInfo) string {
	switch position {
	case model.PositionComprehensive:
		return rankRst.ShowScore
	case model.PositionLike:
		return appcardmdl.Stat64String(arc.Like, "点赞")
	case model.PositionView:
		return appcardmdl.Stat64String(arc.View, "观看")
	case model.PositionShare:
		return appcardmdl.Stat64String(arc.Share, "分享")
	case model.PositionCoin:
		return appcardmdl.Stat64String(arc.Coin, "投币")
	}
	return ""
}

func (bu EditorOrigin) subpageParams(cfg *config.EditorOrigin, offset int64) string {
	return passthrough.Marshal(&api.EditorParams{Offset: offset, ModuleId: cfg.ModuleBase().ModuleID})
}

func (bu EditorOrigin) buildMixEditors(mixExtRly *natpagegrpc.ModuleMixExtReply) map[int]*model.MixExtEditor {
	if mixExtRly == nil {
		return map[int]*model.MixExtEditor{}
	}
	mixEditors := make(map[int]*model.MixExtEditor, len(mixExtRly.List))
	for k, v := range mixExtRly.List {
		if v == nil {
			continue
		}
		if mixEditor, err := model.UnmarshalMixExtEditor(v.Reason); err == nil {
			mixEditors[k] = mixEditor
		}
	}
	return mixEditors
}

func (bu EditorOrigin) buildModuleOfGAT(cfg *config.EditorOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	feedRly, ok := material.ChannelFeedRlys[cfg.ChannelFeedReqID]
	if !ok || len(feedRly.GetList()) == 0 {
		return nil
	}
	edBuilder := Editor{}
	pgcPosition := &config.Position{Position2: model.PositionDuration, Position4: model.PositionView, Position5: model.PositionFollow}
	items := make([]*api.ModuleItem, 0, len(feedRly.GetList()))
	for _, res := range feedRly.GetList() {
		if res == nil || res.GetId() <= 0 {
			continue
		}
		var cd *api.EditorRecommendCard
		switch res.GetType() {
		case hmtgrpc.ResourceType_UGC_RESOURCE:
			arc, ok := material.Arcs[res.GetId()]
			if !ok || !arc.IsNormal() {
				continue
			}
			cd = edBuilder.buildArchiveFolder(nil, &cfg.Position, ss, arc, nil, nil)
		case hmtgrpc.ResourceType_OGV_RESOURCE:
			ep, ok := material.Episodes[res.GetId()]
			if !ok {
				continue
			}
			cd = edBuilder.buildEpisode(nil, pgcPosition, ep)
		default:
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeEditor.String(),
			CardId:     strconv.FormatInt(res.GetId(), 10),
			CardDetail: &api.ModuleItem_EditorRecommendCard{EditorRecommendCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = items
	if feedRly.GetHasMore() {
		module.HasMore = true
		module.SubpageParams = bu.subpageParams(cfg, int64(feedRly.GetOffset()))
	}
	return module
}
