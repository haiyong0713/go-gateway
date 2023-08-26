package dynHandler

import (
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

// 暂时下线，代码保留样式
func (schema *CardSchema) topTag(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_top_tag,
		ModuleItem: &dynamicapi.Module_ModuleTopTag{
			ModuleTopTag: &dynamicapi.ModuleTopTag{TagName: "置顶"},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}
