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

type ResourceTopic struct {
}

func (bu ResourceTopic) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rtCfg, ok := cfg.(*config.ResourceTopic)
	if !ok {
		logCfgAssertionError(config.ResourceTopic{})
		return nil
	}
	briefDynsRly, ok := material.BriefDynsRlys[rtCfg.BriefDynsReqID]
	if !ok || len(briefDynsRly.Dynamics) == 0 {
		return nil
	}
	items := bu.buildModuleItems(rtCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeResource.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfResource(&rtCfg.ResourceCommon),
		ModuleItems: items,
		ModuleUkey:  cfg.ModuleBase().Ukey,
	}
	if rtCfg.DisplayViewMore && briefDynsRly.HasMore == 1 {
		module.HasMore = true
		subpageParams := subpageParamsOfResource(module.ModuleId, 0, 0, briefDynsRly.Offset)
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfResource(module.ModuleId, rtCfg.PageID, 0, briefDynsRly.Offset,
				func() *api.SubpageData {
					return buildSubpageData(rtCfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}

func (bu ResourceTopic) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu ResourceTopic) buildModuleItems(cfg *config.ResourceTopic, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	dyns := material.BriefDynsRlys[cfg.BriefDynsReqID].Dynamics
	riBuilder := &ResourceID{}
	items := make([]*api.ModuleItem, 0, len(dyns))
	for _, dyn := range dyns {
		if dyn == nil || dyn.Rid == 0 {
			continue
		}
		var cd *api.ResourceCard
		switch dyn.Type {
		case model.DynTypeVideo:
			arc, ok := material.Arcs[dyn.Rid]
			if !ok || !arc.IsNormal() {
				continue
			}
			cd = riBuilder.buildArchiveFolder(arc, nil, ss, cfg.DisplayUGCBadge)
		case model.DynTypeArticle:
			art, ok := material.Articles[dyn.Rid]
			if !ok || !art.IsNormal() {
				continue
			}
			cd = riBuilder.buildArticle(art, cfg.DisplayArticleBadge)
		default:
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeResource.String(),
			CardId:     strconv.FormatInt(dyn.Rid, 10),
			CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	return unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
}
