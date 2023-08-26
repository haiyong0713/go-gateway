package jsonsmallcover

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-feed/interface/common"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

//go:generate python3 ../../../../contrib/desc-button-overlapped.py

var (
	convergeCardGotoSet = sets.NewString(appfeedmodel.GotoAvConverge, appfeedmodel.GotoMultilayerConverge)
)

type V2ArcPlayerBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2ArcPlayerBuilder
	SetRcmd(*ai.Item) V2ArcPlayerBuilder
	SetArcPlayer(player *arcgrpc.ArcPlayer) V2ArcPlayerBuilder
	SetChannelCard(*channelgrpc.ChannelCard) V2ArcPlayerBuilder
	SetTag(*taggrpc.Tag) V2ArcPlayerBuilder
	SetAuthorCard(*accountgrpc.Card) V2ArcPlayerBuilder
	SetCoverGif(string) V2ArcPlayerBuilder
	SetStoryIcon(map[int64]*appcardmodel.GotoIcon) V2ArcPlayerBuilder
	SetOpenCourseMark(in map[int64]bool) V2ArcPlayerBuilder

	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2ArcPlayerBuilder
}

type v2ArchiveBuilder struct {
	archvieCommon  jsoncommon.ArchiveCommon
	threePoint     jsoncommon.ThreePoint
	base           *jsoncard.Base
	rcmd           *ai.Item
	parent         *smallCoverV2BuilderFactory
	arcPlayer      *arcgrpc.ArcPlayer
	channelCard    *channelgrpc.ChannelCard
	tag            *taggrpc.Tag
	authorCard     *accountgrpc.Card
	coverGif       string
	afterFn        []func(*jsoncard.SmallCoverV2)
	storyIcon      map[int64]*appcardmodel.GotoIcon
	openCourseMark map[int64]bool

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.SmallCoverV2
}

func (b v2ArchiveBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2ArchiveBuilder) SetBase(base *jsoncard.Base) V2ArcPlayerBuilder {
	b.base = base
	return b
}

func (b v2ArchiveBuilder) SetRcmd(item *ai.Item) V2ArcPlayerBuilder {
	b.rcmd = item
	return b
}

func (b v2ArchiveBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) V2ArcPlayerBuilder {
	b.arcPlayer = in
	return b
}

func (b v2ArchiveBuilder) SetChannelCard(in *channelgrpc.ChannelCard) V2ArcPlayerBuilder {
	b.channelCard = in
	return b
}

func (b v2ArchiveBuilder) SetTag(in *taggrpc.Tag) V2ArcPlayerBuilder {
	b.tag = in
	return b
}

func (b v2ArchiveBuilder) SetCoverGif(in string) V2ArcPlayerBuilder {
	b.coverGif = in
	return b
}

func (b v2ArchiveBuilder) SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) V2ArcPlayerBuilder {
	b.storyIcon = in
	return b
}

func (b v2ArchiveBuilder) SetOpenCourseMark(in map[int64]bool) V2ArcPlayerBuilder {
	b.openCourseMark = in
	return b
}

func (b v2ArchiveBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

func (b v2ArchiveBuilder) constructPGCWithRedirectURI(redirectURL string) string {
	return b.archvieCommon.ConstructPGCRedirectURI(redirectURL, appcardmodel.PGCTrackIDHandler(b.rcmd))
}

func (b v2ArchiveBuilder) constructVerticalArchiveURI() string {
	device := b.parent.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b v2ArchiveBuilder) resolvePGCRedirectURL() (string, bool) {
	if b.rcmd.Goto != appfeedmodel.GotoAv {
		return "", false
	}
	if !b.isPGCArchive() {
		return "", false
	}
	if b.arcPlayer.Arc.RedirectURL == "" {
		return "", false
	}
	return b.arcPlayer.Arc.RedirectURL, true
}

func (b v2ArchiveBuilder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

func (b v2ArchiveBuilder) resolveOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v2ArchiveBuilder) isPGCArchive() bool {
	return b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes
}

func (b v2ArchiveBuilder) SetAuthorCard(in *accountgrpc.Card) V2ArcPlayerBuilder {
	b.authorCard = in
	return b
}

func (b v2ArchiveBuilder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in small cover v2: %+v", b.arcPlayer)
	}
	return nil
}

func (b v2ArchiveBuilder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

func (b v2ArchiveBuilder) isConvergeCard() bool {
	return convergeCardGotoSet.Has(b.rcmd.Goto)
}

func (b v2ArchiveBuilder) isShowV2ReasonStyle() bool {
	if b.base.CardGoto != appcardmodel.CardGotoAvConverge {
		return false
	}
	return appcard.IsShowRcmdReasonStyleV2(b.rcmd)
}

func (b *v2ArchiveBuilder) settingVerticalArchive() error {
	if !b.jumpGotoVerticalAv() {
		return errors.Errorf("not a vertical archive: %+v", b.arcPlayer)
	}
	if !b.parent.BuilderContext.VersionControl().Can("archive.storyPlayerSupported") {
		return errors.Errorf("ignore story archvie to unsupported device: %+v", b.arcPlayer)
	}
	if (b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo && b.arcPlayer.Arc.Rights.Autoplay != 1) ||
		(b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes &&
			b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitBadgepay) == arcgrpc.AttrYes) {
		log.Warn("story cannot support archive: %+v", b.arcPlayer.GetArc())
		return nil
	}
	b.output.FfCover = common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed)
	b.output.GotoIcon = b.constructGotoIcon(b.rcmd.IconType)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(appcardmodel.GotoVerticalAv).
		UpdateURI(b.constructVerticalArchiveURI())
	return nil
}

func (b v2ArchiveBuilder) constructGotoIcon(iconType int) *appcardmodel.GotoIcon {
	return appcardmodel.FillGotoIcon(iconType, b.storyIcon)
}

func (b *v2ArchiveBuilder) settingPGCArchive() error {
	if !b.isPGCArchive() {
		return errors.Errorf("not a PGC archive: %+v", b.arcPlayer)
	}

	redirectURL, ok := b.resolvePGCRedirectURL()
	if ok {
		b.baseUpdater = b.baseUpdater.
			UpdateURI(b.constructPGCWithRedirectURI(redirectURL))
	}
	b.baseUpdater = b.baseUpdater.UpdatePlayerArgs(nil)
	b.output.CanPlay = 0
	return nil
}

func (b *v2ArchiveBuilder) settingV2ReasonStyle(reasonText string) error {
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

func (b v2ArchiveBuilder) constructDescButton() (*jsoncard.Button, error) {
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

func (b v2ArchiveBuilder) constructArgs() jsoncard.Args {
	return b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
}

func (b *v2ArchiveBuilder) settingThreePointOnArchive() error {
	args := b.constructArgs()
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	b.baseUpdater = b.baseUpdater.UpdateThreePoint(b.constructThreePoint(&args)).
		UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.parent.BuilderContext, &args,
			jsoncommon.WatchLater(true),
			jsoncommon.SwitchColumn(enableSwitchColumn),
			jsoncommon.AvDislikeInfo(b.rcmd.AvDislikeInfo),
			jsoncommon.Item(b.rcmd)))
	return nil
}

func (b *v2ArchiveBuilder) constructThreePoint(args *jsoncard.Args) *jsoncard.ThreePoint {
	out := b.threePoint.ConstructArchvieThreePoint(args, b.rcmd.AvDislikeInfo)
	appcard.ReplaceStoryDislikeReason(out.DislikeReasons, b.rcmd)
	return out
}

func (b *v2ArchiveBuilder) settingThreePoint() error {
	return b.settingThreePointOnArchive()
}

func (b *v2ArchiveBuilder) settingRecommendReason() error {
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
func (b *v2ArchiveBuilder) settingCover() error {
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

func (b *v2ArchiveBuilder) settingAsAvCoverge() error {
	if !b.isConvergeCard() {
		return errors.Errorf("not a coverge type: %q", b.base.CardGoto)
	}
	rcmd := b.rcmd
	if rcmd.ConvergeInfo == nil {
		return errors.Errorf("empty `coverage_info` in recommend")
	}
	if len(rcmd.ConvergeInfo.Items) <= 0 {
		return errors.Errorf("empty `items` in coverage_info")
	}

	uriParam, resolvedGoto := func() (string, appcardmodel.Gt) {
		if (rcmd.JumpID == 0 && rcmd.JumpGoto != string(appcardmodel.GotoHotPage)) || rcmd.Goto == "" {
			return strconv.FormatInt(rcmd.ID, 10), appcardmodel.GotoAvConverge
		}
		return strconv.FormatInt(rcmd.JumpID, 10), appcardmodel.Gt(rcmd.JumpGoto)
	}()
	if convergeCardGotoSet.Has(string(resolvedGoto)) {
		resolvedGoto = appcardmodel.GotoAvConverge
	}
	device := b.parent.BuilderContext.Device()
	uriExtraFn := appcardmodel.TrackIDHandler(rcmd.TrackID, rcmd, 0, 0)
	resolvedURI := appcardmodel.FillURI(resolvedGoto, device.Plat(), int(device.Build()), uriParam, uriExtraFn)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(resolvedGoto).
		UpdateURI(resolvedURI).
		UpdatePlayerArgs(nil)
	if rcmd.ID != 0 {
		b.baseUpdater = b.baseUpdater.UpdateParam(strconv.FormatInt(rcmd.ID, 10))
	}
	b.output.CanPlay = 0
	return nil
}

func (b v2ArchiveBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2ArcPlayerBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func (b v2ArchiveBuilder) Build() (*jsoncard.SmallCoverV2, error) {
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
		UpdatePlayerArgs(b.archvieCommon.ConstructPlayerArgs(b.arcPlayer))

	if err := b.ensureArchvieState(); err != nil {
		return nil, err
	}
	if b.jumpGotoVerticalAv() {
		if err := b.settingVerticalArchive(); err != nil {
			return nil, err
		}
	}
	if b.isPGCArchive() {
		if err := b.settingPGCArchive(); err != nil {
			return nil, err
		}
	}
	if b.isConvergeCard() {
		if err := b.settingAsAvCoverge(); err != nil {
			return nil, err
		}
	}
	if err := b.settingCover(); err != nil {
		return nil, err
	}
	if err := b.settingRecommendReason(); err != nil {
		return nil, err
	}
	if err := b.settingThreePoint(); err != nil {
		return nil, err
	}
	if err := b.baseUpdater.Update(); err != nil {
		return nil, err
	}
	b.output.Base = b.base
	if b.rcmd.IconType == appcardmodel.AIUpIconType && b.output.RcmdReason == "" &&
		b.output.CardGoto == appcardmodel.CardGotoAv {
		b.output.GotoIcon = b.constructGotoIcon(b.rcmd.IconType)
	}
	hasBadge, ok := b.openCourseMark[b.rcmd.ID]
	if ok && hasBadge {
		b.output.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "公开课")
		b.output.GotoIcon = nil
	}
	for _, fn := range b.afterFn {
		fn(b.output)
	}
	return b.output, nil
}
