package cardbuilder

import (
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
)

func handleModuleDispute(dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleDispute {
	if dynCtx.Dyn.Extend == nil || dynCtx.Dyn.Extend.Dispute == nil {
		return nil
	}
	dynDisp := dynCtx.Dyn.Extend.Dispute
	if dynDisp.Content == "" {
		return nil
	}
	return &jsonwebcard.ModuleDispute{
		Title:   dynDisp.Content,
		Desc:    dynDisp.Desc,
		JumpUrl: dynDisp.Url,
	}
}
