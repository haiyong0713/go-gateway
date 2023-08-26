package dynHandler

import (
	"go-common/library/log"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	topiccardschema "go-gateway/app/app-svr/topic/card/schema"
)

type Handler func(*topiccardmodel.DynSchemaCtx, *topiccardmodel.GeneralParam) error

type CardSchema struct{}

func (schema *CardSchema) ProcListReply(dynSchemaCtx *topiccardmodel.DynSchemaCtx, dynamics []*dynmdlV2.Dynamic, general *topiccardmodel.GeneralParam, from string) *topiccardmodel.DynRawList {
	dynCtx, rawList := dynSchemaCtx.DynCtx, &topiccardmodel.DynRawList{}
	for _, dyn := range dynamics {
		dynCtx.Dyn = dyn                                                           // 原始数据
		dynCtx.DynamicItem = &dynamicapi.DynamicItem{Extend: &dynamicapi.Extend{}} // 聚合结果
		dynCtx.Interim = &dynmdlV2.Interim{}                                       // 临时逻辑
		var (
			handlerList []Handler
			ok          bool
		)
		switch from {
		case _handleTypeForward:
			handlerList, ok = schema.getHandlerListForward(dynSchemaCtx)
		default:
			handlerList, ok = schema.getHandlerList(dynSchemaCtx)
		}
		if !ok {
			log.Warn("dynamic mid(%v) DynamicID(%d) getHandlerList !ok", general.Mid, dynCtx.Dyn.DynamicID)
			continue
		}
		// 执行拼接func
		if err := schema.conveyer(dynSchemaCtx, general, handlerList...); err != nil {
			log.Warn("dynamic mid(%v) conveyer, err %v", general.Mid, err)
			continue
		}
		if dynCtx.Interim.IsPassCard {
			log.Warn("dynamic mid(%v) IsPassCard dynid %v", general.Mid, dyn.DynamicID)
			continue
		}
		// 收割上下文中组装完成的items
		rawList.List = append(rawList.List, &topiccardmodel.DynRawItem{
			Item: dynCtx.DynamicItem,
		})
	}
	return rawList
}

func (schema *CardSchema) conveyer(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, f ...Handler) error {
	for _, v := range f {
		err := v(dynSchemaCtx, general)
		if err != nil {
			log.Error("Conveyer failed. dynamic: %v, error: %+v", dynSchemaCtx.DynCtx.Dyn.DynamicID, err)
			return err
		}
	}
	return nil
}

func (schema *CardSchema) getHandlerList(dynSchemaCtx *topiccardmodel.DynSchemaCtx) ([]Handler, bool) {
	ret := schema.handlerPreProcess() // handler预处理
	switch {
	case dynSchemaCtx.CanBeForward(): // 转发卡
		ret = append(ret, schema.author, schema.dispute, schema.description, schema.dynCardForward, schema.additional, schema.interaction, schema.stat, schema.fold)
	case dynSchemaCtx.CanBeAv(): // 视频卡
		ret = append(ret, schema.author, schema.dispute, schema.description, schema.dynCardAv, schema.additional, schema.interaction, schema.topDetailsExt, schema.stat, schema.fold)
	case dynSchemaCtx.CanBeDraw(): // 图文卡
		ret = append(ret, schema.author, schema.dispute, schema.description, schema.dynCardDraw, schema.additional, schema.interaction, schema.topDetailsExt, schema.stat, schema.fold)
	case dynSchemaCtx.CanBeWord(): // 纯文字卡
		ret = append(ret, schema.author, schema.dispute, schema.description, schema.additional, schema.interaction, schema.topDetailsExt, schema.stat, schema.fold)
	case dynSchemaCtx.CanBeArticle(): // 专栏卡
		ret = append(ret, schema.author, schema.dispute, schema.description, schema.dynCardArticle, schema.additional, schema.interaction, schema.topDetailsExt, schema.stat, schema.fold)
	case dynSchemaCtx.CanBePGC(): // pgc卡
		ret = append(ret, schema.authorPGC, schema.dispute, schema.dynCardPGC, schema.additional, schema.topDetailsExt, schema.stat, schema.fold)
	case dynSchemaCtx.CanBeCommon(): // 通用模板
		ret = append(ret, schema.author, schema.dispute, schema.description, schema.dynCardCommon, schema.additional, schema.topDetailsExt, schema.stat, schema.fold)
	default:
		return nil, false
	}
	return ret, true
}

func (schema *CardSchema) getHandlerListForward(dynSchemaCtx *topiccardmodel.DynSchemaCtx) ([]Handler, bool) {
	ret := schema.handlerPreProcess() // handler预处理
	switch {
	case dynSchemaCtx.CanBeForward():
		ret = append(ret, schema.authorShell, schema.dispute, schema.description, schema.dynCardForward, schema.additional, schema.statShell, schema.fold)
	case dynSchemaCtx.CanBeAv():
		ret = append(ret, schema.authorShell, schema.dispute, schema.description, schema.dynCardAv, schema.additional, schema.statShell, schema.fold)
	case dynSchemaCtx.CanBeDraw():
		ret = append(ret, schema.authorShell, schema.dispute, schema.description, schema.dynCardDraw, schema.additional, schema.statShell, schema.fold)
	case dynSchemaCtx.CanBeWord():
		ret = append(ret, schema.authorShell, schema.dispute, schema.description, schema.additional, schema.statShell, schema.fold)
	case dynSchemaCtx.CanBeArticle():
		ret = append(ret, schema.authorShell, schema.dispute, schema.description, schema.dynCardArticle, schema.additional, schema.statShell, schema.fold)
	case dynSchemaCtx.CanBePGC():
		ret = append(ret, schema.authorShellPGC, schema.dispute, schema.description, schema.dynCardPGC, schema.additional, schema.statShell, schema.fold)
	case dynSchemaCtx.CanBeCommon():
		ret = append(ret, schema.authorShell, schema.dispute, schema.description, schema.dynCardCommon, schema.additional, schema.statShell, schema.fold)
	default:
		return nil, false
	}
	return ret, true
}

func (schema *CardSchema) handlerPreProcess() []Handler {
	var ret []Handler
	ret = append(ret, topiccardschema.HandleDynamicCardBase)
	return ret
}
