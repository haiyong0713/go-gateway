package builder

import (
	"context"
	"fmt"
	"strconv"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type ResourceOrigin struct{}

func (bu ResourceOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	roCfg, ok := cfg.(*config.ResourceOrigin)
	if !ok {
		logCfgAssertionError(config.ResourceOrigin{})
		return nil
	}
	switch roCfg.OriginType {
	case model.RDBOgv:
		return bu.buildModuleOfOgvWid(roCfg, material, ss)
	case model.RDBLive:
		return bu.buildModuleOfLive(roCfg, material, ss)
	case model.RDBBizCommodity:
		return bu.buildModuleOfBizCommodity(roCfg, material, ss)
	case model.RDBBizIds:
		return bu.buildModuleOfBizIds(roCfg, material, ss)
	}
	return nil
}

func (bu ResourceOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu ResourceOrigin) buildModuleBase(cfg *config.ResourceOrigin) *api.Module {
	module := &api.Module{
		ModuleType:  model.ModuleTypeResource.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfResource(&cfg.ResourceCommon),
		ModuleUkey:  cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu ResourceOrigin) buildModuleOfOgvWid(cfg *config.ResourceOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	widRly, ok := material.QueryWidRlys[cfg.Wid]
	if !ok || len(widRly.GetItems()) == 0 {
		return nil
	}
	var hasMore bool
	widItems := widRly.GetItems()
	// 首页返回 module.Num 条数据，二级页返回剩余的
	if model.IsFromIndex(ss.ReqFrom) {
		if int64(len(widItems)) > cfg.ShowNum {
			hasMore = true
			widItems = widItems[:cfg.ShowNum]
		}
	} else {
		if int64(len(widItems)) <= ss.Offset {
			return nil
		}
		widItems = widItems[ss.Offset:]
	}
	offset := ss.Offset + int64(len(widItems))
	items := make([]*api.ModuleItem, 0, len(widItems))
	for _, widItem := range widItems {
		if widItem == nil {
			continue
		}
		cd := bu.buildOgvWid(widItem)
		if cd == nil {
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(int64(widItem.GetId()), 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	if cfg.DisplayViewMore && hasMore {
		module.HasMore = true
		subpageParams := subpageParamsOfResource(module.ModuleId, 0, offset, "")
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, cfg.PageID, offset, "",
				func() *api.SubpageData {
					return buildSubpageData(cfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}

func (bu ResourceOrigin) buildOgvWid(widItem *pgcappgrpc.QueryWidItem) *api.ResourceCard {
	cd := &api.ResourceCard{
		Title:         widItem.GetTitle(),
		CoverImageUri: widItem.GetCover(),
		Uri:           widItem.GetLink(),
	}
	switch widItem.GetType() {
	case model.WidItemTypeUgc:
		bu.setCardOfWidUgc(cd, widItem)
	case model.WidItemTypeSeason:
		bu.setCardOfWidSeason(cd, widItem)
	case model.WidItemTypeOgv:
		bu.setCardOfWidOgv(cd, widItem)
	case model.WidItemTypeWeb:
		bu.setCardOfWidWeb(cd)
	case model.WidItemTypeOgvFilm:
		bu.setCardOfWidOgvFilm(cd)
	default:
		return nil
	}
	return cd
}

func (bu ResourceOrigin) setCardOfWidUgc(cd *api.ResourceCard, widItem *pgcappgrpc.QueryWidItem) {
	cd.CoverLeftIcon1 = int64(appcardmdl.IconPlay)
	cd.CoverLeftText1 = appcardmdl.Stat64String(widItem.GetPlay(), "")
	cd.CoverLeftIcon2 = int64(appcardmdl.IconDanmaku)
	cd.CoverLeftText2 = appcardmdl.Stat64String(widItem.GetDm(), "")
	cd.CoverRightText = appcardmdl.DurationString(widItem.GetPlayLen())
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeUGC}
	cd.ResourceType = model.ResourceTypeUGC
}

func (bu ResourceOrigin) setCardOfWidSeason(cd *api.ResourceCard, widItem *pgcappgrpc.QueryWidItem) {
	cd.CoverLeftIcon1 = int64(appcardmdl.IconPlay)
	cd.CoverLeftText1 = appcardmdl.Stat64String(widItem.GetPlay(), "")
	cd.CoverLeftIcon2 = int64(appcardmdl.IconFavorite)
	cd.CoverLeftText2 = appcardmdl.Stat64String(widItem.GetFollow(), "")
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeSeason}
	cd.ResourceType = model.ResourceTypeSeason
}

func (bu ResourceOrigin) setCardOfWidOgv(cd *api.ResourceCard, widItem *pgcappgrpc.QueryWidItem) {
	cd.CoverLeftIcon1 = int64(appcardmdl.IconPlay)
	cd.CoverLeftText1 = appcardmdl.Stat64String(widItem.GetPlay(), "")
	cd.CoverLeftIcon2 = int64(appcardmdl.IconDanmaku)
	cd.CoverLeftText2 = appcardmdl.Stat64String(widItem.GetDm(), "")
	cd.CoverRightText = appcardmdl.DurationString(widItem.GetPlayLen())
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypePGC}
	cd.ResourceType = model.ResourceTypeOGV
}

func (bu ResourceOrigin) setCardOfWidWeb(cd *api.ResourceCard) {
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeWeb}
	cd.ResourceType = model.ResourceTypeWeb
}

func (bu ResourceOrigin) setCardOfWidOgvFilm(cd *api.ResourceCard) {
	cd.ReportDic = &api.ReportDic{BizType: model.ReportBizTypeOgvFilm}
	cd.ResourceType = model.ResourceTypeOgvFilm
}

func (bu ResourceOrigin) buildModuleOfLive(cfg *config.ResourceOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	roomsRly, ok := material.RoomsByActIdRlys[cfg.RoomsByActIdReqID]
	if !ok {
		return nil
	}
	items := make([]*api.ModuleItem, 0, len(roomsRly.GetList()))
	for _, room := range roomsRly.GetList() {
		if room == nil {
			continue
		}
		cd := bu.buildLive(room)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(room.GetRoomId(), 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	if cfg.DisplayViewMore && roomsRly.HasMore {
		module.HasMore = true
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, cfg.PageID, roomsRly.Offset, "",
				func() *api.SubpageData {
					return buildSubpageData(cfg.SubpageTitle, cfg.TabList, func(sort int64) string {
						if sort == SubpageCurrSortKey {
							sort = cfg.TabID
						}
						var offset int64
						if sort == cfg.TabID {
							offset = roomsRly.Offset
						}
						return subpageParamsOfResource(module.ModuleId, sort, offset, "")
					})
				},
			))
		} else {
			module.SubpageParams = subpageParamsOfResource(module.ModuleId, 0, roomsRly.Offset, "")
		}
	}
	return module
}

func (bu ResourceOrigin) buildLive(room *liveplaygrpc.RoomList) *api.ResourceCard {
	cd := &api.ResourceCard{
		Title:          room.GetTitle(),
		CoverImageUri:  room.GetIcon(),
		Uri:            fmt.Sprintf("https://live.bilibili.com/%d", room.GetRoomId()),
		CoverRightText: room.GetUserName(),
		CoverLeftText1: room.GetOnline(),
		CoverLeftIcon1: int64(appcardmdl.IconOnline),
		ReportDic:      &api.ReportDic{BizType: model.ReportBizTypeLive},
		ResourceType:   model.ResourceTypeLive,
	}
	if room.GetPendant() != "" {
		cd.Badge = &api.Badge{Text: room.GetPendant()}
	}
	return cd
}

func (bu ResourceOrigin) buildModuleOfBizCommodity(cfg *config.ResourceOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	detailRly, ok := material.ProductDetailRlys[cfg.ProductDetailReqID]
	if !ok || len(detailRly.ItemList) == 0 {
		return nil
	}
	items := make([]*api.ModuleItem, 0, len(detailRly.ItemList))
	for _, v := range detailRly.ItemList {
		if v == nil {
			continue
		}
		cd := bu.buildBizCommodity(v)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(v.ItemId, 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	if cfg.DisplayViewMore && detailRly.HasMore == 1 {
		module.HasMore = true
		subpageParams := subpageParamsOfResource(module.ModuleId, 0, detailRly.Offset, "")
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, cfg.PageID, detailRly.Offset, "",
				func() *api.SubpageData {
					return buildSubpageData(cfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}

func (bu ResourceOrigin) buildBizCommodity(detail *model.ProductDetailItem) *api.ResourceCard {
	return &api.ResourceCard{
		Title:         detail.Title,
		CoverImageUri: detail.ImageUrl,
		Uri:           detail.LinkUrl,
		ReportDic:     &api.ReportDic{BizType: model.ReportBizTypeBizCommodity},
		ResourceType:  model.ResourceTypeCommodity,
	}
}

func (bu ResourceOrigin) buildModuleOfBizIds(cfg *config.ResourceOrigin, material *kernel.Material, ss *kernel.Session) *api.Module {
	detailRly, ok := material.SourceDetailRlys[cfg.SourceDetailReqID]
	if !ok || len(detailRly.ItemList) == 0 {
		return nil
	}
	items := make([]*api.ModuleItem, 0, len(detailRly.ItemList))
	riBuilder := ResourceID{}
	for _, v := range detailRly.ItemList {
		if v == nil {
			continue
		}
		var cd *api.ResourceCard
		switch v.Type {
		case natpagegrpc.MixAvidType, natpagegrpc.MixFolder:
			arc, ok := material.Arcs[v.ItemId]
			if !ok || !arc.IsNormal() {
				continue
			}
			var folder *favmdl.Folder
			if v.Type == natpagegrpc.MixFolder {
				if folder, ok = material.Folders[favmdl.TypeVideo][v.Fid]; !ok {
					continue
				}
			}
			cd = riBuilder.buildArchiveFolder(arc, folder, ss, cfg.DisplayUGCBadge)
		case natpagegrpc.MixEpidType:
			ep, ok := material.Episodes[v.ItemId]
			if !ok {
				continue
			}
			cd = riBuilder.buildEpisode(ep, cfg.DisplayPGCBadge)
		case natpagegrpc.MixCvidType:
			art, ok := material.Articles[v.ItemId]
			if !ok || !art.IsNormal() {
				continue
			}
			cd = riBuilder.buildArticle(art, cfg.DisplayArticleBadge)
		default:
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(v.ItemId, 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg)
	module.ModuleItems = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	if cfg.DisplayViewMore && detailRly.HasMore == 1 {
		module.HasMore = true
		subpageParams := subpageParamsOfResource(module.ModuleId, 0, detailRly.Offset, "")
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, cfg.PageID, detailRly.Offset, "",
				func() *api.SubpageData {
					return buildSubpageData(cfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}
