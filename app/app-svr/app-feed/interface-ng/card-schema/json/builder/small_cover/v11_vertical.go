package jsonsmallcover

import (
	"bytes"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	"go-gateway/app/app-svr/app-feed/interface/common"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

type V11VerticalBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V11VerticalBuilder
	SetBase(*jsoncard.Base) V11VerticalBuilder
	SetRcmd(*ai.Item) V11VerticalBuilder
	SetArcPlayer(player *arcgrpc.ArcPlayer) V11VerticalBuilder
	SetAuthorCard(*accountgrpc.Card) V11VerticalBuilder
	SetTag(*taggrpc.Tag) V11VerticalBuilder
	SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) V11VerticalBuilder

	Build() (*jsoncard.SmallCoverV11, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV11)) V11VerticalBuilder
}

type v11VerticalBuilder struct {
	jsonbuilder.BuilderContext
	archvieCommon jsoncommon.ArchiveCommon
	threePoint    jsoncommon.ThreePoint
	base          *jsoncard.Base
	rcmd          *ai.Item
	arcPlayer     *arcgrpc.ArcPlayer
	tag           *taggrpc.Tag
	authorCard    *accountgrpc.Card
	storyIcon     map[int64]*appcardmodel.GotoIcon
	afterFn       []func(*jsoncard.SmallCoverV11)

	output *jsoncard.SmallCoverV11
}

func NewV11VerticalBuilder(ctx jsonbuilder.BuilderContext) V11VerticalBuilder {
	return v11VerticalBuilder{BuilderContext: ctx}
}

func (b v11VerticalBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V11VerticalBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v11VerticalBuilder) SetBase(base *jsoncard.Base) V11VerticalBuilder {
	b.base = base
	return b
}

func (b v11VerticalBuilder) SetRcmd(in *ai.Item) V11VerticalBuilder {
	b.rcmd = in
	return b
}

func (b v11VerticalBuilder) SetArcPlayer(in *arcgrpc.ArcPlayer) V11VerticalBuilder {
	b.arcPlayer = in
	return b
}

func (b v11VerticalBuilder) SetAuthorCard(in *accountgrpc.Card) V11VerticalBuilder {
	b.authorCard = in
	return b
}

func (b v11VerticalBuilder) SetTag(in *taggrpc.Tag) V11VerticalBuilder {
	b.tag = in
	return b
}

func (b v11VerticalBuilder) SetStoryIcon(in map[int64]*appcardmodel.GotoIcon) V11VerticalBuilder {
	b.storyIcon = in
	return b
}

func (b v11VerticalBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV11)) V11VerticalBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func (b v11VerticalBuilder) constructVerticalArchiveURI() string {
	device := b.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b v11VerticalBuilder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in small cover v11: %+v", b.arcPlayer)
	}
	return nil
}

func (b *v11VerticalBuilder) constructThreePoint(args *jsoncard.Args) *jsoncard.ThreePoint {
	out := b.threePoint.ConstructArchvieThreePoint(args, b.rcmd.AvDislikeInfo)
	appcard.ReplaceStoryDislikeReason(out.DislikeReasons, b.rcmd)
	return out
}

func (b v11VerticalBuilder) constructCover() string {
	if b.rcmd.StoryCover != "" {
		return b.rcmd.StoryCover
	}
	return b.arcPlayer.Arc.Pic
}

func (b v11VerticalBuilder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

func (b *v11VerticalBuilder) settingRecommendReason() {
	rcmdReason, desc := jsonreasonstyle.BuildRecommendReasonText(
		b.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		b.resolveAuthorName(),
		b.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid),
	)
	b.output.RcmdReason = rcmdReason
	b.output.Desc = desc
	if b.output.RcmdReason != "" {
		b.output.RcmdReasonStyle = &jsoncard.ReasonStyle{
			Text:             b.output.RcmdReason,
			TextColor:        "#FFFFFF",
			BgColor:          "#4DFFFFFF",
			BorderColor:      "#FFF1ED",
			TextColorNight:   "#FFFFFF",
			BgColorNight:     "#4DFFFFFF",
			BorderColorNight: "#3D2D29",
			BgStyle:          1,
		}
		return
	}
	upName := b.authorCard.GetName()
	if b.arcPlayer.Arc.Author.Name != "" {
		upName = b.arcPlayer.Arc.Author.Name
	}
	b.output.DescButton = &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    upName,
		URI:     appcardmodel.FillURI(appcardmodel.GotoMid, 0, 0, strconv.FormatInt(b.arcPlayer.Arc.Author.Mid, 10), nil),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
	b.output.GotoIcon = appcardmodel.FillGotoIcon(appcardmodel.AIUpStoryIconType, b.storyIcon)
}

func (b v11VerticalBuilder) Build() (*jsoncard.SmallCoverV11, error) {
	if b.arcPlayer == nil {
		return nil, errors.Errorf("empty `arcPlayer` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if err := b.ensureArchvieState(); err != nil {
		return nil, err
	}
	args := b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.constructCover()).
		UpdateTitle(b.arcPlayer.Arc.Title).
		UpdateURI(b.constructVerticalArchiveURI()).
		UpdateArgs(args).
		UpdatePlayerArgs(b.archvieCommon.ConstructPlayerArgs(b.arcPlayer)).
		UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.BuilderContext, &args,
			jsoncommon.WatchLater(true),
			jsoncommon.SwitchColumn(b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)),
			jsoncommon.AvDislikeInfo(b.rcmd.AvDislikeInfo),
			jsoncommon.Item(b.rcmd))).
		UpdateThreePoint(b.constructThreePoint(&args)).
		Update(); err != nil {
		return nil, err
	}
	b.output = &jsoncard.SmallCoverV11{
		Base:                         b.base,
		FfCover:                      common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed),
		CoverLeftText1:               appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, ""),
		CoverLeftIcon1:               appcardmodel.IconPlay,
		CoverRightText:               appcardmodel.DurationString(b.arcPlayer.Arc.Duration),
		CoverRightContentDescription: appcardmodel.DurationContentDescription(b.arcPlayer.Arc.Duration),
	}
	b.output.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(b.output.CoverLeftIcon1,
		b.output.CoverLeftText1)
	b.settingRecommendReason()

	for _, fn := range b.afterFn {
		fn(b.output)
	}
	return b.output, nil
}

func SmallCoverV11TalkBack() func(*jsoncard.SmallCoverV11) {
	return func(card *jsoncard.SmallCoverV11) {
		buffer := bytes.Buffer{}
		buffer.WriteString(appcardmodel.TalkBackCardType(card.Goto) + ",")
		buffer.WriteString(card.Title + ",")
		buffer.WriteString(card.CoverLeft1ContentDescription + ",")
		if card.CardGoto == appcardmodel.CardGotoAv {
			buffer.WriteString("时长" + card.CoverRightContentDescription + ",")
		}
		if card.Args.UpName != "" {
			buffer.WriteString("UP主" + card.Args.UpName + ",")
		}
		if card.RcmdReason != "" {
			buffer.WriteString(card.RcmdReason + ",")
		}
		if card.DescButton != nil && card.CardGoto != appcardmodel.CardGotoAv {
			buffer.WriteString(card.DescButton.Text)
		}
		card.TalkBack = buffer.String()
	}
}

func V11FilledByMultiMaterials(arg *tunnelV2.Material, item *ai.Item) func(*jsoncard.SmallCoverV11) {
	return func(card *jsoncard.SmallCoverV11) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
		if arg.Desc != "" {
			if card.DescButton != nil {
				card.DescButton.Text = arg.Desc
			}
			card.Desc = arg.Desc
			if item.Goto == "av" {
				card.RcmdReason = ""
				card.RcmdReasonStyle = nil
			}
		}
	}
}
