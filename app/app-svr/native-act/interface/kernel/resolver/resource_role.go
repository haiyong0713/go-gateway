package resolver

import (
	"context"

	chargrpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type ResourceRole struct{}

func (r ResourceRole) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.ResourceRole{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ResourceCommon: buildResourceCommon(natModule, ss),
		ShowNum:        natModule.Num,
	}
	r.setBaseCfg(cfg, natModule)
	return cfg
}

func (r ResourceRole) setBaseCfg(cfg *config.ResourceRole, module *natpagegrpc.NativeModule) {
	charID := int32(module.Length)
	seasonID := int32(module.Width)
	if charID == 0 || seasonID == 0 {
		return
	}
	cfg.RelInfosReqID, _ = cfg.AddMaterialParam(model.MaterialRelInfosRly, &kernel.RelInfosReq{
		Req: &chargrpc.CharacterIdsOidsReq{
			CharacterIdOpusIds: map[int32]*chargrpc.OpusIdsReq{charID: {Ids: []int32{seasonID}}},
			Otype:              model.CharOtypeSeason,
		},
		NeedMultiML: true,
		ShowNum:     cfg.ShowNum,
	})
}
