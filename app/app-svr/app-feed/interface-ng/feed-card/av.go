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
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/storys"
	"go-gateway/app/app-svr/app-feed/interface/model"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

func BuildSmallCoverV2FromArchive(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.Archive
	if appcardmodel.Gt(item.JumpGoto) == appcardmodel.GotoVerticalAv && ctx.VersionControl().Can("archive.storyPlayerSupported") {
		archiveStore = fanoutResult.Archive.StoryArchive
	}
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
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
	card, err := factory.DeriveArcPlayerBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetChannelCard(fanoutResult.Channel[item.Tid]).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetCoverGif(item.CoverGif).
		SetStoryIcon(fanoutResult.StoryIcon).
		SetOpenCourseMark(fanoutResult.OpenCourseMark).
		WithAfter(jsonsmallcover.SmallCoverV2AVCustomizedQuality(item, archive)).
		WithAfter(jsonsmallcover.SmallCoverV2AVCustomizedDesc(item, archive, fanoutResult.Tag[item.Tid], fanoutResult.Account.Card[archive.Arc.Author.Mid])).
		WithAfter(jsonsmallcover.V2FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		WithAfter(func(card *jsoncard.SmallCoverV2) {
			// nolint:gomnd
			if item.StNewCover == 2 {
				card.GotoIcon = nil
				card.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "竖屏")
			}
		}).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverV6FromArchive(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.Archive
	if appcardmodel.Gt(item.JumpGoto) == appcardmodel.GotoVerticalAv {
		archiveStore = fanoutResult.Archive.StoryArchive
	}
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV6).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveArcPlayerBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		SetStoryIcon(fanoutResult.StoryIcon).
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverV5FromArchive(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.Archive
	if appcardmodel.Gt(item.JumpGoto) == appcardmodel.GotoVerticalAv {
		archiveStore = fanoutResult.Archive.StoryArchive
	}
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV5).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonlargecover.NewLargeCoverV5Builder(ctx)
	card, err := builder.SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetChannelCard(fanoutResult.Channel[item.Tid]).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverV9FromArchive(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.Archive
	if appcardmodel.Gt(item.JumpGoto) == appcardmodel.GotoVerticalAv {
		archiveStore = fanoutResult.Archive.StoryArchive
	}
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV9).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
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
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildStorysV2(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildStorys(ctx, appcardmodel.StorysV2, index, item, fanoutResult)
}

func BuildStorysV1(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	return buildStorys(ctx, appcardmodel.StorysV1, index, item, fanoutResult)
}

func BuildLargeCoverV1FromArchive(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.Archive
	if appcardmodel.Gt(item.JumpGoto) == appcardmodel.GotoVerticalAv && ctx.VersionControl().Can("archive.storyPlayerSupported") {
		archiveStore = fanoutResult.Archive.StoryArchive
	}
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
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
	card, err := factory.DeriveArcPlayerBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetChannelCard(fanoutResult.Channel[item.Tid]).
		SetTag(TagOnIPad(ctx, nil, fanoutResult.Tag[item.Tid])).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetCoverGif(item.CoverGif).
		SetStoryIcon(fanoutResult.StoryIcon).
		WithAfter(jsonlargecoverv1.V1FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildStorys(ctx cardschema.FeedContext, cardType appcardmodel.CardType, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(cardType).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}
	builder := storys.NewStorysBuilder(ctx)
	card, err := builder.SetBase(base).
		SetRcmd(item).
		SetArcPlayer(fanoutResult.Archive.StoryArchive).
		SetTags(fanoutResult.Tag).
		SetAuthorCard(fanoutResult.Account.Card).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func CardLenOnIPad(ctx cardschema.FeedContext, onIPad int64, default_ int64) int64 {
	if model.IsPad(ctx.Device().Plat()) {
		return onIPad
	}
	return default_
}

func TagOnIPad(ctx cardschema.FeedContext, onIPad *taggrpc.Tag, default_ *taggrpc.Tag) *taggrpc.Tag {
	if model.IsPad(ctx.Device().Plat()) {
		return onIPad
	}
	return default_
}

func BuildLargeCoverSingleV9(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.Archive
	if appcardmodel.Gt(item.JumpGoto) == appcardmodel.GotoVerticalAv {
		archiveStore = fanoutResult.Archive.StoryArchive
	}
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV9).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
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
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverSingleV9FromSpecialS(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "av" {
		return nil, errors.Errorf("special_s not exist")
	}
	archiveStore := fanoutResult.Archive.Archive
	archive, ok := archiveStore[item.SingleSpecialInfo.SpID]
	if !ok {
		return nil, errors.Errorf("archvie not exist: %d", item.SingleSpecialInfo.SpID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV9).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
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
		WithAfter(jsonlargecover.LargeCoverInlineFromSpecialS(fanoutResult.Specials[item.ID])).
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV11FromArchive(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	archiveStore := fanoutResult.Archive.StoryArchive
	archive, ok := archiveStore[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV11).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoVerticalAvV2).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcover.NewV11VerticalBuilder(ctx)
	card, err := builder.SetBase(base).
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetStoryIcon(fanoutResult.StoryIcon).
		WithAfter(jsonsmallcover.V11FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(func(card *jsoncard.SmallCoverV11) {
			if card.RcmdReason == "已关注" {
				card.CoverLeftText1 = ""
				card.CoverLeftIcon1 = 0
			}
		}).
		WithAfter(jsonsmallcover.SmallCoverV11TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
