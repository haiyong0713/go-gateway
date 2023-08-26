package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type BottomButton struct{}

func (bu BottomButton) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	if ckCfg, ok := cfg.(*config.Click); ok && ckCfg.BgImage == nil {
		return nil
	}
	module := Click{}.Build(c, ss, dep, cfg, material)
	module.ModuleType = model.ModuleTypeBottomButton.String()
	return module
}

func (bu BottomButton) After(data *AfterContextData, current *api.Module) bool {
	return true
}
