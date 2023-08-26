package jsoncarddynword

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicWordCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicWordCardBuilder
	SetBase(*jsonwebcard.Base) DynamicWordCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicWordCardBuilder

	Build() (*jsonwebcard.WebDynamicWordCard, error)
}

type dynamicWordCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicWordCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicWordCardBuilder {
	return dynamicWordCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypeWord}
}

func (b dynamicWordCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicWordCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicWordCardBuilder) SetBase(base *jsonwebcard.Base) DynamicWordCardBuilder {
	b.base = base
	return b
}

func (b dynamicWordCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicWordCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicWordCardBuilder) Build() (*jsonwebcard.WebDynamicWordCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	output := &jsonwebcard.WebDynamicWordCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.WordCommentType, strconv.FormatInt(b.dynMaterials.Dyn.DynamicID, 10)
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
