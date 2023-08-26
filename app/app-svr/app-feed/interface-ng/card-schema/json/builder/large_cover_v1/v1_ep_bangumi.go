package large_cover_v1

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"github.com/pkg/errors"
)

type V1EpBangumiBuilder interface {
	Parent() LargeCoverV1BuilderFactory
	SetBase(*jsoncard.Base) V1EpBangumiBuilder
	SetRcmd(*ai.Item) V1EpBangumiBuilder
	SetBangumiSeason(*episodegrpc.EpisodeCardsProto) V1EpBangumiBuilder
	SetTag(*taggrpc.Tag) V1EpBangumiBuilder

	Build() (*jsoncard.LargeCoverV1, error)
	WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1EpBangumiBuilder
}

type v1EpBangumiBuilder struct {
	seasonCommon jsoncommon.BangumiSeason
	threePoint   jsoncommon.ThreePoint
	base         *jsoncard.Base
	rcmd         *ai.Item
	parent       *largeCoverV1BuilderFactory
	episode      *episodegrpc.EpisodeCardsProto
	tag          *taggrpc.Tag
	afterFn      []func(*jsoncard.LargeCoverV1)
}

func (b v1EpBangumiBuilder) Parent() LargeCoverV1BuilderFactory {
	return b.parent
}

func (b v1EpBangumiBuilder) SetBase(base *jsoncard.Base) V1EpBangumiBuilder {
	b.base = base
	return b
}

func (b v1EpBangumiBuilder) SetRcmd(item *ai.Item) V1EpBangumiBuilder {
	b.rcmd = item
	return b
}

func (b v1EpBangumiBuilder) SetBangumiSeason(in *episodegrpc.EpisodeCardsProto) V1EpBangumiBuilder {
	b.episode = in
	return b
}

func (b v1EpBangumiBuilder) SetTag(in *taggrpc.Tag) V1EpBangumiBuilder {
	b.tag = in
	return b
}

func (b v1EpBangumiBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, enableSwitchColumn)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, enableSwitchColumn)
}

func (b v1EpBangumiBuilder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.episode.Season.Cover,
			Type:  appcardmodel.AvatarSquare,
		}).Build()
	if err != nil {
		log.Warn("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b v1EpBangumiBuilder) constructDescButton() *jsoncard.Button {
	if b.tag == nil {
		return nil
	}
	return b.seasonCommon.ConstructDescButtonFromTag(b.tag)
}

func (b v1EpBangumiBuilder) Build() (*jsoncard.LargeCoverV1, error) {
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
	output := &jsoncard.LargeCoverV1{}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateURI(b.seasonCommon.ConstructEpURI(b.episode, b.parent.Device(), b.rcmd)).
		UpdateCover(b.episode.Cover).
		UpdateTitle(b.seasonCommon.ConstructEpTitle(b.episode)).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateParam(strconv.FormatInt(int64(b.episode.EpisodeId), 10)).
		UpdateBaseInnerDescButton(b.constructDescButton()).
		Update(); err != nil {
		return nil, err
	}

	b.resolveCoverMeta(output)
	output.Avatar = b.constructAvatar()
	output.CoverBadge = b.episode.Season.SeasonTypeName
	output.CoverBadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorRed, b.episode.Season.SeasonTypeName)

	output.Desc = b.episode.Season.NewEpShow
	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		false,
	)
	output.TopRcmdReason = topRcmdReason
	output.BottomRcmdReason = bottomRcmdReason
	output.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		topRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	output.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		bottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)

	output.Base = b.base
	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b v1EpBangumiBuilder) resolveCoverMeta(in *jsoncard.LargeCoverV1) {
	if b.parent.BuilderContext.VersionControl().Can("feed.enablePadNewCover") {
		if epCover := b.seasonCommon.ConstructEpCover(b.episode); epCover != nil {
			in.CoverLeftText1 = epCover.CoverLeftText1
			in.CoverLeftIcon1 = epCover.CoverLeftIcon1
			in.CoverLeftText2 = epCover.CoverLeftText2
			in.CoverLeftIcon2 = epCover.CoverLeftIcon2
			return
		}
	}
	if b.episode.Season.Stat != nil {
		in.CoverLeftText2 = appcardmodel.ArchiveViewString(int32(b.episode.Season.Stat.View))
		in.CoverLeftText3 = appcardmodel.BangumiFavString(int32(b.episode.Season.Stat.Follow), b.episode.Season.SeasonType)
	}
}

func (b v1EpBangumiBuilder) WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1EpBangumiBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
