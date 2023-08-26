package jsoncarddynpgc

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	"go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

type DynamicPGCCardBuilder interface {
	ReplaceContext(jsonwebcard.MetaContext) DynamicPGCCardBuilder
	SetBase(*jsonwebcard.Base) DynamicPGCCardBuilder
	SetDynCtx(*dynmdlV2.DynamicContext) DynamicPGCCardBuilder

	Build() (*jsonwebcard.WebDynamicPGCCard, error)
}

type dynamicPGCCardBuilder struct {
	MetaContext  jsonwebcard.MetaContext
	base         *jsonwebcard.Base
	dynMaterials *dynmdlV2.DynamicContext
	cardType     jsonwebcard.CardType
}

func NewDynamicPGCCardBuilder(metaCtx jsonwebcard.MetaContext) DynamicPGCCardBuilder {
	return dynamicPGCCardBuilder{MetaContext: metaCtx, cardType: jsonwebcard.CardDynamicTypePGC}
}

func (b dynamicPGCCardBuilder) ReplaceContext(metaCtx jsonwebcard.MetaContext) DynamicPGCCardBuilder {
	b.MetaContext = metaCtx
	return b
}

func (b dynamicPGCCardBuilder) SetBase(base *jsonwebcard.Base) DynamicPGCCardBuilder {
	b.base = base
	return b
}

func (b dynamicPGCCardBuilder) SetDynCtx(dynCtx *dynmdlV2.DynamicContext) DynamicPGCCardBuilder {
	b.dynMaterials = dynCtx
	return b
}

func (b dynamicPGCCardBuilder) Build() (*jsonwebcard.WebDynamicPGCCard, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.dynMaterials == nil {
		return nil, errors.Errorf("empty `dynMaterials` field")
	}
	if b.dynMaterials.Dyn == nil {
		return nil, errors.Errorf("empty `dynMaterials.Dyn` field")
	}
	output := &jsonwebcard.WebDynamicPGCCard{Base: b.base}
	basic := cardbuilder.ConstructDynCardBasic(b.dynMaterials.Dyn)
	basic.CommentType, basic.CommentIdStr = model.PgcCommentType, makePGCCommentIdStr(b.dynMaterials)
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

func makePGCCommentIdStr(dynCtx *dynmdlV2.DynamicContext) string {
	if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid)); ok {
		return strconv.FormatInt(pgc.Aid, 10)
	}
	return ""
}
