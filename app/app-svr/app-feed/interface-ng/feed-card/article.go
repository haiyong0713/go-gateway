package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonlargecoverv1 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover_v1"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	jsonthreeitemhv3 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/three_item_h_v3"

	"github.com/pkg/errors"
)

func BuildSmallCoverV2FromArticle(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	article, ok := fanoutResult.Article[item.ID]
	if !ok {
		return nil, errors.Errorf("article not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoArticle).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveArticleBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArticle(article).
		WithAfter(jsonsmallcover.V2FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildThreeItemHV3FromArticle(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	article, ok := fanoutResult.Article[item.ID]
	if !ok {
		return nil, errors.Errorf("article not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.ThreeItemHV3).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoArticle).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	card, err := jsonthreeitemhv3.NewThreeItemHV3BuilderBuilder(ctx).
		SetBase(base).
		SetRcmd(item).
		SetArticle(article).
		SetAuthorCard(fanoutResult.Account.Card[article.Author.Mid]).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

// ipad 专栏卡
func BuildLargeCoverV1FromArticle(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	article, ok := fanoutResult.Article[item.ID]
	if !ok {
		return nil, errors.Errorf("article not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoArticle).
		SetMetricRcmd(item).
		SetCardLen(1).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecoverv1.NewLargeCoverV1Builder(ctx)
	card, err := factory.DeriveArticleBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArticle(article).
		SetAuthorCard(fanoutResult.Account.Card[article.Author.Mid]).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}
