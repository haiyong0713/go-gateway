package jsoncarddynarticle

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicArticleCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicArticleCardBuilder
	SetBase(*jsonwebcard.Base) DynamicArticleCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicArticleCardBuilder

	Build() (*jsonwebcard.WebDynamicArticleCard, error)
}

type dynamicArticleCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicArticleCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicArticleCardBuilder {
	return dynamicArticleCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypeArticle}
}

func (b dynamicArticleCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicArticleCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicArticleCardBuilder) SetBase(base *jsonwebcard.Base) DynamicArticleCardBuilder {
	b.base = base
	return b
}

func (b dynamicArticleCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicArticleCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicArticleCardBuilder) Build() (*jsonwebcard.WebDynamicArticleCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	if b.dynMaterials.Dyn == nil {
		return nil, errors.Errorf("empty `dynMaterials.Dyn` field")
	}
	output := &jsonwebcard.WebDynamicArticleCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.ArticleCommentType, strconv.FormatInt(b.dynMaterials.Dyn.Rid, 10)
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
