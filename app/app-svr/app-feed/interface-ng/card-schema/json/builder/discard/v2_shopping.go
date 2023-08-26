package discard

import (
	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	showmodel "go-gateway/app/app-svr/app-card/interface/model/card/show"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	"github.com/pkg/errors"
)

type V2ShopBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V2ShopBuilder
	SetBase(*jsoncard.Base) V2ShopBuilder
	SetShopping(*showmodel.Shopping) V2ShopBuilder

	Build() (*jsoncard.SmallCoverV2, error)
}

type v2ShopBuilder struct {
	jsonbuilder.BuilderContext
	base     *jsoncard.Base
	shopping *showmodel.Shopping
}

func NewV2ShopBuilder(ctx jsonbuilder.BuilderContext) V2ShopBuilder {
	return v2ShopBuilder{BuilderContext: ctx}
}

func (b v2ShopBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V2ShopBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v2ShopBuilder) SetBase(base *jsoncard.Base) V2ShopBuilder {
	b.base = base
	return b
}

func (b v2ShopBuilder) SetShopping(in *showmodel.Shopping) V2ShopBuilder {
	b.shopping = in
	return b
}

func (b v2ShopBuilder) constructURI() string {
	device := b.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoWeb, device.Plat(), int(device.Build()), b.shopping.URL, nil)
}

func (b v2ShopBuilder) constructArgs() jsoncard.Args {
	return jsoncard.Args{
		Type: b.shopping.Type,
	}
}

func (b v2ShopBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.shopping == nil {
		return nil, errors.Errorf("empty `shopping` field")
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(appcardmodel.ShoppingCover(b.shopping.PerformanceImage)).
		UpdateTitle(b.shopping.Name).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		Update(); err != nil {
		return nil, err
	}

	out := &jsoncard.SmallCoverV2{}
	//nolint:gomnd
	switch b.shopping.Type {
	case 1:
		out.CoverLeftText1 = appcardmodel.ShoppingDuration(b.shopping.STime, b.shopping.ETime)
		out.CoverRightText = b.shopping.CityName
		out.CoverRightIcon = appcardmodel.IconLocation
		if len(b.shopping.Tags) > 0 && b.shopping.Tags[0] != nil {
			out.Desc = b.shopping.Tags[0].TagName
		}
	case 2:
		out.CoverLeftText1 = b.shopping.Want
		out.Desc = b.shopping.Subname
	default:
		log.Warn("Unrecognized shopping type: %d", b.shopping.Type)
	}
	out.Badge = "会员购"
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "会员购")

	return out, nil
}
