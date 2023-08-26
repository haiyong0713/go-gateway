package cm

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
)

type V1AdWebBuilder interface {
	Parent() CmV1BuilderFactory
	SetBase(*jsoncard.Base) V1AdWebBuilder
	SetAdInfo(*cm.AdInfo) V1AdWebBuilder
	Build() (*jsoncard.SmallCoverV1, error)
}

type v1AdWebBuilder struct {
	parent     *cmV1BuilderFactory
	base       *jsoncard.Base
	adInfo     *cm.AdInfo
	threePoint jsoncommon.ThreePoint
}

func (b v1AdWebBuilder) Parent() CmV1BuilderFactory {
	return b.parent
}

func (b v1AdWebBuilder) SetBase(base *jsoncard.Base) V1AdWebBuilder {
	b.base = base
	return b
}

func (b v1AdWebBuilder) SetAdInfo(adInfo *cm.AdInfo) V1AdWebBuilder {
	b.adInfo = adInfo
	return b
}

func (b v1AdWebBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, false)
}

func (b v1AdWebBuilder) Build() (*jsoncard.SmallCoverV1, error) {
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateArgs(jsoncard.Args{}).
		UpdateAdInfo(b.adInfo).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV1{
		Base: b.base,
	}
	return out, nil
}
