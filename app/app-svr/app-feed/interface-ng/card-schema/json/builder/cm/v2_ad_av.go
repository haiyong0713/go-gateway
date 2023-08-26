package cm

import (
	"fmt"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	"go-gateway/app/app-svr/app-feed/interface/common"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

type V2AdAvBuilder interface {
	Parent() CmV2BuilderFactory
	SetBase(*jsoncard.Base) V2AdAvBuilder
	SetRcmd(*ai.Item) V2AdAvBuilder
	SetAdInfo(*cm.AdInfo) V2AdAvBuilder
	SetArcPlayer(*arcgrpc.ArcPlayer) V2AdAvBuilder
	SetChannelCard(*channelgrpc.ChannelCard) V2AdAvBuilder
	SetTag(*taggrpc.Tag) V2AdAvBuilder
	SetAuthorCard(*accountgrpc.Card) V2AdAvBuilder
	SetCoverGif(string) V2AdAvBuilder
	Build() (*jsoncard.SmallCoverV2, error)
}

type v2AdAvBuilder struct {
	parent        *cmV2BuilderFactory
	archvieCommon jsoncommon.ArchiveCommon
	base          *jsoncard.Base
	rcmd          *ai.Item
	adInfo        *cm.AdInfo
	threePoint    jsoncommon.ThreePoint
	arcPlayer     *arcgrpc.ArcPlayer
	channelCard   *channelgrpc.ChannelCard
	tag           *taggrpc.Tag
	authorCard    *accountgrpc.Card
	coverGif      string

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.SmallCoverV2
}

func (b v2AdAvBuilder) Parent() CmV2BuilderFactory {
	return b.parent
}

func (b v2AdAvBuilder) SetBase(base *jsoncard.Base) V2AdAvBuilder {
	b.base = base
	return b
}

func (b v2AdAvBuilder) SetAdInfo(adInfo *cm.AdInfo) V2AdAvBuilder {
	b.adInfo = adInfo
	return b
}

func (b v2AdAvBuilder) SetRcmd(item *ai.Item) V2AdAvBuilder {
	b.rcmd = item
	return b
}

func (b v2AdAvBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) V2AdAvBuilder {
	b.arcPlayer = in
	return b
}

func (b v2AdAvBuilder) SetChannelCard(in *channelgrpc.ChannelCard) V2AdAvBuilder {
	b.channelCard = in
	return b
}

func (b v2AdAvBuilder) SetTag(in *taggrpc.Tag) V2AdAvBuilder {
	b.tag = in
	return b
}

func (b v2AdAvBuilder) SetCoverGif(in string) V2AdAvBuilder {
	b.coverGif = in
	return b
}

func (b v2AdAvBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, nil, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

//nolint:unused
func (b v2AdAvBuilder) constructPGCWithRedirectURI(redirectURL string) string {
	return b.archvieCommon.ConstructPGCRedirectURI(redirectURL, appcardmodel.PGCTrackIDHandler(b.rcmd))
}

func (b v2AdAvBuilder) constructVerticalArchiveURI() string {
	device := b.parent.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b v2AdAvBuilder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

func (b v2AdAvBuilder) resolveOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

//nolint:unused
func (b v2AdAvBuilder) isPGCArchive() bool {
	return b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes
}

func (b v2AdAvBuilder) SetAuthorCard(in *accountgrpc.Card) V2AdAvBuilder {
	b.authorCard = in
	return b
}

func (b v2AdAvBuilder) ensureArchvieState() error {
	if !appcardmodel.AdAvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in small cover v2: %+v", b.arcPlayer)
	}
	return nil
}

func (b v2AdAvBuilder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

func (b v2AdAvBuilder) isShowV2ReasonStyle() bool {
	if b.base.CardGoto != appcardmodel.CardGotoAvConverge {
		return false
	}
	return appcard.IsShowRcmdReasonStyleV2(b.rcmd)
}

func (b *v2AdAvBuilder) settingVerticalArchive() error {
	if !b.jumpGotoVerticalAv() {
		return errors.Errorf("not a vertical archive: %+v", b.arcPlayer)
	}
	if !b.parent.BuilderContext.VersionControl().Can("archive.storyPlayerSupported") {
		return errors.Errorf("ignore story archvie to unsupported device: %+v", b.arcPlayer)
	}
	b.output.FfCover = common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(appcardmodel.GotoVerticalAv).
		UpdateURI(b.constructVerticalArchiveURI())
	return nil
}

func (b *v2AdAvBuilder) settingV2ReasonStyle(reasonText string) error {
	if !b.isShowV2ReasonStyle() {
		return errors.Errorf("cannot set v2 reason style")
	}
	b.output.RcmdReason = ""
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

func (b v2AdAvBuilder) constructDescButton() (*jsoncard.Button, error) {
	if b.output.RcmdReason != "" {
		return nil, errors.Errorf("skip desc button consturction on `RcmdReason` is exist: %+v", b.output)
	}
	if b.channelCard != nil && b.channelCard.ChannelId != 0 && b.channelCard.ChannelName != "" {
		channelName := fmt.Sprintf("%s · %s", b.arcPlayer.Arc.TypeName, b.channelCard.ChannelName)
		return b.archvieCommon.ConstructDescButtonFromChannel(channelName, b.channelCard.ChannelId), nil
	}
	if b.tag != nil {
		tagDup := &taggrpc.Tag{}
		*tagDup = *b.tag
		tagDup.Name = fmt.Sprintf("%s · %s", b.arcPlayer.Arc.TypeName, tagDup.Name)
		return b.archvieCommon.ConstructDescButtonFromTag(tagDup), nil
	}
	return b.archvieCommon.ConstructDescButtonFromArchvieType(b.arcPlayer.Arc.TypeName), nil
}

func (b v2AdAvBuilder) constructArgs() jsoncard.Args {
	return b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
}

func (b *v2AdAvBuilder) settingRecommendReason() error {
	rcmdReason, desc := jsonreasonstyle.BuildRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		b.resolveAuthorName(),
		b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid),
	)
	if b.isShowV2ReasonStyle() {
		return b.settingV2ReasonStyle(rcmdReason)
	}
	b.output.RcmdReason = rcmdReason
	b.output.Desc = desc
	b.output.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		b.output.RcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	if b.output.RcmdReason != "" {
		return nil
	}
	button, err := b.constructDescButton()
	if err != nil {
		log.Error("Failed to construct desc button: %+v", err)
	}
	b.output.DescButton = button
	return nil
}

//nolint:unparam
func (b *v2AdAvBuilder) settingCover() error {
	b.output.CoverLeftText1 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, "")
	b.output.CoverLeftIcon1 = appcardmodel.IconPlay
	b.output.CoverLeftText2 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.Danmaku, "")
	b.output.CoverLeftIcon2 = appcardmodel.IconDanmaku
	if b.parent.BuilderContext.VersionControl().Can("archive.usingFeedIndexLike") {
		b.output.CoverLeftText1 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.Like, "")
		b.output.CoverLeftIcon1 = appcardmodel.IconLike
		b.output.CoverLeftText2 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, "")
		b.output.CoverLeftIcon2 = appcardmodel.IconPlay
	}
	b.output.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(b.output.CoverLeftIcon1,
		b.output.CoverLeftText1)
	b.output.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(b.output.CoverLeftIcon2,
		b.output.CoverLeftText2)
	b.output.CoverRightText = appcardmodel.DurationString(b.arcPlayer.Arc.Duration)
	b.output.CoverRightContentDescription = appcardmodel.DurationContentDescription(b.arcPlayer.Arc.Duration)
	b.output.CoverGif = b.coverGif
	return nil
}

func (b v2AdAvBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, false)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, false)
}

func (b v2AdAvBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.arcPlayer == nil {
		return nil, errors.Errorf("empty `arcPlayer` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if appcard.CheckMidMaxInt32(b.arcPlayer.Arc.Author.Mid) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.arcPlayer.Arc.Author.Mid)
	}

	// initial building context
	b.output = &jsoncard.SmallCoverV2{
		OfficialIcon: b.resolveOfficialIcon(),
		CanPlay:      b.arcPlayer.Arc.Rights.Autoplay,
	}
	b.baseUpdater = jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.arcPlayer.Arc.Pic).
		UpdateTitle(b.arcPlayer.Arc.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.archvieCommon.ConstructPlayerArgs(b.arcPlayer)).
		UpdateAdInfo(b.adInfo).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2())

	if err := b.ensureArchvieState(); err != nil {
		return nil, err
	}
	if b.jumpGotoVerticalAv() {
		if err := b.settingVerticalArchive(); err != nil {
			return nil, err
		}
	}
	if err := b.settingCover(); err != nil {
		return nil, err
	}
	if err := b.settingRecommendReason(); err != nil {
		return nil, err
	}

	if err := b.baseUpdater.Update(); err != nil {
		return nil, err
	}
	b.output.Base = b.base

	return b.output, nil
}
