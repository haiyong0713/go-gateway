package cm

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
)

type V1AdPlayerBuilder interface {
	Parent() CmV1BuilderFactory
	SetBase(*jsoncard.Base) V1AdPlayerBuilder
	SetAdInfo(*cm.AdInfo) V1AdPlayerBuilder
	Build() (*jsoncard.LargeCoverV1, error)
}

type v1AdPlayerBuilder struct {
	parent     *cmV1BuilderFactory
	base       *jsoncard.Base
	adInfo     *cm.AdInfo
	threePoint jsoncommon.ThreePoint
}

func (b v1AdPlayerBuilder) Parent() CmV1BuilderFactory {
	return b.parent
}

func (b v1AdPlayerBuilder) SetBase(base *jsoncard.Base) V1AdPlayerBuilder {
	b.base = base
	return b
}

func (b v1AdPlayerBuilder) SetAdInfo(adInfo *cm.AdInfo) V1AdPlayerBuilder {
	b.adInfo = adInfo
	return b
}

func (b v1AdPlayerBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, false)
}

func (b v1AdPlayerBuilder) constructCmInfo() *jsoncard.CmInfo {
	return &jsoncard.CmInfo{
		HidePlayButton: true,
	}
}

func (b v1AdPlayerBuilder) Build() (*jsoncard.LargeCoverV1, error) {
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateArgs(jsoncard.Args{}).
		UpdateCmInfo(b.constructCmInfo()).
		UpdateAdInfo(b.adInfo).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverV1{
		Base: b.base,
	}
	return out, nil
}
