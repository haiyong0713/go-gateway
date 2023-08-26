package feedcard

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonbanner "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/banner"
)

func BuildBannerV5(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildBanner(ctx, appcardmodel.BannerV5, index, item, fanoutResult)
}

func BuildBannerV4(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildBanner(ctx, appcardmodel.BannerV4, index, item, fanoutResult)
}

func BuildBannerV6(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildBanner(ctx, appcardmodel.BannerV6, index, item, fanoutResult)
}

func buildBanner(ctx cardschema.FeedContext, cardType appcardmodel.CardType, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(cardType).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonbanner.NewBannerV5Builder(ctx)
	card, err := builder.SetBase(base).
		SetRcmd(item).
		SetBanners(fanoutResult.Banner.Banners).
		SetVersion(fanoutResult.Banner.Version).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildBannerIPadV8(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildBanner(ctx, appcardmodel.BannerIPadV8, index, item, fanoutResult)
}

func BuildBannerV8(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildInlineBanner(ctx, appcardmodel.BannerV8, index, item, fanoutResult)
}

func BuildBannerSingleV8(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildInlineBanner(ctx, appcardmodel.BannerSingleV8, index, item, fanoutResult)
}

func buildInlineBanner(ctx cardschema.FeedContext, cardType appcardmodel.CardType, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(cardType).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonbanner.NewBannerInlineBuilder(ctx)
	card, err := builder.SetBase(base).SetRcmd(item).SetBanners(fanoutResult.Banner.Banners).
		SetVersion(fanoutResult.Banner.Version).SetArcPlayer(fanoutResult.Archive.Archive).
		SetEpisode(fanoutResult.Bangumi.InlinePGC).SetLiveRoom(fanoutResult.Live.InlineRoom).
		SetAuthorCard(fanoutResult.Account.Card).SetInline(fanoutResult.Inline).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
