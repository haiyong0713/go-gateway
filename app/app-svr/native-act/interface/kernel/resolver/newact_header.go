package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type NewactHeader struct{}

func (r NewactHeader) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	if natPage.IsNewact() && ss.TabFrom == model.TabFromTopicLayer {
		return nil
	}
	natModule := module.NativeModule
	cfg := &config.NewactHeader{
		BaseCfgManager: config.NewBaseCfg(natModule),
		Sid:            natModule.Fid,
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r NewactHeader) setBaseCfg(cfg *config.NewactHeader) {
	if cfg.Sid <= 0 {
		return
	}
	cfg.ReqID, _ = cfg.AddMaterialParam(model.MaterialActSubject, &kernel.ActSidsReq{
		IDs:         []int64{cfg.Sid},
		NeedMultiML: true,
		NeedAccount: true,
	})
}
