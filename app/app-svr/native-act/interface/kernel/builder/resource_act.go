package builder

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type ResourceAct struct{}

func (bu ResourceAct) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	raCfg, ok := cfg.(*config.ResourceAct)
	if !ok {
		logCfgAssertionError(config.ResourceAct{})
		return nil
	}
	likesRly, ok := material.ActLikesRlys[raCfg.ActLikesReqID]
	if !ok || len(likesRly.List) == 0 {
		return nil
	}
	items := bu.buildModuleItems(raCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeResource.String(),
		ModuleId:    raCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfResource(&raCfg.ResourceCommon),
		ModuleItems: items,
		ModuleUkey:  raCfg.ModuleBase().Ukey,
	}
	if raCfg.DisplayViewMore && likesRly.HasMore == 1 {
		module.HasMore = true
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, raCfg.PageID, likesRly.Offset, "",
				func() *api.SubpageData {
					return buildSubpageData(raCfg.SubpageTitle, raCfg.SortList, func(sort int64) string {
						if sort == SubpageCurrSortKey {
							sort = raCfg.SortType
						}
						// 非当前排序将从0开始
						var offset int64
						if sort == raCfg.SortType {
							offset = likesRly.Offset
						}
						return subpageParamsOfResource(module.ModuleId, sort, offset, "")
					})
				},
			))
		} else {
			module.SubpageParams = subpageParamsOfResource(module.ModuleId, raCfg.SortType, likesRly.Offset, "")
		}
	}
	return module
}

func (bu ResourceAct) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu ResourceAct) buildModuleItems(cfg *config.ResourceAct, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	likesRly, ok := material.ActLikesRlys[cfg.ActLikesReqID]
	if !ok || likesRly.Subject == nil || len(likesRly.List) == 0 {
		return nil
	}
	riBuilder := ResourceID{}
	items := make([]*api.ModuleItem, 0, len(likesRly.List))
	for _, v := range likesRly.List {
		if v == nil || v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		var cd *api.ResourceCard
		switch likesRly.Subject.Type {
		case model.ActSubTypeVideoLike, model.ActSubTypeVideo2, model.ActSubTypePhoneVideo:
			arc, ok := material.Arcs[v.Item.Wid]
			if !ok || !arc.IsNormal() {
				continue
			}
			cd = riBuilder.buildArchiveFolder(arc, nil, ss, cfg.DisplayUGCBadge)
		case model.ActSubTypeArticle:
			art, ok := material.Articles[v.Item.Wid]
			if !ok || !art.IsNormal() {
				continue
			}
			cd = riBuilder.buildArticle(art, cfg.DisplayArticleBadge)
		default:
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(v.Item.Wid, 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	return unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
}
