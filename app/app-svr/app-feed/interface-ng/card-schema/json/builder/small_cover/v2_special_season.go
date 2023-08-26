package jsonsmallcover

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"

	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	"github.com/pkg/errors"
)

type V2SpecialSeasonBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2SpecialSeasonBuilder
	SetRcmd(*ai.Item) V2SpecialSeasonBuilder
	SetSeason(*pgcAppGrpc.SeasonCardInfoProto) V2SpecialSeasonBuilder
	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2SpecialSeasonBuilder
}

type v2SpecialSeasonBuilder struct {
	threePoint jsoncommon.ThreePoint
	parent     *smallCoverV2BuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	season     *pgcAppGrpc.SeasonCardInfoProto
	afterFn    []func(*jsoncard.SmallCoverV2)
}

func (b v2SpecialSeasonBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2SpecialSeasonBuilder) SetBase(base *jsoncard.Base) V2SpecialSeasonBuilder {
	b.base = base
	return b
}

func (b v2SpecialSeasonBuilder) SetRcmd(rcmd *ai.Item) V2SpecialSeasonBuilder {
	b.rcmd = rcmd
	return b
}

func (b v2SpecialSeasonBuilder) SetSeason(in *pgcAppGrpc.SeasonCardInfoProto) V2SpecialSeasonBuilder {
	b.season = in
	return b
}

func (b v2SpecialSeasonBuilder) constructSeasonURI(season *pgcAppGrpc.SeasonCardInfoProto, rcmd *ai.Item) string {
	param := season.Url
	return appcardmodel.FillURI("", 0, 0, param, appcardmodel.PGCTrackIDHandler(rcmd))
}

func (b v2SpecialSeasonBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.season == nil {
		return nil, errors.Errorf("empty `season` field")
	}
	output := &jsoncard.SmallCoverV2{}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateURI(b.constructSeasonURI(b.season, b.rcmd)).
		UpdateCover(b.season.Cover).
		UpdateTitle(b.season.Title).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)).
		Update(); err != nil {
		return nil, err
	}
	output.CoverLeftText1 = appcardmodel.StatString(int32(b.season.Stat.View), "")
	output.CoverLeftIcon1 = appcardmodel.IconPlay
	output.CoverLeftText2 = appcardmodel.StatString(int32(b.season.Stat.Follow), "")
	output.CoverLeftIcon2 = appcardmodel.IconFavorite
	output.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(output.CoverLeftIcon1, output.CoverLeftText1)
	output.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(output.CoverLeftIcon2, output.CoverLeftText2)
	output.Base = b.base
	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b v2SpecialSeasonBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2SpecialSeasonBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
