package jsoncarddynforward

import (
	"strconv"

	"go-common/library/log"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicForwardCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicForwardCardBuilder
	SetBase(*jsonwebcard.Base) DynamicForwardCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicForwardCardBuilder

	Build() (*jsonwebcard.WebDynamicForwardCard, error)
}

type dynamicForwardCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicForwardCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicForwardCardBuilder {
	return dynamicForwardCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypeForward}
}

func (b dynamicForwardCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicForwardCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicForwardCardBuilder) SetBase(base *jsonwebcard.Base) DynamicForwardCardBuilder {
	b.base = base
	return b
}

func (b dynamicForwardCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicForwardCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicForwardCardBuilder) Build() (*jsonwebcard.WebDynamicForwardCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	output := &jsonwebcard.WebDynamicForwardCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.ForwardCommentType, strconv.FormatInt(b.dynMaterials.Dyn.DynamicID, 10)
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
	output.Modules = modules
	origin, err := makeDynamicForwardOriginCard(b)
	if err != nil {
		log.Error("dynamicForwardCardBuilder base=%+v, mid=%d, error=%+v", b.base, b.MetaContext.Mid, err)
		return nil, err
	}
	output.Orig = origin
	return output, nil
}

func makeDynamicForwardOriginCard(b dynamicForwardCardBuilder) (jsonwebcard.TopicCard, error) {
	var (
		dyn       = new(dynmdlV2.Dynamic)
		dynCtxTmp = new(dynmdlV2.DynamicContext)
	)
	// 感知转发卡逻辑
	*dyn = *b.dynMaterials.Dyn.Origin
	dyn.Forward = b.dynMaterials.Dyn
	*dynCtxTmp = *b.dynMaterials
	dynCtxTmp.Dyn = dyn

	// 筛选确定cardType
	cardType, err := makeTopicDynamicType(dyn)
	if err != nil {
		return nil, err
	}
	b.cardType = cardType

	// 构建module
	modules, err := cardbuilder.NewWebModuleBuilder(b.MetaContext, dynCtxTmp).
		HandleModuleAuthor(b.MetaContext, dynCtxTmp).
		HandleModuleDynamic(b.MetaContext, b.cardType, dynCtxTmp).
		HandleModuleDispute(dynCtxTmp).
		HandleModuleInteraction(b.MetaContext, dynCtxTmp).
		Build()
	if err != nil {
		return nil, err
	}
	if b.dynMaterials.Interim.IsPassCard {
		return nil, errors.Errorf("b.dynMaterials.Interim.IsPassCard==true 跳过当前卡片 metadata=%+v, dyn=%+v", b.MetaContext, b.dynMaterials.Dyn)
	}
	return &jsonwebcard.WebDynamicForwardCard{
		Base: &jsonwebcard.Base{
			IdStr:    strconv.FormatInt(dyn.DynamicID, 10),
			CardType: cardType,
			Visible:  b.dynMaterials.Dyn.Visible,
		},
		Basic:   cardbuilder.ConstructDynCardBasic(dyn),
		Modules: modules,
	}, nil
}

func makeTopicDynamicType(dyn *dynmdlV2.Dynamic) (jsonwebcard.CardType, error) {
	switch {
	case dyn.IsForward():
		return jsonwebcard.CardDynamicTypeForward, nil
	case dyn.IsAv():
		return jsonwebcard.CardDynamicTypeAv, nil
	case dyn.IsDraw():
		return jsonwebcard.CardDynamicTypeDraw, nil
	case dyn.IsWord():
		return jsonwebcard.CardDynamicTypeWord, nil
	case dyn.IsPGC():
		return jsonwebcard.CardDynamicTypePGC, nil
	case dyn.IsCommon():
		return jsonwebcard.CardDynamicTypeCommon, nil
	case dyn.IsArticle():
		return jsonwebcard.CardDynamicTypeArticle, nil
	}
	return "", errors.Errorf("Unexpected dynamic type=%+v, dynId=%d", dyn.Type, dyn.DynamicID)
}
