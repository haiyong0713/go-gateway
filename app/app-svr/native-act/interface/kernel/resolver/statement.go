package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Statement struct{}

func (r Statement) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Statement{
		BaseCfgManager:      config.NewBaseCfg(natModule),
		Content:             natModule.Remark,
		FontColor:           natModule.TitleColor,
		BgColor:             natModule.BgColor,
		DisplayUnfoldButton: natModule.IsAttrStatementDisplayButton() == natpagegrpc.AttrModuleYes,
	}
	return cfg
}
