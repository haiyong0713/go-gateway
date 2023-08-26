package jsonogvsmallcover

import (
	"bytes"
	"math"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	"github.com/pkg/errors"
)

type OgvSmallCoverBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) OgvSmallCoverBuilder
	SetBase(*jsoncard.Base) OgvSmallCoverBuilder
	SetRcmd(*ai.Item) OgvSmallCoverBuilder
	SetEpisode(*pgccard.EpisodeCard) OgvSmallCoverBuilder
	SetEpMaterilas(*deliverygrpc.EpMaterial) OgvSmallCoverBuilder

	Build() (*jsoncard.OgvSmallCover, error)
	WithAfter(req ...func(card *jsoncard.OgvSmallCover)) OgvSmallCoverBuilder
}

type ogvSmallCoverBuilder struct {
	jsonbuilder.BuilderContext
	threePoint  jsoncommon.ThreePoint
	ogvCommon   jsoncommon.OgvEpisode
	base        *jsoncard.Base
	rcmd        *ai.Item
	episodeCard *pgccard.EpisodeCard
	epMaterial  *deliverygrpc.EpMaterial
	afterFn     []func(card *jsoncard.OgvSmallCover)
}

type OgvEpCover struct {
	CoverLeftText1 string
	CoverLeftIcon1 appcardmodel.Icon
	CoverLeftText2 string
	CoverLeftIcon2 appcardmodel.Icon
	CoverRightText string
}

func NewOgvSmallCoverBuilder(ctx jsonbuilder.BuilderContext) OgvSmallCoverBuilder {
	return ogvSmallCoverBuilder{BuilderContext: ctx, ogvCommon: jsoncommon.OgvEpisode{}}
}

func (b ogvSmallCoverBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) OgvSmallCoverBuilder {
	b.BuilderContext = ctx
	return b
}

func (b ogvSmallCoverBuilder) SetBase(base *jsoncard.Base) OgvSmallCoverBuilder {
	b.base = base
	return b
}

func (b ogvSmallCoverBuilder) SetRcmd(item *ai.Item) OgvSmallCoverBuilder {
	b.rcmd = item
	return b
}

func (b ogvSmallCoverBuilder) SetEpisode(in *pgccard.EpisodeCard) OgvSmallCoverBuilder {
	b.episodeCard = in
	return b
}

func (b ogvSmallCoverBuilder) SetEpMaterilas(in *deliverygrpc.EpMaterial) OgvSmallCoverBuilder {
	b.epMaterial = in
	return b
}

func (b ogvSmallCoverBuilder) Build() (*jsoncard.OgvSmallCover, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.episodeCard == nil {
		return nil, errors.Errorf("empty `episode` field")
	}
	if b.episodeCard.Season == nil {
		return nil, errors.Errorf("empty `episode.season` field")
	}
	if b.episodeCard.TianmaSmallCardMeta == nil {
		return nil, errors.Errorf("empty `episode.TianmaSmallCardMeta` field")
	}
	output := &jsoncard.OgvSmallCover{}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateURI(b.ogvCommon.ConstructOgvURI(b.episodeCard, b.Device(), b.rcmd)).
		UpdateCover(b.constructCover()).
		UpdateTitle(b.constructTitle()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateParam(strconv.FormatInt(int64(b.episodeCard.EpisodeId), 10)).
		Update(); err != nil {
		return nil, err
	}

	output.Subtitle = b.constructSubTitle()
	output.RcmdReasonStyle, output.DescButton = b.constructEXPReasonStyleOrDescButton()
	if epCover := b.ogvCommon.ConstructEpCover(b.episodeCard); epCover != nil {
		output.CoverLeftText1 = epCover.CoverLeftText1
		output.CoverLeftIcon1 = epCover.CoverLeftIcon1
		output.CoverLeft1ContentDescription = epCover.CoverLeft1ContentDescription
		output.CoverLeftText2 = epCover.CoverLeftText2
		output.CoverLeftIcon2 = epCover.CoverLeftIcon2
		output.CoverLeft2ContentDescription = epCover.CoverLeft2ContentDescription
	}
	output.CoverRightText = b.ogvCommon.ConstructOgvRightText(b.episodeCard, b.rcmd.OgvHasScore(), true)
	output.BadgeStyle = b.constructOgvBadgeStyle()
	output.OgvCreativeId = b.rcmd.CreativeId

	if b.epMaterial != nil && b.epMaterial.GetPowerCorner().GetPowerPicSun() != "" &&
		b.epMaterial.GetPowerCorner().GetPowerPicNight() != "" && b.epMaterial.GetPowerCorner().GetWidth() > 0 &&
		b.epMaterial.GetPowerCorner().GetHeight() > 0 {
		output.LeftCoverBadgeNewStyle = &jsoncard.ReasonStyle{
			IconURL:      b.epMaterial.GetPowerCorner().GetPowerPicSun(),
			IconURLNight: b.epMaterial.GetPowerCorner().GetPowerPicNight(),
			IconWidth:    int32(math.Floor(float64(b.epMaterial.GetPowerCorner().GetWidth()) / float64(b.epMaterial.GetPowerCorner().GetHeight()) * float64(21))),
			IconHeight:   21,
		}
	}

	output.Base = b.base
	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b ogvSmallCoverBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	enableFeedback := b.enableFeedback()
	enableWatched := b.enableWatched()
	if b.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructOGVThreePointV2(b.BuilderContext, enableSwitchColumn, enableFeedback, enableWatched)
	}
	return b.threePoint.ConstructOGVThreePointV2Legacy(b.BuilderContext, enableSwitchColumn, enableFeedback, enableWatched)
}

func (b ogvSmallCoverBuilder) constructReasonStyleOrDescButton(title string) (*jsoncard.ReasonStyle, *jsoncard.Button) {
	if b.rcmd.RcmdReason != nil {
		reasonStyle := jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
		)
		return reasonStyle, nil
	}
	return nil, b.constructDescButtonFromOgvRcmdReason(title)
}

// nolint:gomnd
func (b ogvSmallCoverBuilder) constructEXPReasonStyleOrDescButton() (*jsoncard.ReasonStyle, *jsoncard.Button) {
	title := b.rcmd.CustomizedTitle
	if b.epMaterial != nil && b.epMaterial.Title != "" {
		title = b.epMaterial.Title
	}
	if b.rcmd.RcmdReason != nil && b.rcmd.CustomizedOGVDesc != "" {
		reasonStyle := jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
		)
		desc := b.constructDescButtonFromOgvRcmdReason(title)
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
	return b.constructReasonStyleOrDescButton(title)
}

func (b ogvSmallCoverBuilder) constructCover() string {
	if appcardmodel.IsValidCover(b.rcmd.CustomizedCover) {
		return b.rcmd.CustomizedCover
	}
	if b.epMaterial != nil && b.epMaterial.Cover != "" {
		return b.epMaterial.Cover
	}
	return b.episodeCard.Cover
}

func (b ogvSmallCoverBuilder) constructTitle() string {
	if b.rcmd.CustomizedTitle != "" {
		return b.rcmd.CustomizedTitle
	}
	if b.epMaterial != nil && b.epMaterial.Title != "" {
		return b.epMaterial.Title
	}
	return b.episodeCard.TianmaSmallCardMeta.Title
}

func (b ogvSmallCoverBuilder) constructSubTitle() string {
	if b.rcmd.CustomizedSubtitle != "" {
		return b.rcmd.CustomizedSubtitle
	}
	if b.rcmd.CustomizedTitle != "" {
		return ""
	}
	switch b.rcmd.OgvNewStyle {
	case appcardmodel.OgvCustomizedType:
		return ""
	default:
	}
	return b.episodeCard.TianmaSmallCardMeta.SubTitle
}

func (b ogvSmallCoverBuilder) constructOgvBadgeStyle() *jsoncard.ReasonStyle {
	return jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, b.episodeCard.TianmaSmallCardMeta.BadgeInfo.Text)
}

func (b ogvSmallCoverBuilder) constructDescButtonFromOgvRcmdReason(title string) *jsoncard.Button {
	text := b.episodeCard.TianmaSmallCardMeta.RcmdReason
	if b.epMaterial != nil && b.epMaterial.Desc != "" {
		text = b.epMaterial.Desc
	}
	if b.rcmd.CustomizedOGVDesc != "" {
		text = b.rcmd.CustomizedOGVDesc
	}
	if title == "" {
		text = b.episodeCard.TianmaSmallCardMeta.RcmdReason
	}
	return &jsoncard.Button{
		Text:    text,
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
		Type:    appcardmodel.ButtonGrey,
	}
}

func (b ogvSmallCoverBuilder) enableFeedback() bool {
	return b.BuilderContext.VersionControl().Can("feed.enableOGVFeedback") &&
		b.rcmd.OgvDislikeInfo >= 1 &&
		appcardmodel.Columnm[appcardmodel.ColumnStatus(b.BuilderContext.IndexParam().Column())] == appcardmodel.ColumnSvrDouble
}

func (b ogvSmallCoverBuilder) enableWatched() bool {
	return b.rcmd.OgvDislikeInfo == ai.OgvWatched
}

func (b ogvSmallCoverBuilder) WithAfter(req ...func(card *jsoncard.OgvSmallCover)) OgvSmallCoverBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func OGVSmallCoverTalkBack() func(cover *jsoncard.OgvSmallCover) {
	return func(card *jsoncard.OgvSmallCover) {
		buffer := bytes.Buffer{}
		buffer.WriteString(appcardmodel.TalkBackCardType(card.Goto) + ",")
		buffer.WriteString(card.Title + ",")
		buffer.WriteString(appcardmodel.CoverIconContentDescription(card.CoverLeftIcon1, card.CoverLeftText1) + ",")
		buffer.WriteString(appcardmodel.CoverIconContentDescription(card.CoverLeftIcon2, card.CoverLeftText2) + ",")
		if card.CoverRightText != "" {
			buffer.WriteString(card.CoverRightText + ",")
		}
		if card.Args.UpName != "" {
			buffer.WriteString("UP主" + card.Args.UpName + ",")
		}
		if card.DescButton != nil {
			buffer.WriteString(card.DescButton.Text + ",")
		}
		if card.BadgeStyle != nil {
			buffer.WriteString(card.BadgeStyle.Text)
		}
		card.TalkBack = buffer.String()
	}
}
