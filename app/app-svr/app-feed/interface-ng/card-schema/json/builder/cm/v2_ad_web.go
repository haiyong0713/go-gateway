package cm

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
)

type V2AdWebBuilder interface {
	Parent() CmV2BuilderFactory
	SetBase(*jsoncard.Base) V2AdWebBuilder
	SetAdInfo(*cm.AdInfo) V2AdWebBuilder
	Build() (*jsoncard.MiddleCover, error)
}

type v2AdWebBuilder struct {
	parent     *cmV2BuilderFactory
	base       *jsoncard.Base
	adInfo     *cm.AdInfo
	threePoint jsoncommon.ThreePoint
}

func (b v2AdWebBuilder) Parent() CmV2BuilderFactory {
	return b.parent
}

func (b v2AdWebBuilder) SetBase(base *jsoncard.Base) V2AdWebBuilder {
	b.base = base
	return b
}

func (b v2AdWebBuilder) SetAdInfo(adInfo *cm.AdInfo) V2AdWebBuilder {
	b.adInfo = adInfo
	return b
}

func (b v2AdWebBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, false)
}

func (b v2AdWebBuilder) Build() (*jsoncard.MiddleCover, error) {
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateArgs(jsoncard.Args{}).
		UpdateAdInfo(b.adInfo).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.MiddleCover{
		Base: b.base,
	}
	return out, nil
}
