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

type NewactAward struct{}

func (bu NewactAward) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	naCfg, ok := cfg.(*config.NewactAward)
	if !ok {
		logCfgAssertionError(config.NewactAward{})
		return nil
	}
	items := bu.buildModuleItems(naCfg, material)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeNewactAward.String(),
		ModuleId:    naCfg.ModuleBase().ModuleID,
		ModuleItems: items,
		ModuleUkey:  naCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu NewactAward) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu NewactAward) buildModuleItems(cfg *config.NewactAward, material *kernel.Material) []*api.ModuleItem {
	st, ok := material.ActSubjects[cfg.ReqID][cfg.Sid]
	if !ok {
		return nil
	}
	cd := &api.NewactAward{
		Title: "活动奖励",
		Items: make([]*api.NewactAwardItem, 0, len(st.ActivityAward)),
	}
	for _, ad := range st.ActivityAward {
		if ad == nil {
			continue
		}
		cd.Items = append(cd.Items, &api.NewactAwardItem{Title: ad.Title, Content: ad.Desc})
	}
	if len(cd.Items) == 0 {
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeNewactAward.String(),
		CardId:     strconv.FormatInt(cfg.Sid, 10),
		CardDetail: &api.ModuleItem_NewactAwardCard{NewactAwardCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}
