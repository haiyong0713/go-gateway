package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type NewactStatement struct{}

func (r NewactStatement) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.NewactStatement{
		BaseCfgManager: config.NewBaseCfg(natModule),
		Sid:            natModule.Fid,
		Type:           natModule.ConfUnmarshal().StatementType,
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r NewactStatement) setBaseCfg(cfg *config.NewactStatement) {
	if cfg.Sid <= 0 {
		return
	}
	cfg.ReqID, _ = cfg.AddMaterialParam(model.MaterialActSubject, &kernel.ActSidsReq{IDs: []int64{cfg.Sid}})
}
