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

type VideoAct struct{}

func (bu VideoAct) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	vaCfg, ok := cfg.(*config.VideoAct)
	if !ok {
		logCfgAssertionError(config.VideoAct{})
		return nil
	}
	likesRly, ok := material.ActLikesRlys[vaCfg.ActLikesReqID]
	if !ok || len(likesRly.List) == 0 {
		return nil
	}
	items := bu.buildModuleItems(vaCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeVideo.String(),
		ModuleId:      vaCfg.ModuleBase().ModuleID,
		ModuleColor:   buildModuleColorOfVideo(&vaCfg.VideoCommon),
		ModuleSetting: &api.Setting{DisplayTitle: !vaCfg.HideTitle, AutoPlay: vaCfg.AutoPlay},
		ModuleItems:   items,
		ModuleUkey:    vaCfg.ModuleBase().Ukey,
	}
	if vaCfg.DisplayViewMore && likesRly.HasMore == 1 {
		module.HasMore = true
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildMoreCardOfVideo(module.ModuleId, vaCfg.PageID, likesRly.Offset, "",
				func() *api.SubpageData {
					return buildSubpageData(vaCfg.SubpageTitle, vaCfg.SortList, func(sort int64) string {
						if sort == SubpageCurrSortKey {
							sort = vaCfg.SortType
						}
						var offset int64
						if sort == vaCfg.SortType {
							offset = likesRly.Offset
						}
						return subpageParamsOfVideo(module.ModuleId, sort, offset, "")
					})
				},
			))
		} else {
			module.SubpageParams = subpageParamsOfVideo(module.ModuleId, 0, likesRly.Offset, "")
		}
	}
	return module
}

func (bu VideoAct) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu VideoAct) buildModuleItems(cfg *config.VideoAct, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	likesRly, ok := material.ActLikesRlys[cfg.ActLikesReqID]
	if !ok || likesRly.Subject == nil || len(likesRly.List) == 0 {
		return nil
	}
	viBuilder := VideoID{}
	items := make([]*api.ModuleItem, 0, len(likesRly.List))
	for _, v := range likesRly.List {
		if v == nil || v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		arcPlayer, ok := material.ArcsPlayer[v.Item.Wid]
		if !ok || arcPlayer.GetArc() == nil || !arcPlayer.GetArc().IsNormal() {
			continue
		}
		cd := viBuilder.buildArchive(arcPlayer, ss)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeVideo.String(),
			CardId:     strconv.FormatInt(v.Item.Wid, 10),
			CardDetail: &api.ModuleItem_VideoCard{VideoCard: cd},
		})
	}
	if len(items) == 0 {
		return nil
	}
	items = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	return items
}
