package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Statement struct{}

func (bu Statement) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	stCfg, ok := cfg.(*config.Statement)
	if !ok {
		logCfgAssertionError(config.Statement{})
		return nil
	}
	item := &api.ModuleItem{
		CardType: model.CardTypeStatement.String(),
		CardDetail: &api.ModuleItem_StatementCard{
			StatementCard: &api.StatementCard{Content: stCfg.Content},
		},
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeStatement.String(),
		ModuleId:      stCfg.ModuleBase().ModuleID,
		ModuleColor:   &api.Color{BgColor: stCfg.BgColor, FontColor: stCfg.FontColor},
		ModuleSetting: &api.Setting{DisplayUnfoldButton: stCfg.DisplayUnfoldButton},
		ModuleItems:   []*api.ModuleItem{item},
		ModuleUkey:    stCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Statement) After(data *AfterContextData, current *api.Module) bool {
	return true
}
