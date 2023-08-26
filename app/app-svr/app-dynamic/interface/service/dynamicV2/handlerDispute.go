package dynamicV2

import (
	"context"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

// 争议小黄条
func (s *Service) dispute(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.Extend == nil || dynCtx.Dyn.Extend.Dispute == nil {
		return nil
	}
	dynDisp := dynCtx.Dyn.Extend.Dispute
	if dynDisp.Content == "" {
		return nil
	}
	disp := &api.Module_ModuleDispute{
		ModuleDispute: &api.ModuleDispute{
			Title: dynDisp.Content,
			Desc:  dynDisp.Desc,
			Uri:   dynDisp.Url,
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dispute,
		ModuleItem: disp,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}
