package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonselect "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/select"

	"github.com/pkg/errors"
)

func BuildSelectFromFollowMode(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.FollowMode == nil {
		return nil, errors.Errorf("empty `FollowMode` filed")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.Select).
		SetGoto(appcardmodel.Gt(item.Goto)).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(0).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := jsonselect.NewSelectBuilder(ctx)
	card, err := builder.SetBase(base).SetFollowMode(fanoutResult.FollowMode).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
