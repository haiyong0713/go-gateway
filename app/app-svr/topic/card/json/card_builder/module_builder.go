package cardbuilder

import (
	"go-common/library/log"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	topiccardschema "go-gateway/app/app-svr/topic/card/schema"

	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/pkg/errors"
)

type ModuleBuilder interface {
	HandleModulePre(jsonwebcard.MetaContext, *dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleAuthor(jsonwebcard.MetaContext, *dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleDynamic(jsonwebcard.MetaContext, jsonwebcard.CardType, *dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleDispute(*dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleStat(jsonwebcard.CardType, *dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleInteraction(jsonwebcard.MetaContext, *dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleShareInfo(*dynmdlV2.DynamicContext) ModuleBuilder
	HandleModuleMore() ModuleBuilder

	Build() (*jsonwebcard.Modules, error)
}

type webModuleBuilder struct {
	*jsonwebcard.MetaContext
	dynMaterials *dynmdlV2.DynamicContext

	moduleTag         *jsonwebcard.ModuleTag
	moduleAuthor      *jsonwebcard.ModuleAuthor
	moduleDynamic     *jsonwebcard.ModuleDynamic
	moduleDispute     *jsonwebcard.ModuleDispute
	moduleStat        *jsonwebcard.ModuleStat
	moduleInteraction *jsonwebcard.ModuleInteraction
	moduleShareInfo   *jsonwebcard.ModuleShareInfo
	moduleMore        *jsonwebcard.ModuleMore
}

func NewWebModuleBuilder(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	dynCtx.DynamicItem = &dynamicapi.DynamicItem{Extend: &dynamicapi.Extend{}} // 聚合结果
	dynCtx.Interim = &dynmdlV2.Interim{}                                       // 临时逻辑
	schemaCtx := &topiccardmodel.DynSchemaCtx{DynCtx: dynCtx}
	if err := topiccardschema.HandleDynamicCardBase(schemaCtx, &topiccardmodel.GeneralParam{
		Restriction: metaCtx.Restriction,
		Device:      metaCtx.Device,
		Mid:         metaCtx.Mid,
		IP:          metaCtx.IP,
		LocalTime:   metaCtx.LocalTime,
		Source:      topiccardmodel.BaseSourceWeb,
	}); err != nil {
		log.Error("cardbuilder.HandleDynamicCardBase mid=%d, dynCtx=%+v, error=%+v", metaCtx.Mid, dynCtx, err)
		return webModuleBuilder{MetaContext: &metaCtx, dynMaterials: schemaCtx.DynCtx}
	}
	return webModuleBuilder{MetaContext: &metaCtx, dynMaterials: schemaCtx.DynCtx}
}

func (mb webModuleBuilder) HandleModulePre(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	mb.moduleTag = handleModuleTag(metaCtx, dynCtx)
	return mb
}

func (mb webModuleBuilder) HandleModuleAuthor(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	switch {
	case dynCtx.Dyn.IsPGC():
		mb.moduleAuthor = handleModuleAuthorPgc(metaCtx, dynCtx)
	default:
		mb.moduleAuthor = handleModuleAuthorUgc(metaCtx, dynCtx)
	}
	if mb.moduleAuthor == nil || metaCtx.Config.ItemFromControl == nil {
		return mb
	}
	if val, ok := metaCtx.Config.ItemFromControl[dynCtx.Dyn.DynamicID]; ok && val == topicsvc.ItemFrom_TopShow.String() {
		// 置顶
		mb.moduleAuthor.IsTop = true
	}
	return mb
}

func (mb webModuleBuilder) HandleModuleDynamic(metaCtx jsonwebcard.MetaContext, cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	if dynCtx.Interim == nil {
		return mb
	}
	mb.moduleDynamic = handleModuleDynamic(metaCtx, cardType, dynCtx)
	return mb
}

func (mb webModuleBuilder) HandleModuleDispute(dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	mb.moduleDispute = handleModuleDispute(dynCtx)
	return mb
}

func (mb webModuleBuilder) HandleModuleStat(cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	mb.moduleStat = handleModuleStat(cardType, dynCtx)
	return mb
}

func (mb webModuleBuilder) HandleModuleInteraction(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	mb.moduleInteraction = handleModuleInteraction(metaCtx, dynCtx)
	return mb
}

func (mb webModuleBuilder) HandleModuleShareInfo(dynCtx *dynmdlV2.DynamicContext) ModuleBuilder {
	mb.moduleShareInfo = handleModuleShareInfo(dynCtx)
	return mb
}

func (mb webModuleBuilder) HandleModuleMore() ModuleBuilder {
	mb.moduleMore = handleModuleMore()
	return mb
}

func (mb webModuleBuilder) Build() (*jsonwebcard.Modules, error) {
	if mb.moduleAuthor == nil || mb.moduleDynamic == nil {
		return nil, errors.Errorf("Invalid web module built mb=%+v", mb.MetaContext)
	}
	return &jsonwebcard.Modules{
		ModuleTag:         mb.moduleTag,
		ModuleAuthor:      mb.moduleAuthor,
		ModuleDynamic:     mb.moduleDynamic,
		ModuleDispute:     mb.moduleDispute,
		ModuleStat:        mb.moduleStat,
		ModuleInteraction: mb.moduleInteraction,
		ModuleShareInfo:   mb.moduleShareInfo,
		ModuleMore:        mb.moduleMore,
	}, nil
}
