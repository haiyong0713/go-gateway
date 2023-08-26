package jsonsmallcover

import (
	"github.com/pkg/errors"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
)

type V2WebBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2WebBuilder
	SetRcmd(*ai.Item) V2WebBuilder
	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2WebBuilder
}

type v2WebBuilder struct {
	threePoint jsoncommon.ThreePoint
	parent     *smallCoverV2BuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	afterFn    []func(*jsoncard.SmallCoverV2)
}

func (b v2WebBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2WebBuilder) SetBase(base *jsoncard.Base) V2WebBuilder {
	b.base = base
	return b
}

func (b v2WebBuilder) SetRcmd(rcmd *ai.Item) V2WebBuilder {
	b.rcmd = rcmd
	return b
}

func (b v2WebBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	output := &jsoncard.SmallCoverV2{}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	output.Base = b.base
	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b v2WebBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	enableFeedback := b.enableFeedback()
	enableWatched := b.enableWatched()
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructOGVThreePointV2(b.parent.BuilderContext, enableSwitchColumn, enableFeedback,
			enableWatched)
	}
	return b.threePoint.ConstructOGVThreePointV2Legacy(b.parent.BuilderContext, enableSwitchColumn, enableFeedback,
		enableWatched)
}

func (b v2WebBuilder) enableWatched() bool {
	return b.rcmd.OgvDislikeInfo == ai.OgvWatched
}

func (b v2WebBuilder) enableFeedback() bool {
	return b.parent.BuilderContext.VersionControl().Can("feed.enableOGVFeedback") &&
		b.rcmd.OgvDislikeInfo >= 1 &&
		appcardmodel.Columnm[appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column())] == appcardmodel.ColumnSvrDouble
}

func (b v2WebBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2WebBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
