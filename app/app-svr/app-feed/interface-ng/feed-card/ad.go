package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/cm"
	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"

	"github.com/pkg/errors"
)

func BuildCmV2AdWebS(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV2BuilderFactory(ctx)
	card, err := factory.DeriveAdWebsBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV2AdAv(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archive, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.CmV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := cm.NewCmV2BuilderFactory(ctx)
	card, err := factory.DeriveAdAvBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetChannelCard(fanoutResult.Channel[item.Tid]).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetCoverGif(item.CoverGif).
		SetAdInfo(item.CardStatusAd()).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV2AdWeb(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV2BuilderFactory(ctx)
	card, err := factory.DeriveAdWebBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV2AdPlayer(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV2BuilderFactory(ctx)
	card, err := factory.DeriveAdPlayerBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV1AdWeb(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetCardLen(CardLenOnIPad(ctx, 2, 0)).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV1BuilderFactory(ctx)
	card, err := factory.DeriveAdWebBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV2AdInlineLive(ctx cardschema.FeedContext, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV2BuilderFactory(ctx)
	card, err := factory.DeriveAdInlineBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV1AdWebS(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetCardLen(CardLenOnIPad(ctx, 1, 0)).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV1BuilderFactory(ctx)
	card, err := factory.DeriveAdWebBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmSingleV1AdWebS(ctx cardschema.FeedContext, _ int64, item *ai.Item, _ *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmSingleV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV1BuilderFactory(ctx)
	card, err := factory.DeriveAdWebBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV1AdAv(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archive, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.CmV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetCardLen(CardLenOnIPad(ctx, 1, 0)).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := cm.NewCmV1BuilderFactory(ctx)
	card, err := factory.DeriveAdAvBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetChannelCard(fanoutResult.Channel[item.Tid]).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetCoverGif(item.CoverGif).
		SetAdInfo(item.CardStatusAd()).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV9FromArchive(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archive, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.CmDoubleV9).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveArcPlayerV2Builder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		SetStoryIcon(fanoutResult.StoryIcon).
		SetHasFav(fanoutResult.Favourite).
		SetHotAidSet(fanoutResult.HotAidSet).
		SetHasCoin(fanoutResult.Coin).
		SetLikeStatState(fanoutResult.LikeStatState).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmSingleV9FromArchive(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archive, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.CmSingleV9).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetAdInfo(item.CardStatusAd()).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveSingleArcPlayerBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		SetStoryIcon(fanoutResult.StoryIcon).
		SetHasFav(fanoutResult.Favourite).
		SetHotAidSet(fanoutResult.HotAidSet).
		SetHasCoin(fanoutResult.Coin).
		SetLikeStatState(fanoutResult.LikeStatState).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV2FromReservation(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildAdReservation(ctx, appcardmodel.CmV2, index, item, fanoutResult)
}

func BuildCmV1FromReservation(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildAdReservation(ctx, appcardmodel.CmSingleV1, index, item, fanoutResult)
}

func buildAdReservation(ctx cardschema.FeedContext, cardType appcardmodel.CardType, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
	reservation, ok := fanoutResult.Reservation[item.ID]
	if !ok || reservation == nil {
		return nil, errors.Errorf("找不到对应的预约id物料: %d", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(cardType).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	factory := cm.NewCmV2BuilderFactory(ctx)
	card, err := factory.DeriveAdReservation().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		SetReservation(reservation).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV1AdPlayer(ctx cardschema.FeedContext, index int64, item *ai.Item, result *FanoutResult) (cardschema.FeedCard, error) {
	return buildSingleAdPlayer(ctx, appcardmodel.CmV1, index, item, result)
}

func BuildCmSingleV1AdPlayer(ctx cardschema.FeedContext, index int64, item *ai.Item, result *FanoutResult) (cardschema.FeedCard, error) {
	return buildSingleAdPlayer(ctx, appcardmodel.CmSingleV1, index, item, result)
}

func buildSingleAdPlayer(ctx cardschema.FeedContext, cardType appcardmodel.CardType, _ int64, item *ai.Item, _ *FanoutResult) (cardschema.FeedCard, error) {
	if item.CardStatusAd() == nil {
		return nil, errors.Errorf("ad is not exist")
	}
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
	factory := cm.NewCmV1BuilderFactory(ctx)
	card, err := factory.DeriveAdPlayerBuilder().
		SetBase(base).
		SetAdInfo(item.CardStatusAd()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmV2FromPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	episode, ok := fanoutResult.Bangumi.PgcEpisodeByEpids[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("ad episode not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.CmV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetAdInfo(item.CardStatusAd()).
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
	card, err := factory.DeriveEpPGCBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(episode).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildCmV7FromInlinePGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("ad inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.CmDoubleV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPGC).
		SetMetricRcmd(item).
		SetAdInfo(item.CardStatusAd()).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DerivePgcBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(inlinePgc).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		WithAfter(jsonlargecover.DoubleInlineDbClickLike(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildCmSingleV7WithPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("ad single inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.CmSingleV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPGC).
		SetMetricRcmd(item).
		SetAdInfo(item.CardStatusAd()).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveSingleBangumiBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(inlinePgc).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		WithAfter(jsonlargecover.SingleInlineDbClickLike(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
