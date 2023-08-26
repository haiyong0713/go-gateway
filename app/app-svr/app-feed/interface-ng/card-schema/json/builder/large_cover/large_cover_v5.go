package large_cover

import (
	"fmt"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	"go-gateway/app/app-svr/app-feed/interface/common"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

type LargeCoverV5Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) LargeCoverV5Builder
	SetBase(*jsoncard.Base) LargeCoverV5Builder
	SetRcmd(*ai.Item) LargeCoverV5Builder
	SetArcPlayer(*arcgrpc.ArcPlayer) LargeCoverV5Builder
	SetTag(*taggrpc.Tag) LargeCoverV5Builder
	SetAuthorCard(*accountgrpc.Card) LargeCoverV5Builder
	SetChannelCard(*channelgrpc.ChannelCard) LargeCoverV5Builder

	Build() (*jsoncard.LargeCoverV5, error)
	WithAfter(req ...func(*jsoncard.LargeCoverV5)) LargeCoverV5Builder
}

type largeCoverV5Builder struct {
	archvieCommon jsoncommon.ArchiveCommon
	jsonbuilder.BuilderContext
	rcmd        *ai.Item
	base        *jsoncard.Base
	arcPlayer   *arcgrpc.ArcPlayer
	tag         *taggrpc.Tag
	authorCard  *accountgrpc.Card
	channelCard *channelgrpc.ChannelCard

	baseUpdater jsonbuilder.BaseUpdater
	output      *jsoncard.LargeCoverV5
	threePoint  jsoncommon.ThreePoint
	afterFn     []func(*jsoncard.LargeCoverV5)
}

func NewLargeCoverV5Builder(ctx jsonbuilder.BuilderContext) LargeCoverV5Builder {
	return largeCoverV5Builder{BuilderContext: ctx}
}

func (b largeCoverV5Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) LargeCoverV5Builder {
	b.BuilderContext = ctx
	return b
}

func (b largeCoverV5Builder) SetBase(base *jsoncard.Base) LargeCoverV5Builder {
	b.base = base
	return b
}

func (b largeCoverV5Builder) SetRcmd(item *ai.Item) LargeCoverV5Builder {
	b.rcmd = item
	return b
}

func (b largeCoverV5Builder) SetArcPlayer(in *arcgrpc.ArcPlayer) LargeCoverV5Builder {
	b.arcPlayer = in
	return b
}

func (b largeCoverV5Builder) SetTag(in *taggrpc.Tag) LargeCoverV5Builder {
	b.tag = in
	return b
}

func (b largeCoverV5Builder) SetAuthorCard(in *accountgrpc.Card) LargeCoverV5Builder {
	b.authorCard = in
	return b
}

func (b largeCoverV5Builder) SetChannelCard(in *channelgrpc.ChannelCard) LargeCoverV5Builder {
	b.channelCard = in
	return b
}

func (b largeCoverV5Builder) ensureArchvieState() error {
	if !appcardmodel.AvIsNormalGRPC(b.arcPlayer) {
		return errors.Errorf("insufficient archvie in large cover v6: %+v", b.arcPlayer)
	}
	return nil
}

func (b largeCoverV5Builder) constructURI() string {
	device := b.BuilderContext.Device()
	trackID := b.rcmd.TrackID
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		trackID, nil, int(device.Build()), device.RawMobiApp(), true)
	if b.arcPlayer.Arc.RedirectURL != "" {
		extraFn = nil
	}
	return b.archvieCommon.ConstructArchiveURI(b.arcPlayer.Arc.Aid, extraFn)
}

func (b largeCoverV5Builder) constructVerticalArchiveURI() string {
	device := b.BuilderContext.Device()
	extraFn := appcardmodel.ArcPlayHandler(b.arcPlayer.Arc, appcardmodel.ArcPlayURL(b.arcPlayer, 0),
		b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true)
	return b.archvieCommon.ConstructVerticalArchiveURI(b.arcPlayer.Arc.Aid, device, extraFn)
}

func (b largeCoverV5Builder) constructArgs() jsoncard.Args {
	return b.archvieCommon.ConstructArgs(b.arcPlayer, b.tag)
}

func (b largeCoverV5Builder) resolveOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b largeCoverV5Builder) jumpGotoVerticalAv() bool {
	return b.rcmd.JumpGoto == appfeedmodel.GotoVerticalAv
}

func (b *largeCoverV5Builder) settingVerticalArchive() error {
	if !b.jumpGotoVerticalAv() {
		return errors.Errorf("not a vertical archive: %+v", b.arcPlayer)
	}
	b.output.FfCover = common.Ffcover(b.arcPlayer.Arc.FirstFrame, appcardmodel.FfCoverFromFeed)
	b.baseUpdater = b.baseUpdater.
		UpdateGoto(appcardmodel.GotoVerticalAv).
		UpdateURI(b.constructVerticalArchiveURI())
	return nil
}

func (b largeCoverV5Builder) resolveAuthorName() string {
	if b.authorCard == nil {
		return b.arcPlayer.Arc.Author.Name
	}
	return b.authorCard.Name
}

//nolint:unparam
func (b *largeCoverV5Builder) settingRecommendReason() error {
	rcmdReason, desc := jsonreasonstyle.BuildInlineReasonText(
		b.rcmd.RcmdReason,
		b.resolveAuthorName(),
		b.BuilderContext.IsAttentionTo(b.arcPlayer.Arc.Author.Mid),
		true,
	)
	b.output.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		rcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
	)
	button, err := b.constructDescButton(desc)
	if err != nil {
		log.Error("Failed to construct desc button: %+v", err)
	}
	b.output.DescButton = button
	return nil
}

//nolint:unparam
func (b largeCoverV5Builder) constructDescButton(in string) (*jsoncard.Button, error) {
	if in != "" {
		return b.archvieCommon.ConstructDescButtonFromDesc(in), nil
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

//nolint:unparam
func (b *largeCoverV5Builder) settingThreePoint() error {
	args := b.constructArgs()
	b.baseUpdater = b.baseUpdater.UpdateThreePoint(b.threePoint.ConstructArchvieThreePoint(&args, b.rcmd.AvDislikeInfo)).
		UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.BuilderContext, &args,
			jsoncommon.WatchLater(true),
			jsoncommon.SwitchColumn(false),
			jsoncommon.AvDislikeInfo(b.rcmd.AvDislikeInfo),
			jsoncommon.Item(b.rcmd)))
	return nil
}

func (b largeCoverV5Builder) Build() (*jsoncard.LargeCoverV5, error) {
	if b.arcPlayer == nil {
		return nil, errors.Errorf("empty `arcPlayer` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if card.CheckMidMaxInt32(b.arcPlayer.Arc.Author.Mid) && b.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.arcPlayer.Arc.Author.Mid)
	}
	if err := b.ensureArchvieState(); err != nil {
		return nil, err
	}
	b.output = &jsoncard.LargeCoverV5{
		OfficialIcon:   b.resolveOfficialIcon(),
		CoverLeftText1: appcardmodel.StatString(b.arcPlayer.Arc.Stat.View, ""),
		CoverLeftIcon1: appcardmodel.IconPlay,
		CoverLeftText2: appcardmodel.StatString(b.arcPlayer.Arc.Stat.Danmaku, ""),
		CoverLeftIcon2: appcardmodel.IconDanmaku,
		CoverRightText: appcardmodel.DurationString(b.arcPlayer.Arc.Duration),
	}
	b.baseUpdater = jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.arcPlayer.Arc.Pic).
		UpdateTitle(b.arcPlayer.Arc.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs())

	if b.arcPlayer.Arc.RedirectURL == "" {
		b.baseUpdater = b.baseUpdater.UpdatePlayerArgs(b.archvieCommon.ConstructPlayerArgs(b.arcPlayer))
		b.output.CanPlay = b.arcPlayer.Arc.Rights.Autoplay
	}
	if err := b.settingRecommendReason(); err != nil {
		return nil, err
	}
	if b.jumpGotoVerticalAv() {
		if err := b.settingVerticalArchive(); err != nil {
			return nil, err
		}
	}
	if err := b.settingThreePoint(); err != nil {
		return nil, err
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

func (b largeCoverV5Builder) WithAfter(req ...func(*jsoncard.LargeCoverV5)) LargeCoverV5Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func LargeCoverV5FromOp(cpr *resourcegrpc.CardPosRec) func(*jsoncard.LargeCoverV5) {
	return func(card *jsoncard.LargeCoverV5) {
		if cpr == nil {
			return
		}
		if cpr.Title != "" {
			card.Title = cpr.Title
		}
		if cpr.Cover != "" {
			card.Cover = cpr.Cover
		}
	}
}
