package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsononepic "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/one_pic"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	jsonthreepic "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/three_pic"

	"github.com/pkg/errors"
)

type CardBuilder interface {
	Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error)
}

type CardBuilderSingle interface {
	Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error)
}

func BuildOnePicV3FromPicture(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.OnePicV3).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := jsononepic.NewOnePicV3Builder(ctx)
	card, err := builder.SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		WithAfter(jsononepic.OnePicV3ByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildOnePicV2FromPicture(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.OnePicV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := jsononepic.NewOnePicV2Builder(ctx)
	card, err := builder.
		SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		WithAfter(jsononepic.OnePicV2ByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV2FromPicture(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DerivePictureBuilder().
		SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		WithAfter(jsonsmallcover.V2FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func BuildThreePicV3FromPicture(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.ThreePicV3).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := jsonthreepic.NewThreePicV3Builder(ctx)
	card, err := builder.SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		WithAfter(jsonthreepic.ThreePicV3ByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildThreePicV2FromPicture(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.ThreePicV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := jsonthreepic.NewThreePicV2Builder(ctx)
	card, err := builder.SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		WithAfter(jsonthreepic.ThreePicV2ByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildOnePicV1FromPicture(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.OnePicV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := jsononepic.NewOnePicV1Builder(ctx)
	card, err := builder.
		SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		SetAuthor(fanoutResult.Account.Card[picture.Mid]).
		WithAfter(jsononepic.OnePicV1ByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildThreePicV1FromPicture(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.ThreePicV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoPicture).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	card, err := jsonthreepic.NewThreePicV1Builder(ctx).
		SetBase(base).
		SetPicture(picture).
		SetRcmd(item).
		WithAfter(jsonthreepic.ThreePicV1ByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId])).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
