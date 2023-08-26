package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type OgvOrigin struct{}

func (bu OgvOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	ooCfg, ok := cfg.(*config.OgvOrigin)
	if !ok {
		logCfgAssertionError(config.OgvOrigin{})
		return nil
	}
	hasMore, offset, items := bu.buildModuleItems(ooCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeOgv.String(),
		ModuleId:    ooCfg.ModuleBase().ModuleID,
		ModuleColor: buildModuleColorOfOgv(ooCfg.Color, ooCfg.IsThreeCard),
		ModuleItems: items,
		ModuleUkey:  ooCfg.ModuleBase().Ukey,
		HasMore:     hasMore,
	}
	if hasMore && ooCfg.DisplayMore {
		if model.IsFromIndex(ss.ReqFrom) {
			module.ModuleItems = append(module.ModuleItems, buildOgvMore(&ooCfg.OgvCommon, offset, ooCfg.ModuleBase().ModuleID))
		} else {
			module.SubpageParams = ogvSupernatantParams(offset, -1, ooCfg.ModuleBase().ModuleID)
		}
	}
	return module
}

func (bu OgvOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu OgvOrigin) buildModuleItems(cfg *config.OgvOrigin, material *kernel.Material, ss *kernel.Session) (bool, int64, []*api.ModuleItem) {
	seasonsRly, ok := material.SeasonByPlayIdRlys[cfg.SeasonByPlayIdReq]
	if !ok {
		return false, 0, nil
	}
	ogvBu := Ogv{}
	items := make([]*api.ModuleItem, 0, len(seasonsRly.SeasonInfos))
	for _, sc := range seasonsRly.SeasonInfos {
		if sc == nil {
			continue
		}
		if cfg.IsThreeCard {
			items = append(items, ogvBu.buildOgvThree(&cfg.OgvCommon, sc, ""))
		} else {
			items = append(items, ogvBu.buildOgvOne(&cfg.OgvCommon, sc, ""))
		}
	}
	if len(items) == 0 {
		return false, 0, nil
	}
	items = unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
	return seasonsRly.HasNext, int64(seasonsRly.NexOffset), items
}
