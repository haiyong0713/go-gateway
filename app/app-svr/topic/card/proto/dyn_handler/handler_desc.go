package dynHandler

import (
	"fmt"
	"strconv"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	topiccardschema "go-gateway/app/app-svr/topic/card/schema"
)

// 描述信息
func (schema *CardSchema) description(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	descArr := topiccardschema.DescProc(dynCtx, dynCtx.Interim.Desc, general)
	if len(descArr) == 0 {
		return nil
	}
	moduleDesc := &dynamicapi.Module_ModuleDesc{
		ModuleDesc: &dynamicapi.ModuleDesc{
			Desc:    descArr,
			Text:    dynCtx.Interim.Desc,
			JumpUri: dynCtx.Interim.PromoURI, // 帮推
		},
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_desc,
		ModuleItem: moduleDesc,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	// 拓展字段内容
	if dynCtx.Dyn.IsForward() {
		var descTmp []*dynamicapi.Description
		descTmp = append(descTmp, &dynamicapi.Description{
			Type: dynamicapi.DescType_desc_type_text,
			Text: "//",
		})
		descTmp = append(descTmp, &dynamicapi.Description{
			Text: fmt.Sprintf("@%s", dynCtx.DynamicItem.Extend.OrigName),
			Type: dynamicapi.DescType_desc_type_aite,
			Uri:  topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, strconv.FormatInt(dynCtx.DynamicItem.Extend.Uid, 10), nil),
			Rid:  strconv.FormatInt(dynCtx.DynamicItem.Extend.Uid, 10),
		})
		descTmp = append(descTmp, &dynamicapi.Description{
			Type: dynamicapi.DescType_desc_type_text,
			Text: ":",
		})
		dynCtx.DynamicItem.Extend.Desc = descTmp
		dynCtx.DynamicItem.Extend.Desc = append(dynCtx.DynamicItem.Extend.Desc, descArr...)
	}
	return nil
}
