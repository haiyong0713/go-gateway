package cm

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

type V1AdAvBuilder interface {
	Parent() CmV1BuilderFactory
	SetBase(*jsoncard.Base) V1AdAvBuilder
	SetRcmd(*ai.Item) V1AdAvBuilder
	SetArcPlayer(*arcgrpc.ArcPlayer) V1AdAvBuilder
	SetChannelCard(*channelgrpc.ChannelCard) V1AdAvBuilder
	SetTag(*taggrpc.Tag) V1AdAvBuilder
	SetAuthorCard(*accountgrpc.Card) V1AdAvBuilder
	SetCoverGif(string) V1AdAvBuilder
	SetStoryIcon(map[int64]*appcardmodel.GotoIcon) V1AdAvBuilder
	SetAdInfo(*cm.AdInfo) V1AdAvBuilder

	Build() (*jsoncard.LargeCoverV1, error)
}

type v1AdAvBuilder struct {
	parent        *cmV1BuilderFactory
	archvieCommon jsoncommon.ArchiveCommon
	threePoint    jsoncommon.ThreePoint
	base          *jsoncard.Base
	rcmd          *ai.Item
	adInfo        *cm.AdInfo
	arcPlayer     *arcgrpc.ArcPlayer
	channelCard   *channelgrpc.ChannelCard
	tag           *taggrpc.Tag
	authorCard    *accountgrpc.Card
	coverGif      string
	storyIcon     map[int64]*appcardmodel.GotoIcon

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.LargeCoverV1
}

func (b v1AdAvBuilder) Parent() CmV1BuilderFactory {
	return b.parent
}

func (b v1AdAvBuilder) SetBase(base *jsoncard.Base) V1AdAvBuilder {
	b.base = base
	return b
}

func (b v1AdAvBuilder) SetRcmd(item *ai.Item) V1AdAvBuilder {
	b.rcmd = item
	return b
}

func (b v1AdAvBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) V1AdAvBuilder {
	b.arcPlayer = in
	return b
}

func (b v1AdAvBuilder) SetChannelCard(in *channelgrpc.ChannelCard) V1AdAvBuilder {
	b.channelCard = in
	return b
}

func (b v1AdAvBuilder) SetTag(in *taggrpc.Tag) V1AdAvBuilder {
	b.tag = in
	return b
}

func (b v1AdAvBuilder) SetCoverGif(in string) V1AdAvBuilder {
	b.coverGif = in
	return b
}

func (b v1AdAvBuilder) SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) V1AdAvBuilder {
	b.storyIcon = in
	return b
}

func (b v1AdAvBuilder) SetAdInfo(adInfo *cm.AdInfo) V1AdAvBuilder {
	b.adInfo = adInfo
	return b
}

func (b v1AdAvBuilder) SetAuthorCard(in *accountgrpc.Card) V1AdAvBuilder {
	b.authorCard = in
	return b
}

func (b v1AdAvBuilder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in small cover v2: %+v", b.arcPlayer)
	}
	return nil
}

func (b v1AdAvBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, nil, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

func (b v1AdAvBuilder) constructArgs() jsoncard.Args {
	return b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
}

func (b v1AdAvBuilder) resolveOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

//nolint:unused
func (b v1AdAvBuilder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

//nolint:unused
func (b v1AdAvBuilder) constructGotoIcon(iconType int) *appcardmodel.GotoIcon {
	return appcardmodel.FillGotoIcon(iconType, b.storyIcon)
}

//nolint:unused
func (b v1AdAvBuilder) constructVerticalArchiveURI() string {
	device := b.parent.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

//nolint:unused
func (b v1AdAvBuilder) isPGCArchive() bool {
	return b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes
}

//nolint:unparam
func (b *v1AdAvBuilder) settingCover() error {
	b.output.CoverGif = b.coverGif
	if b.parent.BuilderContext.VersionControl().Can("feed.enablePadNewCover") {
		b.output.CoverLeftText1 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, "")
		b.output.CoverLeftIcon1 = appcardmodel.IconPlay
		b.output.CoverLeftText2 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.Danmaku, "")
		b.output.CoverLeftIcon2 = appcardmodel.IconDanmaku
		b.output.CoverRightText = appcardmodel.DurationString(b.arcPlayer.Arc.Duration)
		return nil
	}
	b.output.CoverLeftText1 = appcardmodel.DurationString(b.arcPlayer.Arc.Duration)
	b.output.CoverLeftText2 = appcardmodel.ArchiveViewString(b.arcPlayer.Arc.Stat.View)
	b.output.CoverLeftText3 = appcardmodel.DanmakuString(b.arcPlayer.Arc.Stat.Danmaku)
	if b.parent.BuilderContext.VersionControl().Can("archive.usingFeedIndexLike") {
		b.output.CoverLeftText2 = appcardmodel.LikeString(b.arcPlayer.Arc.Stat.Like)
		b.output.CoverLeftText3 = appcardmodel.ArchiveViewString(b.arcPlayer.Arc.Stat.View)
	}
	return nil
}

func (b v1AdAvBuilder) isShowV2ReasonStyle() bool {
	if b.base.CardGoto != appcardmodel.CardGotoAvConverge {
		return false
	}
	return appcard.IsShowRcmdReasonStyleV2(b.rcmd)
}

func (b *v1AdAvBuilder) settingV2ReasonStyle(reasonText string) error {
	if !b.isShowV2ReasonStyle() {
		return errors.Errorf("cannot set v2 reason style")
	}
	b.output.Desc = ""
	versionControl := b.parent.BuilderContext.VersionControl()
	if versionControl.Can("feed.usingNewRcmdReasonV2") {
		b.output.RcmdReasonStyleV2 = jsonreasonstyle.ConstructReasonStyleV4(
			b.parent.BuilderContext,
			reasonText,
			b.rcmd,
		)
	}
	if versionControl.Can("feed.usingNewRcmdReason") {
		b.output.RcmdReasonStyleV2 = jsonreasonstyle.ConstructReasonStyleV3(
			b.parent.BuilderContext,
			reasonText,
			b.rcmd,
		)
	}
	b.output.RcmdReasonStyleV2 = jsonreasonstyle.ConstructReasonStyleV2(
		b.parent.BuilderContext,
		reasonText,
		b.rcmd,
	)
	return nil
}

func (b *v1AdAvBuilder) constructDesc() string {
	return fmt.Sprintf("%s · %s", b.arcPlayer.Arc.Author.Name, appcardmodel.PubDataByRequestAt(b.arcPlayer.Arc.PubDate.Time(), b.rcmd.RequestAt()))
}

func (b *v1AdAvBuilder) settingRecommendReason() error {
	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid),
	)
	if b.isShowV2ReasonStyle() {
		return b.settingV2ReasonStyle(topRcmdReason)
	}
	b.output.TopRcmdReason = topRcmdReason
	b.output.BottomRcmdReason = bottomRcmdReason
	b.output.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		topRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	b.output.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		bottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	b.output.Desc = b.constructDesc()
	return nil
}

func (b v1AdAvBuilder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.arcPlayer.Arc.Author.Face,
			Goto:  appcardmodel.GotoMid,
			Param: strconv.FormatInt(b.arcPlayer.Arc.Author.Mid, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Warn("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b v1AdAvBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, false)
}

func (b v1AdAvBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	if b.parent.BuilderContext.VersionControl().Can("feed.adAvHasPlayerArgs") {
		return b.archvieCommon.ConstructPlayerArgs(b.arcPlayer)
	}
	return nil
}

func (b v1AdAvBuilder) Build() (*jsoncard.LargeCoverV1, error) {
	if b.arcPlayer == nil {
		return nil, errors.Errorf("empty `arcPlayer` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if appcard.CheckMidMaxInt32(b.arcPlayer.Arc.Author.Mid) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.arcPlayer.Arc.Author.Mid)
	}
	if err := b.ensureArchvieState(); err != nil {
		return nil, err
	}

	// initial building context
	b.output = &jsoncard.LargeCoverV1{
		OfficialIcon: b.resolveOfficialIcon(),
		CanPlay:      b.arcPlayer.Arc.Rights.Autoplay,
	}
	b.baseUpdater = jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.arcPlayer.Arc.Pic).
		UpdateTitle(b.arcPlayer.Arc.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.constructPlayerArgs()).
		UpdateAdInfo(b.adInfo).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2())

	if err := b.settingCover(); err != nil {
		return nil, err
	}
	if err := b.settingRecommendReason(); err != nil {
		return nil, err
	}
	b.output.Avatar = b.constructAvatar()

	if err := b.baseUpdater.Update(); err != nil {
		return nil, err
	}
	b.output.Base = b.base
	return b.output, nil
}
