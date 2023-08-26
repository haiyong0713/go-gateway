package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover_v10"

	"github.com/pkg/errors"
)

func BuildSmallCoverV10FromGame(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	game, ok := fanoutResult.Game[item.ID]
	if !ok {
		return nil, errors.Errorf("game: %d not exist", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV10).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoGame).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		SetCreativeId(item.CreativeId).
		Build()
	if err != nil {
		return nil, err
	}

	builder := small_cover_v10.NewV10GameBuilder(ctx)
	card, err := builder.SetBase(base).
		SetRcmd(item).
		SetGame(game).
		WithAfter(small_cover_v10.V10FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}
