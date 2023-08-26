package dynHandler

import (
	"strconv"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

func (schema *CardSchema) fold(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	if dynSchemaCtx.MergedResource == nil {
		return nil
	}
	mergedResource := dynSchemaCtx.MergedResource[dynSchemaCtx.DynCtx.Dyn.DynamicID]
	if mergedResource.MergeType <= 0 && mergedResource.MergedResCnt <= 0 {
		return nil
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_fold,
		ModuleItem: &dynamicapi.Module_ModuleFold{
			ModuleFold: &dynamicapi.ModuleFold{
				FoldType: dynamicapi.FoldType_FoldTypeTopicMerged,
				Text:     "展开" + strconv.Itoa(int(mergedResource.MergedResCnt)) + "条相关动态",
				TopicMergedResource: &dynamicapi.TopicMergedResource{
					MergeType:    mergedResource.MergeType,
					MergedResCnt: mergedResource.MergedResCnt,
				},
			},
		},
	}
	dynSchemaCtx.DynCtx.DynamicItem.Modules = append(dynSchemaCtx.DynCtx.DynamicItem.Modules, module)
	return nil
}
