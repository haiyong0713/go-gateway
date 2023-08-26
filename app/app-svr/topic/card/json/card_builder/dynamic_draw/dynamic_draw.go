package jsoncarddyndraw

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicDrawCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicDrawCardBuilder
	SetBase(*jsonwebcard.Base) DynamicDrawCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicDrawCardBuilder

	Build() (*jsonwebcard.WebDynamicDrawCard, error)
}

type dynamicDrawCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicDrawCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicDrawCardBuilder {
	return dynamicDrawCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypeDraw}
}

func (b dynamicDrawCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicDrawCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicDrawCardBuilder) SetBase(base *jsonwebcard.Base) DynamicDrawCardBuilder {
	b.base = base
	return b
}

func (b dynamicDrawCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicDrawCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicDrawCardBuilder) Build() (*jsonwebcard.WebDynamicDrawCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	output := &jsonwebcard.WebDynamicDrawCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.DrawCommentType, strconv.FormatInt(b.dynMaterials.Dyn.Rid, 10)
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
