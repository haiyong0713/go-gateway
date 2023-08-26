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
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"

	"github.com/pkg/errors"
)

func BuildSmallCoverV2FromLiveRoom(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	room, ok := fanoutResult.Live.Room[item.ID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoLive).
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
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(jsonsmallcover.V2FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverV8FromLiveRoom(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	room, ok := fanoutResult.Live.InlineRoom[item.ID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV8).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetInline(fanoutResult.Inline).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(func(inline *jsoncard.LargeCoverInline) {
			if item.LiveInlineDanmu == 0 {
				inline.DisableDanmu = true
				inline.HideDanmuSwitch = true
			}
		}).
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverV1FromLiveRoom(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	room, ok := fanoutResult.Live.Room[item.ID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCardLen(CardLenOnIPad(ctx, 1, 0)).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecoverv1.NewLargeCoverV1Builder(ctx)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(jsonlargecoverv1.V1FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverSingleV8(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	room, ok := fanoutResult.Live.InlineRoom[item.ID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV8).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetInline(fanoutResult.Inline).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.SingleInlineLiveHideMeta()).
		WithAfter(jsonlargecover.SingleInlineLivePrivateVal(room, item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildLargeCoverSingleV8WithInlineLive(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	room, ok := fanoutResult.Live.InlineRoom[item.ID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV8).
		SetCardGoto(appcardmodel.CardGotoLive). // inline_live单列goto转换为live，直播与客户端约定根据goto添加jump_from
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetInline(fanoutResult.Inline).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.SingleInlineLiveHideMeta()).
		WithAfter(jsonlargecover.SingleInlineLivePrivateVal(room, item)).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV9FromLiveRoom(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	room, ok := fanoutResult.Live.Room[item.ID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV9).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoLive).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcover.NewV9LiveBuilder(ctx)
	card, err := builder.SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		SetLeftBottomBadgeStyle(fanoutResult.LiveBadge.LeftBottomBadgeStyle).
		SetLeftCoverBadgeStyle(fanoutResult.LiveBadge.LeftCoverBadgeStyle).
		WithAfter(jsonsmallcover.V9FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		WithAfter(jsonsmallcover.SmallCoverV9TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildLargeCoverSingleV8FromSpecialS(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "live" {
		return nil, errors.Errorf("special_s not exist")
	}
	room, ok := fanoutResult.Live.InlineRoom[item.SingleSpecialInfo.SpID]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV8).
		SetCardGoto(appcardmodel.CardGotoLive). // inline_live单列goto转换为live，直播与客户端约定根据goto添加jump_from
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetInline(fanoutResult.Inline).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(jsonlargecover.InlineFilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item)).
		WithAfter(jsonlargecover.SingleInlineLiveHideMeta()).
		WithAfter(jsonlargecover.SingleInlineLivePrivateVal(room, item)).
		WithAfter(jsonlargecover.LargeCoverInlineFromSpecialS(fanoutResult.Specials[item.ID])).
		WithAfter(jsonlargecover.LargeCoverInlineTalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
