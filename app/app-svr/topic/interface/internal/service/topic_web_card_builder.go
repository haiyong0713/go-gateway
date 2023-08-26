package service

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	cardbuilder "go-gateway/app/app-svr/topic/card/json/card_builder"
	jsoncarddynarticle "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_article"
	jsoncarddynav "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_av"
	jsoncarddyncommon "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_common"
	jsoncarddyndraw "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_draw"
	jsoncarddynforward "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_forward"
	jsoncarddynpgc "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_pgc"
	jsoncarddynword "go-gateway/app/app-svr/topic/card/json/card_builder/dynamic_word"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

// 话题详情页流卡片构造器
type CardBuilder interface {
	Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error)
	BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard
}

func webDynCardGetBuilder(dyn *dynmdlV2.Dynamic) (CardBuilder, bool) {
	switch {
	case dyn.IsForward():
		return dynForward{}, true
	case dyn.IsAv():
		return dynAv{}, true
	case dyn.IsDraw():
		return dynDraw{}, true
	case dyn.IsWord():
		return dynWord{}, true
	case dyn.IsArticle():
		return dynArticle{}, true
	case dyn.IsCommon():
		return dynCommon{}, true
	case dyn.IsPGC():
		return dynPGC{}, true
	}
	return nil, false
}

// 视频卡
type dynAv struct{}

func (dynAv) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypeAv,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddynav.NewDynamicAvCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynAv) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	return cardbuilder.BackfillCard(card, dynCtx)
}

// 图文卡
type dynDraw struct{}

func (dynDraw) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypeDraw,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddyndraw.NewDynamicDrawCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynDraw) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	return cardbuilder.BackfillCard(card, dynCtx)
}

// 转发卡
type dynForward struct{}

func (dynForward) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypeForward,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddynforward.NewDynamicForwardCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynForward) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	// 转发原卡文案
	v, ok := card.(*jsonwebcard.WebDynamicForwardCard)
	if ok {
		v.Orig = *cardbuilder.BackfillCard(v.Orig, dynCtx)
	}
	return cardbuilder.BackfillCard(card, dynCtx)
}

// 纯文字卡
type dynWord struct{}

func (dynWord) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypeWord,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddynword.NewDynamicWordCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynWord) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	return cardbuilder.BackfillCard(card, dynCtx)
}

func resolveDynSchemaCtxForWeb(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) jsonwebcard.MetaContext {
	return jsonwebcard.MetaContext{
		Restriction: general.Restriction,
		Device:      general.Device,
		Mid:         general.Mid,
		IP:          general.IP,
		LocalTime:   general.LocalTime,
		Config:      &jsonwebcard.Config{DynCmtTopicControl: dynSchemaCtx.DynCmtMode, ItemFromControl: dynSchemaCtx.ItemFrom, HiddenAttached: dynSchemaCtx.HiddenAttached},
	}
}

// 专栏卡
type dynArticle struct{}

func (dynArticle) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypeArticle,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddynarticle.NewDynamicArticleCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynArticle) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	return cardbuilder.BackfillCard(card, dynCtx)
}

// 通用模板
type dynCommon struct{}

func (dynCommon) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypeCommon,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddyncommon.NewDynamicCommonCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynCommon) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	return cardbuilder.BackfillCard(card, dynCtx)
}

// ogv视频
type dynPGC struct{}

func (dynPGC) Build(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) (jsonwebcard.TopicCard, error) {
	metaCtx := resolveDynSchemaCtxForWeb(dynSchemaCtx, general)
	base := &jsonwebcard.Base{
		IdStr:    strconv.FormatInt(dynSchemaCtx.DynCtx.Dyn.DynamicID, 10),
		CardType: jsonwebcard.CardDynamicTypePGC,
		Visible:  dynSchemaCtx.DynCtx.Dyn.Visible,
		TopicId:  dynSchemaCtx.TopicId,
	}
	return jsoncarddynpgc.NewDynamicPGCCardBuilder(metaCtx).
		SetBase(base).
		SetDynCtx(dynSchemaCtx.DynCtx).
		Build()
}

func (dynPGC) BackFill(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	return cardbuilder.BackfillCard(card, dynCtx)
}
