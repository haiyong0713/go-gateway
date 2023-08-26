package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	v6 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover_v6"
	v7 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover_v7"

	"github.com/pkg/errors"
)

func BuildSmallCoverV7FromVip(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Vip == nil {
		return nil, errors.New("empty vip")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV7).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoVip)).
		SetGoto(appcardmodel.GotoWeb).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	build := v7.NewV7VipBuilder(ctx)
	card, err := build.SetBase(base).SetVip(fanoutResult.Vip).SetRcmd(item).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV6FromVip(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Vip == nil {
		return nil, errors.New("empty vip")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV6).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoVip)).
		SetGoto(appcardmodel.GotoWeb).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	build := v6.NewV6VipBuilder(ctx)
	card, err := build.SetBase(base).SetVip(fanoutResult.Vip).SetRcmd(item).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
