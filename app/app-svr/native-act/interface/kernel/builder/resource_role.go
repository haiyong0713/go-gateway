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

type ResourceRole struct{}

func (bu ResourceRole) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rrCfg, ok := cfg.(*config.ResourceRole)
	if !ok {
		logCfgAssertionError(config.ResourceRole{})
		return nil
	}
	items := bu.buildModuleItems(rrCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeResource.String(),
		ModuleId:    rrCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfResource(&rrCfg.ResourceCommon),
		ModuleItems: items,
		ModuleUkey:  rrCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu ResourceRole) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu ResourceRole) buildModuleItems(cfg *config.ResourceRole, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	relInfosRly, ok := material.RelInfosRlys[cfg.RelInfosReqID]
	if !ok {
		return nil
	}
	riBuilder := &ResourceID{}
	items := make([]*api.ModuleItem, 0, cfg.ShowNum)
	for _, info := range relInfosRly.GetInfos() {
		for _, epList := range info.GetCharacterEp() {
			for _, charEp := range epList.GetCharacterEp() {
				if charEp == nil || charEp.GetEpId() == 0 {
					continue
				}
				ep, ok := material.Episodes[int64(charEp.GetEpId())]
				if !ok {
					continue
				}
				cd := riBuilder.buildEpisode(ep, cfg.DisplayPGCBadge)
				items = append(items, &api.ModuleItem{
					CardType:   model.CardTypeResource.String(),
					CardId:     strconv.FormatInt(ep.EpID, 10),
					CardDetail: &api.ModuleItem_ResourceCard{ResourceCard: cd},
				})
			}
		}
	}
	if int64(len(items)) > cfg.ShowNum {
		items = items[:cfg.ShowNum]
	}
	return unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
}
