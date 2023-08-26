package dynHandler

import (
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

// 争议小黄条
func (schema *CardSchema) dispute(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
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
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dispute,
		ModuleItem: &dynamicapi.Module_ModuleDispute{
			ModuleDispute: &dynamicapi.ModuleDispute{
				Title: dynDisp.Content,
				Desc:  dynDisp.Desc,
				Uri:   dynDisp.Url,
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}
