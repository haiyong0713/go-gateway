package large_cover

import (
	"fmt"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	"github.com/pkg/errors"
)

type InlinePGCBuilder interface {
	Parent() LargeCoverInlineBuilderFactory
	SetBase(*jsoncard.Base) InlinePGCBuilder
	SetRcmd(*ai.Item) InlinePGCBuilder
	SetEpisode(*pgcinline.EpisodeCard) InlinePGCBuilder
	SetHasLike(map[int64]int8) InlinePGCBuilder
	SetInline(*Inline) InlinePGCBuilder
	Build() (*jsoncard.LargeCoverInline, error)
	WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlinePGCBuilder
}

type v7InlinePGCBuilder struct {
	jsoncommon.ThreePoint
	parent  *largeCoverInlineBuilderFactory
	base    *jsoncard.Base
	rcmd    *ai.Item
	hasLike map[int64]int8
	inline  *Inline
	episode *pgcinline.EpisodeCard
	afterFn []func(*jsoncard.LargeCoverInline)
}

func (b v7InlinePGCBuilder) Parent() LargeCoverInlineBuilderFactory {
	return b.parent
}

func (b v7InlinePGCBuilder) SetBase(base *jsoncard.Base) InlinePGCBuilder {
	b.base = base
	return b
}

func (b v7InlinePGCBuilder) SetRcmd(in *ai.Item) InlinePGCBuilder {
	b.rcmd = in
	return b
}

func (b v7InlinePGCBuilder) SetHasLike(in map[int64]int8) InlinePGCBuilder {
	b.hasLike = in
	return b
}

func (b v7InlinePGCBuilder) SetInline(in *Inline) InlinePGCBuilder {
	b.inline = in
	return b
}

func (b v7InlinePGCBuilder) SetEpisode(in *pgcinline.EpisodeCard) InlinePGCBuilder {
	b.episode = in
	return b
}

func (b v7InlinePGCBuilder) constructThreePointV2(enableSwitchColumn bool) []*jsoncard.ThreePointV2 {
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.ConstructDefaultThreePointV2(b.parent.BuilderContext, enableSwitchColumn)
	}
	return b.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, enableSwitchColumn)
}

func (b v7InlinePGCBuilder) constructURI(in *pgcinline.EpisodeCard, device cardschema.Device, rcmd *ai.Item) string {
	param := in.Url
	if param == "" {
		plat := device.Plat()
		build := int(device.Build())
		param = appcardmodel.FillURI(appcardmodel.GotoBangumi, plat, build, strconv.FormatInt(int64(in.EpisodeId), 10), nil)
	}
	return appcardmodel.FillURI("", 0, 0, param, appcardmodel.PGCTrackIDHandler(rcmd))
}

func (b v7InlinePGCBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		Aid:            b.episode.Aid,
		Cid:            b.episode.Cid,
		EpID:           int64(b.episode.EpisodeId),
		IsPreview:      b.episode.IsPreview,
		Type:           appcardmodel.GotoBangumi,
		Duration:       b.episode.Duration,
		SubType:        b.episode.Season.Type,
		SeasonID:       int64(b.episode.Season.SeasonId),
		ManualPlay:     b.rcmd.ManualInline(),
		HidePlayButton: appcardmodel.HidePlayButton,
	}
}

func (b v7InlinePGCBuilder) constructPlayerWidget() *jsoncard.InlinePlayerWidget {
	if b.episode.Widget == nil {
		return nil
	}
	return &jsoncard.InlinePlayerWidget{
		Title: b.episode.Widget.Title,
		Desc:  b.episode.Widget.Desc,
	}
}

func (b v7InlinePGCBuilder) settingInlineIcon(out *jsoncard.LargeCoverInline) {
	if b.inline != nil {
		out.InlineProgressBar = &card.InlineProgressBar{
			IconDrag:     b.inline.IconDrag,
			IconDragHash: b.inline.IconDragHash,
			IconStop:     b.inline.IconStop,
			IconStopHash: b.inline.IconStopHash,
		}
	}
}

func (b v7InlinePGCBuilder) Build() (*jsoncard.LargeCoverInline, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.inline == nil {
		return nil, errors.Errorf("empty `inline` field")
	}
	if b.episode == nil {
		return nil, errors.Errorf("empty `episode` field")
	}
	if b.episode.Season == nil {
		return nil, errors.Errorf("empty `episode.Season` field")
	}
	meta, switchColumn := b.constructThreePointPanelMeta()
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.episode.Cover).
		UpdateTitle(fmt.Sprintf("%s %s", b.episode.Season.Title, b.episode.NewDesc)).
		UpdateURI(b.constructURI(b.episode, b.parent.Device(), b.rcmd)).
		UpdatePlayerArgs(b.constructPlayerArgs()).
		UpdateThreePoint(b.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2(switchColumn)).
		UpdateThreePointMeta(meta).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverInline{
		Base:                         b.base,
		CoverRightText:               appcardmodel.DurationString(b.episode.Duration),
		CoverRightContentDescription: appcardmodel.DurationContentDescription(b.episode.Duration),
		BadgeStyle:                   jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, b.episode.Season.TypeName),
		PlayerWidget:                 b.constructPlayerWidget(),
		LikeButton:                   b.constructLikeButton(),
		SharePlane:                   b.constructSharePlane(),
	}
	b.settingInlineIcon(out)
	if b.episode.PlayerInfo != nil {
		out.CanPlay = 1
	}
	if b.rcmd.NoPlay == 1 {
		out.CanPlay = 0
	}
	if b.episode.Stat != nil {
		out.CoverLeftText1 = appcardmodel.StatString(int32(b.episode.Stat.Play), "")
		out.CoverLeftIcon1 = appcardmodel.IconPlay
		out.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon1,
			out.CoverLeftText1)
		out.CoverLeftText2 = appcardmodel.StatString(int32(b.episode.Stat.Follow), "")
		out.CoverLeftIcon2 = appcardmodel.IconFavorite
		out.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon2,
			out.CoverLeftText2)
	}
	if b.rcmd.RcmdReason != nil && b.enableInlineRcmd() {
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext))
	}
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v7InlinePGCBuilder) enableInlineRcmd() bool {
	if appcardmodel.Columnm[appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column())] == appcardmodel.ColumnSvrDouble {
		return true
	}
	return b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSingleRcmdReason)
}

func (b v7InlinePGCBuilder) WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlinePGCBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func (b v7InlinePGCBuilder) constructLikeButton() *jsoncard.LikeButton {
	selected := int8(0)
	if b.hasLike[b.episode.Aid] == 1 {
		selected = 1
	}
	out := &jsoncard.LikeButton{
		Aid:      b.episode.Aid,
		Selected: selected,
		Count:    int32(b.episode.Stat.Like),
		Event:    appcardmodel.EventlikeClick,
		EventV2:  appcardmodel.EventV2likeClick,
	}
	if b.inline != nil {
		out.ShowCount = b.inline.LikeButtonShowCount
		out.LikeResource = constructLikeButtonResource(b.inline.LikeResource, b.inline.LikeResourceHash)
		out.DisLikeResource = constructLikeButtonResource(b.inline.DisLikeResource, b.inline.DisLikeResourceHash)
		out.LikeNightResource = constructLikeButtonResource(b.inline.LikeNightResource, b.inline.LikeNightResourceHash)
		out.DisLikeNightResource = constructLikeButtonResource(b.inline.DisLikeNightResource, b.inline.DisLikeNightResourceHash)
	}
	return out
}

func (b v7InlinePGCBuilder) constructThreePointPanelMeta() (*threePointMeta.PanelMeta, bool) {
	const (
		_inlineShareOrigin = "tm_inline"
		_inlineOgvShareId  = "tm.recommend.ogv.0"
	)
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.inline.ThreePointPanelType == 0 {
		return nil, enableSwitchColumn
	}
	return &threePointMeta.PanelMeta{
		PanelType:   int8(b.inline.ThreePointPanelType),
		ShareOrigin: _inlineShareOrigin,
		ShareId:     _inlineOgvShareId,
		FunctionalButtons: threePointMeta.ConstructFunctionalButton(true,
			b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint),
			appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column()),
			b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureDislikeText)),
	}, false
}

func (b v7InlinePGCBuilder) constructSharePlane() *appcardmodel.SharePlane {
	bvid_, _ := card.GetBvID(b.episode.Aid)
	shareSubtitle, playNumber := card.GetShareSubtitle(int32(b.episode.GetStat().Play))
	return &appcardmodel.SharePlane{
		Title:         fmt.Sprintf("%s %s", b.episode.Season.Title, b.episode.NewDesc),
		ShareSubtitle: shareSubtitle,
		Desc:          b.episode.GetSeason().GetNewEpShow(),
		Cover:         b.episode.Cover,
		Aid:           b.episode.Aid,
		Bvid:          bvid_,
		EpId:          b.episode.EpisodeId,
		SeasonId:      b.episode.GetSeason().GetSeasonId(),
		ShareTo:       appcardmodel.ShareTo,
		PlayNumber:    playNumber,
		ShareFrom:     appcardmodel.InlinePGCShareFrom,
		SeasonTitle:   b.episode.Season.Title,
	}
}
