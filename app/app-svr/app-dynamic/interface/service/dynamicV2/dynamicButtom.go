package dynamicV2

import (
	"context"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

func (s *Service) buttom(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_bottom,
		ModuleItem: &api.Module_ModuleButtom{
			ModuleButtom: &api.ModuleButtom{
				ModuleStat: s.statInfo(c, dynCtx, general),
			},
		},
	}
	if s.dynDetailBottomBar(c, general) {
		mdlBtm := module.ModuleItem.(*api.Module_ModuleButtom).ModuleButtom
		mdlBtm.CommentBox, mdlBtm.CommentBoxMsg = true, "点我发评论"
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}
