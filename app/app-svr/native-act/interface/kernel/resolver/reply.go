package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Reply struct{}

func (r Reply) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	return &config.Reply{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ReplyID:        natModule.Fid,
		Type:           natModule.AvSort,
	}
}
