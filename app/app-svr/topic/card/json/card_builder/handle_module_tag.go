package cardbuilder

import (
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"

	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func handleModuleTag(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleTag {
	if metaCtx.Config.ItemFromControl == nil {
		return nil
	}
	if val, ok := metaCtx.Config.ItemFromControl[dynCtx.Dyn.DynamicID]; ok && val == topicsvc.ItemFrom_TopShow.String() {
		return &jsonwebcard.ModuleTag{
			Text: "置顶",
		}
	}
	return nil
}
