package jsonsmallcover

import (
	"fmt"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"github.com/pkg/errors"
)

type V2EpBangumiBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2EpBangumiBuilder
	SetRcmd(*ai.Item) V2EpBangumiBuilder
	SetEpisode(*episodegrpc.EpisodeCardsProto) V2EpBangumiBuilder
	SetTag(*taggrpc.Tag) V2EpBangumiBuilder
	SetArchive(player *arcgrpc.ArcPlayer) V2EpBangumiBuilder
	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2EpBangumiBuilder
}

type v2EpBangumiBuilder struct {
	threePoint   jsoncommon.ThreePoint
	seasonCommon jsoncommon.BangumiSeason
	parent       *smallCoverV2BuilderFactory
	base         *jsoncard.Base
	rcmd         *ai.Item
	episode      *episodegrpc.EpisodeCardsProto
	tag          *taggrpc.Tag
	archive      *arcgrpc.ArcPlayer
	afterFn      []func(*jsoncard.SmallCoverV2)
}

func (b v2EpBangumiBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2EpBangumiBuilder) SetBase(base *jsoncard.Base) V2EpBangumiBuilder {
	b.base = base
	return b
}

func (b v2EpBangumiBuilder) SetRcmd(rcmd *ai.Item) V2EpBangumiBuilder {
	b.rcmd = rcmd
	return b
}

func (b v2EpBangumiBuilder) SetEpisode(in *episodegrpc.EpisodeCardsProto) V2EpBangumiBuilder {
	b.episode = in
	return b
}

func (b v2EpBangumiBuilder) SetTag(tag *taggrpc.Tag) V2EpBangumiBuilder {
	b.tag = tag
	return b
}

func (b v2EpBangumiBuilder) SetArchive(archive *arcgrpc.ArcPlayer) V2EpBangumiBuilder {
	b.archive = archive
	return b
}

func (b v2EpBangumiBuilder) Build() (*jsoncard.SmallCoverV2, error) {
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
		UpdateURI(b.seasonCommon.ConstructEpURI(b.episode, b.parent.Device(), b.rcmd)).
		UpdateCover(b.episode.Cover).
		UpdateTitle(b.seasonCommon.ConstructEpTitle(b.episode)).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateParam(strconv.FormatInt(int64(b.episode.EpisodeId), 10)).
		Update(); err != nil {
		return nil, err
	}

	output.RcmdReasonStyle, output.DescButton = b.constructEXPReasonStyleOrDescButton()
	if epCover := b.seasonCommon.ConstructEpCover(b.episode); epCover != nil {
		output.CoverLeftText1 = epCover.CoverLeftText1
		output.CoverLeftIcon1 = epCover.CoverLeftIcon1
		output.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(output.CoverLeftIcon1,
			output.CoverLeftText1)
		output.CoverLeftText2 = epCover.CoverLeftText2
		output.CoverLeftIcon2 = epCover.CoverLeftIcon2
		output.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(output.CoverLeftIcon2,
			output.CoverLeftText2)
	}
	output.Badge, output.BadgeStyle = b.seasonCommon.ConstructEpBadge(b.episode)
	output.Base = b.base

	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b v2EpBangumiBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
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

func (b v2EpBangumiBuilder) enableWatched() bool {
	return b.rcmd.OgvDislikeInfo == ai.OgvWatched
}

func (b v2EpBangumiBuilder) constructDescButtonFromTag() (*jsoncard.Button, bool) {
	if b.tag == nil || b.tag.Name == "" {
		return nil, false
	}
	if b.archive == nil || b.archive.Arc.TypeName == "" {
		return nil, false
	}
	tag := &taggrpc.Tag{}
	*tag = *b.tag
	tag.Name = fmt.Sprintf("%s · %s", b.archive.Arc.TypeName, tag.Name)
	return jsoncommon.ConstructDescButtonFromTag(tag), true
}

// 实验逻辑，实验完成后下线
// nolint:gomnd
func (b v2EpBangumiBuilder) constructEXPReasonStyleOrDescButton() (*jsoncard.ReasonStyle, *jsoncard.Button) {
	if b.rcmd.RcmdReason != nil && b.rcmd.CustomizedOGVDesc != "" {
		reasonStyle := jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
		)
		desc := b.seasonCommon.ConstructDescButtonFromNewEpShow(b.rcmd.CustomizedOGVDesc)
		switch b.rcmd.OGVDescPriority {
		case 1:
			return nil, desc
		case 2:
			return reasonStyle, nil
		case 3:
			if b.rcmd.RcmdReason.Content == "继续观看" || b.rcmd.RcmdReason.Content == "繼續觀看" {
				return reasonStyle, nil
			}
			return nil, desc
		case 4:
			if b.rcmd.RcmdReason.Content != "继续观看" && b.rcmd.RcmdReason.Content != "繼續觀看" {
				return reasonStyle, nil
			}
			return nil, desc
		}
	}
	return b.constructReasonStyleOrDescButton()
}

func (b v2EpBangumiBuilder) constructReasonStyleOrDescButton() (*jsoncard.ReasonStyle, *jsoncard.Button) {
	if b.rcmd.RcmdReason != nil {
		reasonStyle := jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
		)
		return reasonStyle, nil
	}

	descButton, ok := b.constructDescButtonFromTag()
	if ok {
		return nil, descButton
	}

	return nil, b.seasonCommon.ConstructDescButtonFromNewEpShow(b.episode.Season.NewEpShow)
}

func (b v2EpBangumiBuilder) enableFeedback() bool {
	return b.parent.BuilderContext.VersionControl().Can("feed.enableOGVFeedback") &&
		b.rcmd.OgvDislikeInfo >= 1 &&
		appcardmodel.Columnm[appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column())] == appcardmodel.ColumnSvrDouble
}

func (b v2EpBangumiBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2EpBangumiBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
