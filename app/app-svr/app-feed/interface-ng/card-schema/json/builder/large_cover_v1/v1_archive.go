package large_cover_v1

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
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
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

var (
	convergeCardGotoSet = sets.NewString(appfeedmodel.GotoAvConverge, appfeedmodel.GotoMultilayerConverge)
)

type V1ArcPlayerBuilder interface {
	Parent() LargeCoverV1BuilderFactory
	SetBase(*jsoncard.Base) V1ArcPlayerBuilder
	SetRcmd(*ai.Item) V1ArcPlayerBuilder
	SetArcPlayer(*arcgrpc.ArcPlayer) V1ArcPlayerBuilder
	SetChannelCard(*channelgrpc.ChannelCard) V1ArcPlayerBuilder
	SetTag(*taggrpc.Tag) V1ArcPlayerBuilder
	SetAuthorCard(*accountgrpc.Card) V1ArcPlayerBuilder
	SetCoverGif(string) V1ArcPlayerBuilder
	SetStoryIcon(map[int64]*appcardmodel.GotoIcon) V1ArcPlayerBuilder

	Build() (*jsoncard.LargeCoverV1, error)
	WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1ArcPlayerBuilder
}

type v1ArchiveBuilder struct {
	archvieCommon jsoncommon.ArchiveCommon
	threePoint    jsoncommon.ThreePoint
	base          *jsoncard.Base
	rcmd          *ai.Item
	parent        *largeCoverV1BuilderFactory
	arcPlayer     *arcgrpc.ArcPlayer
	channelCard   *channelgrpc.ChannelCard
	tag           *taggrpc.Tag
	authorCard    *accountgrpc.Card
	coverGif      string
	storyIcon     map[int64]*appcardmodel.GotoIcon
	afterFn       []func(*jsoncard.LargeCoverV1)

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.LargeCoverV1
}

func (b v1ArchiveBuilder) Parent() LargeCoverV1BuilderFactory {
	return b.parent
}

func (b v1ArchiveBuilder) SetBase(base *jsoncard.Base) V1ArcPlayerBuilder {
	b.base = base
	return b
}

func (b v1ArchiveBuilder) SetRcmd(item *ai.Item) V1ArcPlayerBuilder {
	b.rcmd = item
	return b
}

func (b v1ArchiveBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) V1ArcPlayerBuilder {
	b.arcPlayer = in
	return b
}

func (b v1ArchiveBuilder) SetChannelCard(in *channelgrpc.ChannelCard) V1ArcPlayerBuilder {
	b.channelCard = in
	return b
}

func (b v1ArchiveBuilder) SetTag(in *taggrpc.Tag) V1ArcPlayerBuilder {
	b.tag = in
	return b
}

func (b v1ArchiveBuilder) SetCoverGif(in string) V1ArcPlayerBuilder {
	b.coverGif = in
	return b
}

func (b v1ArchiveBuilder) SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) V1ArcPlayerBuilder {
	b.storyIcon = in
	return b
}

func (b v1ArchiveBuilder) SetAuthorCard(in *accountgrpc.Card) V1ArcPlayerBuilder {
	b.authorCard = in
	return b
}

func (b v1ArchiveBuilder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in small cover v2: %+v", b.arcPlayer)
	}
	return nil
}

func (b v1ArchiveBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

func (b v1ArchiveBuilder) constructArgs() jsoncard.Args {
	return b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
}

func (b v1ArchiveBuilder) resolveOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v1ArchiveBuilder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

func (b v1ArchiveBuilder) constructGotoIcon(iconType int) *appcardmodel.GotoIcon {
	return appcardmodel.FillGotoIcon(iconType, b.storyIcon)
}

func (b v1ArchiveBuilder) constructVerticalArchiveURI() string {
	device := b.parent.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b *v1ArchiveBuilder) settingVerticalArchive() error {
	if !b.jumpGotoVerticalAv() {
		return errors.Errorf("not a vertical archive: %+v", b.arcPlayer)
	}
	if !b.parent.BuilderContext.VersionControl().Can("archive.storyPlayerSupported") {
		return errors.Errorf("ignore story archvie to unsupported device: %+v", b.arcPlayer)
	}
	b.output.FfCover = common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed)
	b.output.GotoIcon = b.constructGotoIcon(b.rcmd.IconType)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(appcardmodel.GotoVerticalAv).
		UpdateURI(b.constructVerticalArchiveURI())
	return nil
}

func (b v1ArchiveBuilder) isPGCArchive() bool {
	return b.arcPlayer.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes
}

func (b v1ArchiveBuilder) constructPGCWithRedirectURI(redirectURL string) string {
	return b.archvieCommon.ConstructPGCRedirectURI(redirectURL, appcardmodel.PGCTrackIDHandler(b.rcmd))
}

func (b v1ArchiveBuilder) resolvePGCRedirectURL() (string, bool) {
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

func (b *v1ArchiveBuilder) settingPGCArchive() error {
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

//nolint:unparam
func (b *v1ArchiveBuilder) settingCover() error {
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

func (b v1ArchiveBuilder) isConvergeCard() bool {
	return convergeCardGotoSet.Has(b.rcmd.Goto)
}

func (b *v1ArchiveBuilder) settingAsAvCoverge() error {
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

func (b v1ArchiveBuilder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

func (b v1ArchiveBuilder) isShowV2ReasonStyle() bool {
	if b.base.CardGoto != appcardmodel.CardGotoAvConverge {
		return false
	}
	return appcard.IsShowRcmdReasonStyleV2(b.rcmd)
}

func (b *v1ArchiveBuilder) settingV2ReasonStyle(reasonText string) error {
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

func (b *v1ArchiveBuilder) constructDesc() string {
	authorName := b.resolveAuthorName()
	if b.arcPlayer.Arc.Rights.IsCooperation == 1 &&
		b.parent.BuilderContext.VersionControl().Can("archvie.showCooperation") {
		authorName = fmt.Sprintf("%s 等联合创作", authorName)
	}
	if b.rcmd.RcmdReason != nil &&
		b.rcmd.RcmdReason.Style == 3 &&
		b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid) {
		return authorName
	}
	return fmt.Sprintf("%s · %s", authorName, appcardmodel.PubDataByRequestAt(b.arcPlayer.Arc.PubDate.Time(), b.rcmd.RequestAt()))
}

func (b *v1ArchiveBuilder) settingRecommendReason() error {
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
	button, err := b.constructDescButton()
	if err != nil {
		log.Error("Failed to construct desc button: %+v", err)
	}
	b.baseUpdater = b.baseUpdater.UpdateBaseInnerDescButton(button)
	return nil
}

func (b v1ArchiveBuilder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.resolveAvatarCover(),
			Goto:  appcardmodel.GotoMid,
			Param: strconv.FormatInt(b.arcPlayer.Arc.Author.Mid, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Warn("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b v1ArchiveBuilder) resolveAvatarCover() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Face
	}
	return b.authorCard.Face
}

func (b v1ArchiveBuilder) constructMask() *jsoncard.Mask {
	out := &jsoncard.Mask{}
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.arcPlayer.Arc.Author.Face,
			Text:  b.arcPlayer.Arc.Author.Name,
			Goto:  appcardmodel.GotoMid,
			Param: strconv.FormatInt(b.arcPlayer.Arc.Author.Mid, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Warn("Failed to build mask avatar: %+v", err)
	}
	out.Avatar = avatar
	out.Button = b.archvieCommon.ConstructDescButtonFromMid(
		b.parent.BuilderContext,
		b.arcPlayer.Arc.Author.Mid,
	)
	return out
}

//nolint:unparam
func (b v1ArchiveBuilder) constructDescButton() (*jsoncard.Button, error) {
	if b.channelCard != nil && b.channelCard.ChannelId != 0 && b.channelCard.ChannelName != "" {
		channelName := fmt.Sprintf("%s · %s", b.arcPlayer.Arc.TypeName, b.channelCard.ChannelName)
		return b.archvieCommon.ConstructDescButtonFromChannel(channelName, b.channelCard.ChannelId), nil
	}
	if b.tag != nil {
		return b.archvieCommon.ConstructDescButtonFromTag(b.tag), nil
	}
	return b.archvieCommon.ConstructDescButtonFromArchvieType(b.arcPlayer.Arc.TypeName), nil
}

func (b *v1ArchiveBuilder) settingThreePointOnArchive() error {
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

func (b *v1ArchiveBuilder) constructThreePoint(args *jsoncard.Args) *jsoncard.ThreePoint {
	out := b.threePoint.ConstructArchvieThreePoint(args, b.rcmd.AvDislikeInfo)
	appcard.ReplaceStoryDislikeReason(out.DislikeReasons, b.rcmd)
	return out
}

func (b *v1ArchiveBuilder) settingThreePoint() error {
	return b.settingThreePointOnArchive()
}

func (b v1ArchiveBuilder) Build() (*jsoncard.LargeCoverV1, error) {
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
	bvid, _ := appcard.GetBvIDStr(b.base.Param)
	b.baseUpdater = jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.arcPlayer.Arc.Pic).
		UpdateTitle(b.arcPlayer.Arc.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.archvieCommon.ConstructPlayerArgs(b.arcPlayer)).
		UpdateMask(b.constructMask()).
		UpdateBvid(bvid)

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
	if err := b.settingCover(); err != nil {
		return nil, err
	}
	if err := b.settingRecommendReason(); err != nil {
		return nil, err
	}
	if b.isConvergeCard() {
		if err := b.settingAsAvCoverge(); err != nil {
			return nil, err
		}
	}
	if err := b.settingThreePoint(); err != nil {
		return nil, err
	}
	b.output.Avatar = b.constructAvatar()

	if err := b.baseUpdater.Update(); err != nil {
		return nil, err
	}
	b.output.Base = b.base
	for _, fn := range b.afterFn {
		fn(b.output)
	}
	return b.output, nil
}

func (b v1ArchiveBuilder) WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1ArcPlayerBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
