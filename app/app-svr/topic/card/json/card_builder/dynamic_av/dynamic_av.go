package jsoncarddynav

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicAvCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicAvCardBuilder
	SetBase(*jsonwebcard.Base) DynamicAvCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicAvCardBuilder

	Build() (*jsonwebcard.WebDynamicAvCard, error)
}

type dynamicAvCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicAvCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicAvCardBuilder {
	return dynamicAvCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypeAv}
}

func (b dynamicAvCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicAvCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicAvCardBuilder) SetBase(base *jsonwebcard.Base) DynamicAvCardBuilder {
	b.base = base
	return b
}

func (b dynamicAvCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicAvCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicAvCardBuilder) Build() (*jsonwebcard.WebDynamicAvCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	if b.dynMaterials.Dyn == nil {
		return nil, errors.Errorf("empty `dynMaterials.Dyn` field")
	}
	output := &jsonwebcard.WebDynamicAvCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.VideoCommentType, strconv.FormatInt(b.dynMaterials.Dyn.Rid, 10)
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
