package large_cover

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
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
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"github.com/pkg/errors"
)

type InlineAvV2Builder interface {
	Parent() LargeCoverInlineBuilderFactory

	SetBase(*jsoncard.Base) InlineAvV2Builder
	SetRcmd(*ai.Item) InlineAvV2Builder
	SetArcPlayer(*arcgrpc.ArcPlayer) InlineAvV2Builder
	SetTag(*taggrpc.Tag) InlineAvV2Builder
	SetAuthorCard(*accountgrpc.Card) InlineAvV2Builder
	SetHasLike(map[int64]int8) InlineAvV2Builder
	SetInline(*Inline) InlineAvV2Builder
	SetStoryIcon(map[int64]*appcardmodel.GotoIcon) InlineAvV2Builder
	SetHasFav(map[int64]int8) InlineAvV2Builder
	SetHotAidSet(sets.Int64) InlineAvV2Builder
	SetHasCoin(map[int64]int64) InlineAvV2Builder
	SetLikeStatState(map[int64]*thumbupgrpc.StatState) InlineAvV2Builder

	Build() (*jsoncard.LargeCoverInline, error)
	WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineAvV2Builder
}

type v9InlineAvBuilder struct {
	archvieCommon jsoncommon.ArchiveCommon
	threePoint    jsoncommon.ThreePoint
	base          *jsoncard.Base
	parent        *largeCoverInlineBuilderFactory
	rcmd          *ai.Item
	arcPlayer     *arcgrpc.ArcPlayer
	tag           *taggrpc.Tag
	authorCard    *accountgrpc.Card
	hasLike       map[int64]int8
	inline        *Inline
	storyIcon     map[int64]*appcardmodel.GotoIcon
	hasFav        map[int64]int8
	hotAidSet     sets.Int64
	hasCoin       map[int64]int64
	likeStatState map[int64]*thumbupgrpc.StatState

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.LargeCoverInline
	afterFn     []func(*jsoncard.LargeCoverInline)
}

func (b v9InlineAvBuilder) Parent() LargeCoverInlineBuilderFactory {
	return b.parent
}

func (b v9InlineAvBuilder) SetBase(base *jsoncard.Base) InlineAvV2Builder {
	b.base = base
	return b
}

func (b v9InlineAvBuilder) SetRcmd(item *ai.Item) InlineAvV2Builder {
	b.rcmd = item
	return b
}

func (b v9InlineAvBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) InlineAvV2Builder {
	b.arcPlayer = in
	return b
}

func (b v9InlineAvBuilder) SetTag(in *taggrpc.Tag) InlineAvV2Builder {
	b.tag = in
	return b
}

func (b v9InlineAvBuilder) SetAuthorCard(in *accountgrpc.Card) InlineAvV2Builder {
	b.authorCard = in
	return b
}

func (b v9InlineAvBuilder) SetHasLike(in map[int64]int8) InlineAvV2Builder {
	b.hasLike = in
	return b
}

func (b v9InlineAvBuilder) SetInline(in *Inline) InlineAvV2Builder {
	b.inline = in
	return b
}

func (b v9InlineAvBuilder) SetHasFav(in map[int64]int8) InlineAvV2Builder {
	b.hasFav = in
	return b
}

func (b v9InlineAvBuilder) SetHotAidSet(in sets.Int64) InlineAvV2Builder {
	b.hotAidSet = in
	return b
}

func (b v9InlineAvBuilder) SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) InlineAvV2Builder {
	b.storyIcon = in
	return b
}

func (b v9InlineAvBuilder) SetHasCoin(in map[int64]int64) InlineAvV2Builder {
	b.hasCoin = in
	return b
}

func (b v9InlineAvBuilder) SetLikeStatState(in map[int64]*thumbupgrpc.StatState) InlineAvV2Builder {
	b.likeStatState = in
	return b
}

func (b v9InlineAvBuilder) constructArgs() jsoncard.Args {
	out := b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
	if b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid) {
		out.IsFollow = 1
	}
	return out
}

func (b v9InlineAvBuilder) constructUpArgs() *jsoncard.UpArgs {
	out := &jsoncard.UpArgs{}
	if b.arcPlayer == nil {
		return out
	}
	out.UpID = b.arcPlayer.Arc.Author.Mid
	out.UpName = b.arcPlayer.Arc.Author.Name
	out.UpFace = b.arcPlayer.Arc.Author.Face
	out.Selected = 0
	if b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid) {
		out.Selected = 1
	}
	return out
}

func (b v9InlineAvBuilder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in large cover v6: %+v", b.arcPlayer)
	}
	return nil
}

func (b v9InlineAvBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, nil, int(device.Build()), device.RawMobiApp(), true)
	if b.arcPlayer.Arc.RedirectURL != "" {
		extraFn = nil
	}
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

func (b v9InlineAvBuilder) constructVerticalArchiveURI() string {
	device := b.parent.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b v9InlineAvBuilder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

func (b *v9InlineAvBuilder) settingVerticalArchive() error {
	if !b.jumpGotoVerticalAv() {
		return errors.Errorf("not a vertical archive: %+v", b.arcPlayer)
	}
	b.output.FfCover = common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed)
	b.output.GotoIcon = b.constructGotoIcon(b.rcmd.IconType)
	args := b.constructArgs()
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(appcardmodel.GotoVerticalAv).
		UpdateURI(b.constructVerticalArchiveURI()).
		UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.parent.BuilderContext, &args,
			jsoncommon.WatchLater(false),
			jsoncommon.SwitchColumn(enableSwitchColumn),
			jsoncommon.AvDislikeInfo(b.rcmd.AvDislikeInfo),
			jsoncommon.Item(b.rcmd)))
	return nil
}

func (b v9InlineAvBuilder) constructGotoIcon(iconType int) *appcardmodel.GotoIcon {
	return appcardmodel.FillGotoIcon(iconType, b.storyIcon)
}

//nolint:unparam
func (b *v9InlineAvBuilder) settingCover() error {
	b.output.CoverLeftText1 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, "")
	b.output.CoverLeftIcon1 = appcardmodel.IconPlay
	b.output.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(b.output.CoverLeftIcon1,
		b.output.CoverLeftText1)
	b.output.CoverLeftText2 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.Danmaku, "")
	b.output.CoverLeftIcon2 = appcardmodel.IconDanmaku
	b.output.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(b.output.CoverLeftIcon2,
		b.output.CoverLeftText2)
	b.output.CoverRightText = appcardmodel.DurationString(b.arcPlayer.Arc.Duration)
	b.output.CoverRightContentDescription = appcardmodel.DurationContentDescription(b.arcPlayer.Arc.Duration)
	return nil
}

//nolint:unparam
func (b *v9InlineAvBuilder) settingRecommendReason() error {
	rcmdReason, _ := jsonreasonstyle.BuildInlineReasonText(
		b.rcmd.RcmdReason,
		b.resolveAuthorName(),
		b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid),
		true,
	)
	b.output.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		rcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	return nil
}

//nolint:unparam
func (b *v9InlineAvBuilder) settingThreePoint() error {
	args := b.constructArgs()
	meta, watchLater, switchColumn := b.constructThreePointPanelMeta()
	b.baseUpdater = b.baseUpdater.UpdateThreePoint(b.threePoint.ConstructArchvieThreePoint(&args, b.rcmd.AvDislikeInfo)).
		UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.parent.BuilderContext, &args,
			jsoncommon.WatchLater(watchLater),
			jsoncommon.SwitchColumn(switchColumn),
			jsoncommon.AvDislikeInfo(b.rcmd.AvDislikeInfo),
			jsoncommon.Item(b.rcmd))).
		UpdateThreePointMeta(meta)
	return nil
}

func (b *v9InlineAvBuilder) settingLikeButton() {
	selected := int8(0)
	if b.hasLike[b.arcPlayer.Arc.Aid] == 1 {
		selected = 1
	}
	count := b.arcPlayer.Arc.Stat.Like
	if stat, ok := b.likeStatState[b.arcPlayer.Arc.Aid]; ok {
		count = int32(stat.LikeNumber)
	}
	b.output.LikeButton = &jsoncard.LikeButton{
		Aid:      b.arcPlayer.Arc.Aid,
		Selected: selected,
		Count:    count,
		Event:    appcardmodel.EventlikeClick,
		EventV2:  appcardmodel.EventV2likeClick,
	}
	if b.inline != nil {
		b.output.LikeButton.ShowCount = b.inline.LikeButtonShowCount
		b.output.LikeButton.LikeResource = constructLikeButtonResource(b.inline.LikeResource, b.inline.LikeResourceHash)
		b.output.LikeButton.DisLikeResource = constructLikeButtonResource(b.inline.DisLikeResource, b.inline.DisLikeResourceHash)
		b.output.LikeButton.LikeNightResource = constructLikeButtonResource(b.inline.LikeNightResource, b.inline.LikeNightResourceHash)
		b.output.LikeButton.DisLikeNightResource = constructLikeButtonResource(b.inline.DisLikeNightResource, b.inline.DisLikeNightResourceHash)
	}
}

func (b v9InlineAvBuilder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

func (b v9InlineAvBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v9InlineAvBuilder) settingIsFav() {
	if b.hasFav[b.arcPlayer.Arc.Aid] == 1 {
		b.output.IsFav = true
	}
}

func (b v9InlineAvBuilder) settingIsHot() {
	if b.hotAidSet.Has(b.arcPlayer.Arc.Aid) {
		b.output.IsHot = true
	}
}

func (b v9InlineAvBuilder) settingIsCoin() {
	if isCoin, ok := b.hasCoin[b.arcPlayer.Arc.Aid]; ok && isCoin > 0 {
		b.output.IsCoin = true
	}
}

func (b v9InlineAvBuilder) settingInlineIcon() {
	if b.inline != nil {
		b.output.InlineProgressBar = &card.InlineProgressBar{
			IconDrag:     b.inline.IconDrag,
			IconDragHash: b.inline.IconDragHash,
			IconStop:     b.inline.IconStop,
			IconStopHash: b.inline.IconStopHash,
		}
	}
}

func (b *v9InlineAvBuilder) constructThreePointPanelMeta() (*threePointMeta.PanelMeta, bool, bool) {
	const (
		_inlineShareOrigin = "tm_inline"
		_inlineUgcShareId  = "tm.recommend.ugc.0"
	)
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.inline.ThreePointPanelType == 0 {
		return nil, true, enableSwitchColumn
	}
	watchLaterOnVersion := !b.parent.BuilderContext.VersionControl().Can("feed.inlineThreePointPanelSupported")
	return &threePointMeta.PanelMeta{
		PanelType:   int8(b.inline.ThreePointPanelType),
		ShareOrigin: _inlineShareOrigin,
		ShareId:     _inlineUgcShareId,
		FunctionalButtons: threePointMeta.ConstructFunctionalButton(false,
			b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint),
			appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column()),
			b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureDislikeText)),
	}, watchLaterOnVersion, false
}

func (b v9InlineAvBuilder) constructSharePlane() *appcardmodel.SharePlane {
	shareSubtitle, playNumber := card.GetShareSubtitle(b.arcPlayer.Arc.Stat.View)
	bvid_, _ := card.GetBvID(b.arcPlayer.Arc.Aid)
	return &appcardmodel.SharePlane{
		Title:         b.base.Title,
		ShareSubtitle: shareSubtitle,
		Desc:          b.arcPlayer.Arc.Desc,
		Cover:         b.base.Cover,
		Aid:           b.arcPlayer.Arc.Aid,
		Bvid:          bvid_,
		ShareTo:       appcardmodel.ShareTo,
		Author:        b.arcPlayer.Arc.Author.Name,
		AuthorId:      b.arcPlayer.Arc.Author.Mid,
		ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%d", b.rcmd.ID),
		PlayNumber:    playNumber,
	}
}

func (b v9InlineAvBuilder) constructPlayerArgs(canPlay bool) *jsoncard.PlayerArgs {
	if canPlay {
		out := b.archvieCommon.ConstructPlayerArgs(b.arcPlayer)
		if out == nil {
			return nil
		}
		out.ManualPlay = b.rcmd.ManualInline()
		out.HidePlayButton = appcardmodel.HidePlayButton
		out.ReportHistory = appcardmodel.ReportHistory
		out.ReportRequiredPlayDuration = appcardmodel.ReportRequiredPlayDuration
		out.ReportRequiredTime = appcardmodel.ReportRequiredTime
		return out
	}
	return &jsoncard.PlayerArgs{
		ManualPlay:                 b.rcmd.ManualInline(),
		HidePlayButton:             appcardmodel.HidePlayButton,
		ReportHistory:              appcardmodel.ReportHistory,
		ReportRequiredPlayDuration: appcardmodel.ReportRequiredPlayDuration,
		ReportRequiredTime:         appcardmodel.ReportRequiredTime,
	}
}

func (b v9InlineAvBuilder) Build() (*jsoncard.LargeCoverInline, error) {
	if b.arcPlayer == nil {
		return nil, errors.Errorf("empty `arcPlayer` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.inline == nil {
		return nil, errors.Errorf("empty `inline` field")
	}
	if card.CheckMidMaxInt32(b.arcPlayer.Arc.Author.Mid) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.arcPlayer.Arc.Author.Mid)
	}
	if err := b.ensureArchvieState(); err != nil {
		return nil, err
	}
	b.output = &jsoncard.LargeCoverInline{
		OfficialIcon:   b.constructOfficialIcon(),
		OfficialIconV2: b.constructOfficialIcon(),
	}
	b.baseUpdater = jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.arcPlayer.Arc.Pic).
		UpdateTitle(b.arcPlayer.Arc.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdateUpArgs(b.constructUpArgs())
	if err := b.settingCover(); err != nil {
		return nil, err
	}
	if err := b.settingRecommendReason(); err != nil {
		return nil, err
	}
	b.settingLikeButton()
	if err := b.settingThreePoint(); err != nil {
		return nil, err
	}
	if b.jumpGotoVerticalAv() {
		if err := b.settingVerticalArchive(); err != nil {
			return nil, err
		}
	}
	b.baseUpdater = b.baseUpdater.UpdatePlayerArgs(b.constructPlayerArgs(false))
	if b.arcPlayer.Arc.RedirectURL == "" {
		b.output.CanPlay = b.arcPlayer.Arc.Rights.Autoplay
		playerArgs := b.constructPlayerArgs(true)
		if playerArgs == nil {
			return nil, errors.New(fmt.Sprintf("LargeCoverV6 can not auto play: %d", b.arcPlayer.Arc.Aid))
		}
		b.baseUpdater = b.baseUpdater.UpdatePlayerArgs(playerArgs)
	}
	if b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid) {
		avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
			SetAvatarStatus(&jsoncard.AvatarStatus{
				Cover: b.arcPlayer.Arc.Author.Face,
				Text:  b.arcPlayer.Arc.Author.Name,
				Goto:  appcardmodel.GotoMid,
				Param: strconv.FormatInt(b.arcPlayer.Arc.Author.Mid, 10),
				Type:  appcardmodel.AvatarRound,
			}).Build()
		if err != nil {
			log.Error("Failed to build avatar: %+v", err)
		}
		b.output.Avatar = avatar
		b.output.OfficialIcon = appcardmodel.IconIsAttenm
		b.output.IsAtten = true
	}
	b.settingIsFav()
	b.settingIsHot()
	b.settingInlineIcon()
	b.settingIsCoin()
	if err := b.baseUpdater.Update(); err != nil {
		return nil, err
	}
	b.output.SharePlane = b.constructSharePlane()
	b.output.EnableDoubleClickLike = b.rcmd.DoubleInlineDbClickLike()
	b.output.Base = b.base
	for _, fn := range b.afterFn {
		fn(b.output)
	}

	return b.output, nil
}

func (b v9InlineAvBuilder) WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineAvV2Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
