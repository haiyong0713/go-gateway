package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	jsonlargecoverv1 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover_v1"
	jsonogvsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/ogv_small_cover"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	jsonsmallcoverv1 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover_v1"

	"github.com/pkg/errors"
)

func BuildLargeCoverV7FromPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPGC).
		SetMetricRcmd(item).
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
		WithAfter(jsonlargecover.InlineFilledByEpMaterials(fanoutResult.Bangumi.EpMaterial[item.CreativeId])).
		//WithAfter(jsonlargecover.InlineReplacedByRcmd(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV2FromBangumi(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	episode, ok := fanoutResult.Bangumi.SeasonByAid[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("episode not exist: %d", int32(item.ID))
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
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
	card, err := factory.DeriveEpBangumiBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(episode).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetArchive(fanoutResult.Archive.Archive[item.ID]).
		WithAfter(jsonsmallcover.V2FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, false)).
		WithAfter(jsonsmallcover.V2ReplacedByRcmd(item)).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildSmallCoverV2FromPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	episode, ok := fanoutResult.Bangumi.PgcEpisodeByEpids[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("episode not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetParam(strconv.FormatInt(item.ID, 10)).
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
		WithAfter(jsonsmallcover.V2FilledByEpMaterials(fanoutResult.Bangumi.EpMaterial[item.CreativeId], item)).
		WithAfter(jsonsmallcover.V2ReplacedByRcmd(item)).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		WithAfter(func(v2 *jsoncard.SmallCoverV2) {
			if episode.GetType() != nil && !episode.GetType().GetIsFormal() {
				v2.Badge = ""
				v2.BadgeStyle = nil
			}
		}).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildSmallCoverV4FromRemind(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Bangumi.Remind == nil {
		return nil, errors.Errorf("Remind is empty")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV4).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcover.NewV4RemindBuilder(ctx)
	card, err := builder.SetBase(base).SetRcmd(item).SetRemind(fanoutResult.Bangumi.Remind).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV4FromUpdate(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Bangumi.Update == nil {
		return nil, errors.Errorf("Update is empty")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV4).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcover.NewV4UpdateBuilder(ctx)
	card, err := builder.SetBase(base).SetRcmd(item).SetUpdate(fanoutResult.Bangumi.Update).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV4FromBangumiRcmd(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if ctx.VersionControl().Can("feed.usingRemind") {
		return BuildSmallCoverV4FromRemind(ctx, index, item, fanoutResult)
	}
	return BuildSmallCoverV4FromUpdate(ctx, index, item, fanoutResult)
}

func FixtureForIOS617(ctx cardschema.FeedContext) func(*jsoncard.SmallCoverV1) {
	return func(in *jsoncard.SmallCoverV1) {
		if !ctx.VersionControl().Can("feed.isIOS617") {
			return
		}
		if in.CoverBadgeStyle == nil {
			in.CoverBadgeStyle = &jsoncard.ReasonStyle{
				Text:    " ",
				BgStyle: 1,
			}
		}
		if in.LeftCoverBadgeNewStyle == nil {
			in.LeftCoverBadgeNewStyle = &jsoncard.ReasonStyle{
				IconURL:      "https://i0.hdslb.com/bfs/feed-admin/084f3275802a6bb2797a0a1ba106e676c04ce2e1.png",
				IconURLNight: "https://i0.hdslb.com/bfs/feed-admin/8d345f61fa8081ef2097d71740531857b59ed06e.png",
			}
		}
	}
}

func BuildSmallCoverV1FromRemind(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Bangumi.Remind == nil {
		return nil, errors.Errorf("Remind is empty")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcoverv1.NewV1BangumiRemindBuilder(ctx)
	card, err := builder.SetBase(base).
		SetBangumiRemind(fanoutResult.Bangumi.Remind).
		SetRcmd(item).
		WithAfter(FixtureForIOS617(ctx)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV1FromUpdate(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Bangumi.Update == nil {
		return nil, errors.Errorf("Update is empty")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcoverv1.NewV1BangumiUpdateBuilder(ctx)
	card, err := builder.SetBase(base).
		SetBangumiUpdate(fanoutResult.Bangumi.Update).
		SetRcmd(item).
		WithAfter(FixtureForIOS617(ctx)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV1FromBangumiRcmd(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if ctx.VersionControl().Can("feed.usingRemind") {
		return BuildSmallCoverV1FromRemind(ctx, index, item, fanoutResult)
	}
	return BuildSmallCoverV1FromUpdate(ctx, index, item, fanoutResult)
}

func BuildLargeCoverV1FromBangumi(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	season, ok := fanoutResult.Bangumi.SeasonByAid[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("season not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetMetricRcmd(item).
		SetCardLen(CardLenOnIPad(ctx, 1, 0)).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}
	factory := jsonlargecoverv1.NewLargeCoverV1Builder(ctx)
	card, err := factory.DeriveEpBangumiBuilder().
		SetBase(base).
		SetRcmd(item).
		SetBangumiSeason(season).
		SetTag(TagOnIPad(ctx, nil, fanoutResult.Tag[item.Tid])).
		WithAfter(jsonlargecoverv1.V1FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, false)).
		WithAfter(jsonlargecoverv1.V1ReplacedByRcmd(item)).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverV1FromPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	episode, ok := fanoutResult.Bangumi.Season[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("episode not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetMetricRcmd(item).
		SetCardLen(CardLenOnIPad(ctx, 1, 0)).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecoverv1.NewLargeCoverV1Builder(ctx)
	card, err := factory.DeriveEpPGCBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(episode).
		WithAfter(jsonlargecoverv1.V1FilledByEpMaterials(fanoutResult.Bangumi.EpMaterial[item.CreativeId], item)).
		WithAfter(jsonlargecoverv1.V1ReplacedByRcmd(item)).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildOgvSmallCoverFromPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	episode, ok := fanoutResult.Bangumi.PgcEpisodeByAids[item.ID]
	if !ok {
		return nil, errors.Errorf("episode not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.OgvSmallCover).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetCardLen(1).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	card, err := jsonogvsmallcover.NewOgvSmallCoverBuilder(ctx).
		SetBase(base).
		SetRcmd(item).
		SetEpisode(episode).
		SetEpMaterilas(fanoutResult.Bangumi.EpMaterial[item.CreativeId]).
		WithAfter(jsonogvsmallcover.OGVSmallCoverTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverSingleV7(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[item.Epid]
	if !ok {
		return nil, errors.Errorf("inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(int64(item.Epid), 10)).
		SetCardType(appcardmodel.LargeCoverSingleV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoBangumi).
		SetMetricRcmd(item).
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
		WithAfter(jsonlargecover.InlineFilledByEpMaterials(fanoutResult.Bangumi.EpMaterial[item.CreativeId])).
		WithAfter(jsonlargecover.InlineReplacedByRcmd(item)).
		WithAfter(jsonlargecover.SingleInlineDbClickLike(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverSingleV7WithInlinePGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPGC).
		SetMetricRcmd(item).
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
		WithAfter(jsonlargecover.InlineFilledByEpMaterials(fanoutResult.Bangumi.EpMaterial[item.CreativeId])).
		//WithAfter(jsonlargecover.InlineReplacedByRcmd(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverSingleV7WithPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPGC).
		SetMetricRcmd(item).
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
		WithAfter(jsonlargecover.InlineFilledByEpMaterials(fanoutResult.Bangumi.EpMaterial[item.CreativeId])).
		WithAfter(jsonlargecover.InlineReplacedByRcmd(item)).
		WithAfter(jsonlargecover.SingleInlineDbClickLike(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverSingleV7FromSpecialS(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "pgc" {
		return nil, errors.Errorf("special_s not exist")
	}
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.SingleSpecialInfo.SpID)]
	if !ok {
		return nil, errors.Errorf("inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoBangumi).
		SetMetricRcmd(item).
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
		WithAfter(jsonlargecover.LargeCoverInlineFromSpecialS(fanoutResult.Specials[item.ID])).
		//WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.OgvCreativeId])).
		//WithAfter(jsonlargecover.InlineReplacedByRcmd(item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
