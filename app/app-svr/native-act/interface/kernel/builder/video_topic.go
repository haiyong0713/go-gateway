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

type VideoTopic struct{}

func (bu VideoTopic) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	vtCfg, ok := cfg.(*config.VideoTopic)
	if !ok {
		logCfgAssertionError(config.VideoTopic{})
		return nil
	}
	briefDynsRly, ok := material.BriefDynsRlys[vtCfg.BriefDynsReqID]
	if !ok || len(briefDynsRly.Dynamics) == 0 {
		return nil
	}
	items := bu.buildModuleItems(vtCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeVideo.String(),
		ModuleId:      vtCfg.ModuleBase().ModuleID,
		ModuleColor:   buildModuleColorOfVideo(&vtCfg.VideoCommon),
		ModuleSetting: &api.Setting{DisplayTitle: !vtCfg.HideTitle, AutoPlay: vtCfg.AutoPlay},
		ModuleItems:   items,
		ModuleUkey:    vtCfg.ModuleBase().Ukey,
	}
	if vtCfg.DisplayViewMore && briefDynsRly.HasMore == 1 {
		module.HasMore = true
		subpageParams := subpageParamsOfVideo(module.ModuleId, 0, 0, briefDynsRly.Offset)
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfVideo(module.ModuleId, vtCfg.PageID, 0, briefDynsRly.Offset,
				func() *api.SubpageData {
					return buildSubpageData(vtCfg.SubpageTitle, nil, func(sort int64) string { return subpageParams })
				},
			))
		} else {
			module.SubpageParams = subpageParams
		}
	}
	return module
}

func (bu VideoTopic) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu VideoTopic) buildModuleItems(cfg *config.VideoTopic, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	dynRly, ok := material.BriefDynsRlys[cfg.BriefDynsReqID]
	if !ok || len(dynRly.Dynamics) == 0 {
		return nil
	}
	viBuilder := VideoID{}
	items := make([]*api.ModuleItem, 0, len(dynRly.Dynamics))
	for _, dyn := range dynRly.Dynamics {
		if dyn == nil || dyn.Rid == 0 || dyn.Type != model.DynTypeVideo {
			continue
		}
		arcPlayer, ok := material.ArcsPlayer[dyn.Rid]
		if !ok || arcPlayer.GetArc() == nil || !arcPlayer.GetArc().IsNormal() {
			continue
		}
		cd := viBuilder.buildArchive(arcPlayer, ss)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeVideo.String(),
			CardId:     strconv.FormatInt(arcPlayer.GetArc().GetAid(), 10),
			CardDetail: &api.ModuleItem_VideoCard{VideoCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	items = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	return items
}
