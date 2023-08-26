package dynHandler

import (
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

func (schema *CardSchema) topDetailsExt(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_topic_details_ext,
		ModuleItem: &dynamicapi.Module_ModuleTopicDetailsExt{
			ModuleTopicDetailsExt: &dynamicapi.ModuleTopicDetailsExt{
				CommentGuide: "有想法的话就来说点什么吧",
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}
