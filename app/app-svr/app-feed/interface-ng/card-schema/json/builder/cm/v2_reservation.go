package cm

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
)

type V2AdReservationBuilder interface {
	Parent() CmV2BuilderFactory
	SetBase(*jsoncard.Base) V2AdReservationBuilder
	SetAdInfo(*cm.AdInfo) V2AdReservationBuilder
	SetReservation(*activitygrpc.UpActReserveRelationInfo) V2AdReservationBuilder
	Build() (*jsoncard.LargeCoverInline, error)
}

type v2AdReservationBuilder struct {
	parent      *cmV2BuilderFactory
	base        *jsoncard.Base
	adInfo      *cm.AdInfo
	threePoint  jsoncommon.ThreePoint
	reservation *activitygrpc.UpActReserveRelationInfo
}

func (b v2AdReservationBuilder) Parent() CmV2BuilderFactory {
	return b.parent
}

func (b v2AdReservationBuilder) SetBase(base *jsoncard.Base) V2AdReservationBuilder {
	b.base = base
	return b
}

func (b v2AdReservationBuilder) SetAdInfo(adInfo *cm.AdInfo) V2AdReservationBuilder {
	b.adInfo = adInfo
	return b
}

func (b v2AdReservationBuilder) SetReservation(in *activitygrpc.UpActReserveRelationInfo) V2AdReservationBuilder {
	b.reservation = in
	return b
}

func (b v2AdReservationBuilder) constructCmInfo() *jsoncard.CmInfo {
	return &jsoncard.CmInfo{
		HidePlayButton:    true,
		ReservationTime:   b.reservation.LivePlanStartTime,
		ReservationNum:    b.reservation.Total,
		ReservationStatus: b.reservation.IsFollow,
	}
}

func (b v2AdReservationBuilder) Build() (*jsoncard.LargeCoverInline, error) {
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateArgs(jsoncard.Args{}).
		UpdateAdInfo(b.adInfo).
		UpdateCmInfo(b.constructCmInfo()).
		UpdateThreePointV2(b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)).Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverInline{
		Base: b.base,
	}
	return out, nil
}
