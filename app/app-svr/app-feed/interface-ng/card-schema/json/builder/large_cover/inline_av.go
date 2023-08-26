package large_cover

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	"go-gateway/app/app-svr/app-feed/interface/common"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

type InlineAvBuilder interface {
	Parent() LargeCoverInlineBuilderFactory

	SetBase(*jsoncard.Base) InlineAvBuilder
	SetRcmd(*ai.Item) InlineAvBuilder
	SetArcPlayer(*arcgrpc.ArcPlayer) InlineAvBuilder
	SetTag(*taggrpc.Tag) InlineAvBuilder
	SetAuthorCard(*accountgrpc.Card) InlineAvBuilder
	SetHasLike(map[int64]int8) InlineAvBuilder
	SetInline(*Inline) InlineAvBuilder
	SetStoryIcon(map[int64]*appcardmodel.GotoIcon) InlineAvBuilder

	Build() (*jsoncard.LargeCoverInline, error)
	WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineAvBuilder
}

type v6InlineAvBuilder struct {
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

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.LargeCoverInline
	afterFn     []func(*jsoncard.LargeCoverInline)
}

func (b v6InlineAvBuilder) Parent() LargeCoverInlineBuilderFactory {
	return b.parent
}

func (b v6InlineAvBuilder) SetBase(base *jsoncard.Base) InlineAvBuilder {
	b.base = base
	return b
}

func (b v6InlineAvBuilder) SetRcmd(item *ai.Item) InlineAvBuilder {
	b.rcmd = item
	return b
}

func (b v6InlineAvBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) InlineAvBuilder {
	b.arcPlayer = in
	return b
}

func (b v6InlineAvBuilder) SetTag(in *taggrpc.Tag) InlineAvBuilder {
	b.tag = in
	return b
}

func (b v6InlineAvBuilder) SetAuthorCard(in *accountgrpc.Card) InlineAvBuilder {
	b.authorCard = in
	return b
}

func (b v6InlineAvBuilder) SetHasLike(in map[int64]int8) InlineAvBuilder {
	b.hasLike = in
	return b
}

func (b v6InlineAvBuilder) SetInline(in *Inline) InlineAvBuilder {
	b.inline = in
	return b
}

func (b v6InlineAvBuilder) SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) InlineAvBuilder {
	b.storyIcon = in
	return b
}

func (b v6InlineAvBuilder) constructArgs() jsoncard.Args {
	out := b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
	if b.parent.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid) {
		out.IsFollow = 1
	}
	return out
}

func (b v6InlineAvBuilder) constructUpArgs() *jsoncard.UpArgs {
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

func (b v6InlineAvBuilder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in large cover v6: %+v", b.arcPlayer)
	}
	return nil
}

func (b v6InlineAvBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, nil, int(device.Build()), device.RawMobiApp(), true)
	if b.arcPlayer.Arc.RedirectURL != "" {
		extraFn = nil
	}
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

func (b v6InlineAvBuilder) constructVerticalArchiveURI() string {
	device := b.parent.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b v6InlineAvBuilder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

func (b *v6InlineAvBuilder) settingVerticalArchive() error {
	if !b.jumpGotoVerticalAv() {
		return errors.Errorf("not a vertical archive: %+v", b.arcPlayer)
	}
	b.output.FfCover = common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed)
	b.output.GotoIcon = b.constructGotoIcon(b.rcmd.IconType)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(appcardmodel.GotoVerticalAv).
		UpdateURI(b.constructVerticalArchiveURI())
	return nil
}

func (b v6InlineAvBuilder) constructGotoIcon(iconType int) *appcardmodel.GotoIcon {
	return appcardmodel.FillGotoIcon(iconType, b.storyIcon)
}

//nolint:unparam
func (b *v6InlineAvBuilder) settingCover() error {
	b.output.CoverLeftText1 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, "")
	b.output.CoverLeftIcon1 = appcardmodel.IconPlay
	b.output.CoverLeftText2 = appcardmodel.StatString(b.arcPlayer.Arc.Stat.Danmaku, "")
	b.output.CoverLeftIcon2 = appcardmodel.IconDanmaku
	b.output.CoverRightText = appcardmodel.DurationString(b.arcPlayer.Arc.Duration)
	return nil
}

//nolint:unparam
func (b *v6InlineAvBuilder) settingRecommendReason() error {
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
func (b *v6InlineAvBuilder) settingThreePoint() error {
	args := b.constructArgs()
	b.baseUpdater = b.baseUpdater.UpdateThreePoint(b.threePoint.ConstructArchvieThreePoint(&args, b.rcmd.AvDislikeInfo)).
		UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.parent.BuilderContext, &args,
			jsoncommon.WatchLater(true),
			jsoncommon.SwitchColumn(false),
			jsoncommon.AvDislikeInfo(b.rcmd.AvDislikeInfo),
			jsoncommon.Item(b.rcmd)))
	return nil
}

func (b *v6InlineAvBuilder) settingLikeButton() {
	selected := int8(0)
	if b.hasLike[b.arcPlayer.Arc.Aid] == 1 {
		selected = 1
	}
	b.output.LikeButton = &jsoncard.LikeButton{
		Aid:      b.arcPlayer.Arc.Aid,
		Selected: selected,
		Count:    b.arcPlayer.Arc.Stat.Like,
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

func constructLikeButtonResource(url, hash string) *jsoncard.LikeButtonResource {
	return &jsoncard.LikeButtonResource{
		URL:  url,
		Hash: hash,
	}
}

func (b v6InlineAvBuilder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

func (b v6InlineAvBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v6InlineAvBuilder) constructPlayerArgs(canPlay bool) *jsoncard.PlayerArgs {
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

func (b v6InlineAvBuilder) Build() (*jsoncard.LargeCoverInline, error) {
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
	b.settingLikeButton()
	if err := b.settingThreePoint(); err != nil {
		return nil, err
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
	if err := b.baseUpdater.Update(); err != nil {
		return nil, err
	}
	b.output.Base = b.base
	for _, fn := range b.afterFn {
		fn(b.output)
	}

	return b.output, nil
}

func (b v6InlineAvBuilder) WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineAvBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
