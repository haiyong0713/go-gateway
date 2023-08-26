package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Relativeact struct{}

func (r Relativeact) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Relativeact{
		BaseCfgManager:     config.NewBaseCfg(natModule),
		ImageTitle:         natModule.Meta,
		BgColor:            natModule.BgColor,
		CardTitleFontColor: natModule.TitleColor,
	}
	if module.Act != nil {
		cfg.Acts = module.Act.List
	}
	return cfg
}
