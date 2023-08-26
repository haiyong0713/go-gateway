package jsonsmallcover

import (
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	"github.com/pkg/errors"
)

type V2EpPGCBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2EpPGCBuilder
	SetRcmd(*ai.Item) V2EpPGCBuilder
	SetEpisode(*pgccard.EpisodeCard) V2EpPGCBuilder
	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2EpPGCBuilder
}

type v2EpPGCBuilder struct {
	threePoint   jsoncommon.ThreePoint
	seasonCommon jsoncommon.BangumiSeason
	ogvCommon    jsoncommon.OgvEpisode
	parent       *smallCoverV2BuilderFactory
	base         *jsoncard.Base
	rcmd         *ai.Item
	episode      *pgccard.EpisodeCard
	afterFn      []func(*jsoncard.SmallCoverV2)
}

func (b v2EpPGCBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2EpPGCBuilder) SetBase(base *jsoncard.Base) V2EpPGCBuilder {
	b.base = base
	return b
}

func (b v2EpPGCBuilder) SetRcmd(rcmd *ai.Item) V2EpPGCBuilder {
	b.rcmd = rcmd
	return b
}

func (b v2EpPGCBuilder) SetEpisode(in *pgccard.EpisodeCard) V2EpPGCBuilder {
	b.episode = in
	return b
}

func (b v2EpPGCBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.episode == nil {
		return nil, errors.Errorf("empty `episode` field")
	}
	if b.episode.Season == nil {
		return nil, errors.Errorf("empty `episode.Season` field")
	}
	output := &jsoncard.SmallCoverV2{}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateURI(b.ogvCommon.ConstructOgvURI(b.episode, b.parent.BuilderContext.Device(), b.rcmd)).
		UpdateCover(b.episode.Cover).
		UpdateTitle(b.episode.TianmaSmallCardMeta.Title).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}

	output.RcmdReasonStyle, output.DescButton = b.constructReasonStyleOrDescButton()
	if epCover := b.ogvCommon.ConstructEpCover(b.episode); epCover != nil {
		output.CoverLeftText1 = epCover.CoverLeftText1
		output.CoverLeftIcon1 = epCover.CoverLeftIcon1
		output.CoverLeft1ContentDescription = epCover.CoverLeft1ContentDescription
		output.CoverLeftText2 = epCover.CoverLeftText2
		output.CoverLeftIcon2 = epCover.CoverLeftIcon2
		output.CoverLeft2ContentDescription = epCover.CoverLeft2ContentDescription
	}
	output.CoverRightText = b.ogvCommon.ConstructOgvRightText(b.episode, b.rcmd.OgvHasScore(),
		b.parent.BuilderContext.VersionControl().Can("feed.pgcScore"))
	output.Badge = b.episode.TianmaSmallCardMeta.BadgeInfo.Text
	output.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, b.episode.TianmaSmallCardMeta.BadgeInfo.Text)

	output.Base = b.base
	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b v2EpPGCBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
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

func (b v2EpPGCBuilder) enableWatched() bool {
	return b.rcmd.OgvDislikeInfo == ai.OgvWatched
}

func (b v2EpPGCBuilder) constructReasonStyleOrDescButton() (*jsoncard.ReasonStyle, *jsoncard.Button) {
	if b.rcmd.RcmdReason != nil {
		reasonStyle := jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
		)
		return reasonStyle, nil
	}
	return nil, b.seasonCommon.ConstructDescButtonFromNewEpShow(b.episode.TianmaSmallCardMeta.RcmdReason)
}

func (b v2EpPGCBuilder) enableFeedback() bool {
	return b.parent.BuilderContext.VersionControl().Can("feed.enableOGVFeedback") &&
		b.rcmd.OgvDislikeInfo >= 1 &&
		appcardmodel.Columnm[appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column())] == appcardmodel.ColumnSvrDouble
}

func (b v2EpPGCBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2EpPGCBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
