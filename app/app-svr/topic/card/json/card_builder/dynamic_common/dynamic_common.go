package jsoncarddyncommon

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicCommonCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicCommonCardBuilder
	SetBase(*jsonwebcard.Base) DynamicCommonCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicCommonCardBuilder

	Build() (*jsonwebcard.WebDynamicCommonCard, error)
}

type dynamicCommonCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicCommonCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicCommonCardBuilder {
	return dynamicCommonCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypeCommon}
}

func (b dynamicCommonCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicCommonCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicCommonCardBuilder) SetBase(base *jsonwebcard.Base) DynamicCommonCardBuilder {
	b.base = base
	return b
}

func (b dynamicCommonCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicCommonCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicCommonCardBuilder) Build() (*jsonwebcard.WebDynamicCommonCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	if b.dynMaterials.Dyn == nil {
		return nil, errors.Errorf("empty `dynMaterials.Dyn` field")
	}
	output := &jsonwebcard.WebDynamicCommonCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.CommonCommentType, strconv.FormatInt(b.dynMaterials.Dyn.DynamicID, 10)
	output.Basic = basic
	modules, err := cardbuilder.NewWebModuleBuilder(b.MetaContext, b.dynMaterials).
		HandleModulePre(b.MetaContext, b.dynMaterials).
		HandleModuleAuthor(b.MetaContext, b.dynMaterials).
		HandleModuleDynamic(b.MetaContext, b.cardType, b.dynMaterials).
		HandleModuleDispute(b.dynMaterials).
		HandleModuleStat(b.cardType, b.dynMaterials).
		HandleModuleInteraction(b.MetaContext, b.dynMaterials).
		HandleModuleMore().
		Build()
	if err != nil {
		return nil, err
	}
	if b.dynMaterials.Interim.IsPassCard {
		return nil, errors.Errorf("b.dynMaterials.Interim.IsPassCard==true 跳过当前卡片 metadata=%+v, dyn=%+v", b.MetaContext, b.dynMaterials.Dyn)
	}
	output.Modules = modules

	return output, nil
}
