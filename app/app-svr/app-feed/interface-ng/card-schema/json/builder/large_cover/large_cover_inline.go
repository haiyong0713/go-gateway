package large_cover

import (
	"bytes"
	"math"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
)

// LargeCoverInlineBuilderFactory is
type LargeCoverInlineBuilderFactory interface {
	ReplaceContext(jsonbuilder.BuilderContext) LargeCoverInlineBuilderFactory
	MarkAsSpecialCard(bool) LargeCoverInlineBuilderFactory
	SpecialCard() bool

	DeriveArcPlayerBuilder() InlineAvBuilder
	DeriveLiveRoomBuilder() InlineLiveRoomBuilder
	DeriveLiveEntryRoomBuilder() InlineLiveEntryRoomBuilder
	DerivePgcBuilder() InlinePGCBuilder
	DeriveArcPlayerV2Builder() InlineAvV2Builder
	DeriveSingleBangumiBuilder() SingleInlineBangumiBuilder
	DeriveSingleArcPlayerBuilder() SingleInlineAvBuilder
}

type largeCoverInlineBuilderFactory struct {
	jsonbuilder.BuilderContext
	asSP bool
}

// NewLargeCoverInlineBuilder is
func NewLargeCoverInlineBuilder(ctx jsonbuilder.BuilderContext) LargeCoverInlineBuilderFactory {
	return largeCoverInlineBuilderFactory{BuilderContext: ctx}
}

func (b largeCoverInlineBuilderFactory) ReplaceContext(ctx jsonbuilder.BuilderContext) LargeCoverInlineBuilderFactory {
	b.BuilderContext = ctx
	return b
}

func (b largeCoverInlineBuilderFactory) MarkAsSpecialCard(in bool) LargeCoverInlineBuilderFactory {
	b.asSP = in
	return b
}

func (b largeCoverInlineBuilderFactory) SpecialCard() bool {
	return b.asSP
}

func (b largeCoverInlineBuilderFactory) DeriveArcPlayerBuilder() InlineAvBuilder {
	if !b.asSP {
		return v6InlineAvBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b largeCoverInlineBuilderFactory) DeriveLiveRoomBuilder() InlineLiveRoomBuilder {
	if !b.asSP {
		return v8InlineLiveRoomBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b largeCoverInlineBuilderFactory) DeriveLiveEntryRoomBuilder() InlineLiveEntryRoomBuilder {
	return v8InlineLiveEntryRoomBuilder{parent: &b}
}

func (b largeCoverInlineBuilderFactory) DerivePgcBuilder() InlinePGCBuilder {
	if !b.asSP {
		return v7InlinePGCBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b largeCoverInlineBuilderFactory) DeriveArcPlayerV2Builder() InlineAvV2Builder {
	if !b.asSP {
		return v9InlineAvBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b largeCoverInlineBuilderFactory) DeriveSingleBangumiBuilder() SingleInlineBangumiBuilder {
	if !b.asSP {
		return v7SingleInlineBangumiBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b largeCoverInlineBuilderFactory) DeriveSingleArcPlayerBuilder() SingleInlineAvBuilder {
	if !b.asSP {
		return v9SingleInlineAvBuilder{parent: &b}
	}
	panic("unimpl")
}

type Inline struct {
	LikeButtonShowCount bool
	// 点赞按钮资源
	LikeResource             string
	LikeResourceHash         string
	DisLikeResource          string
	DisLikeResourceHash      string
	LikeNightResource        string
	LikeNightResourceHash    string
	DisLikeNightResource     string
	DisLikeNightResourceHash string
	IconDrag                 string
	IconDragHash             string
	IconStop                 string
	IconStopHash             string
	ThreePointPanelType      int
}

func InlineFilledByEpMaterials(arg *deliverygrpc.EpMaterial) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
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
			card.Desc = arg.Desc
		}
		if arg.GetPowerCorner().GetPowerPicSun() != "" && arg.GetPowerCorner().GetPowerPicNight() != "" &&
			arg.GetPowerCorner().GetWidth() > 0 && arg.GetPowerCorner().GetHeight() > 0 {
			card.LeftCoverBadgeNewStyle = &jsoncard.ReasonStyle{
				IconURL:      arg.GetPowerCorner().GetPowerPicSun(),
				IconURLNight: arg.GetPowerCorner().GetPowerPicNight(),
				IconWidth:    int32(math.Floor(float64(arg.GetPowerCorner().GetWidth()) / float64(arg.GetPowerCorner().GetHeight()) * float64(21))),
				IconHeight:   21,
			}
		}
	}
}

func InlineFilledByMultiMaterials(arg *tunnelV2.Material, item *ai.Item) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
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
			card.Desc = arg.Desc
		}
		if arg.GetPowerCorner().GetPowerPicSun() != "" && arg.GetPowerCorner().GetPowerPicNight() != "" &&
			arg.GetPowerCorner().GetWidth() > 0 && arg.GetPowerCorner().GetHeight() > 0 {
			card.LeftCoverBadgeNewStyle = &jsoncard.ReasonStyle{
				IconURL:      arg.GetPowerCorner().GetPowerPicSun(),
				IconURLNight: arg.GetPowerCorner().GetPowerPicNight(),
				IconWidth:    int32(math.Floor(float64(arg.GetPowerCorner().GetWidth()) / float64(arg.GetPowerCorner().GetHeight()) * float64(21))),
				IconHeight:   21,
			}
		}
	}
}

func InlineReplacedByRcmd(item *ai.Item) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		if item.CustomizedTitle != "" {
			card.Title = item.CustomizedTitle
		}
		if appcardmodel.IsValidCover(item.CustomizedCover) {
			card.Cover = item.CustomizedCover
		}
		if item.CustomizedOGVDesc != "" {
			card.Desc = item.CustomizedOGVDesc
		}
	}
}

func SingleInlineDbClickLike(item *ai.Item) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		card.EnableDoubleClickLike = item.SingleInlineDbClickLike()
	}
}

func DoubleInlineDbClickLike(item *ai.Item) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		card.EnableDoubleClickLike = item.DoubleInlineDbClickLike()
	}
}

// 订阅与banner 因单双列不一致特殊处理
func DbClickLike(ctx cardschema.FeedContext, item *ai.Item) func(*jsoncard.LargeCoverInline) {
	switch appcardmodel.Columnm[appcardmodel.ColumnStatus(ctx.IndexParam().Column())] {
	case appcardmodel.ColumnSvrSingle:
		return SingleInlineDbClickLike(item)
	default:
	}
	return func(*jsoncard.LargeCoverInline) {}
}

func LargeCoverInlineFromSpecialS(arg *operate.Card) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
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
			card.Desc = arg.Desc
		}
	}
}

func LargeCoverInlineTalkBack() func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		buffer := bytes.Buffer{}
		buffer.WriteString(appcardmodel.TalkBackCardType(card.Goto) + ",")
		buffer.WriteString(card.Title + ",")
		buffer.WriteString(card.CoverLeft1ContentDescription + ",")
		if card.CoverLeft2ContentDescription != "" {
			buffer.WriteString(card.CoverLeft2ContentDescription + ",")
		}
		if card.CoverRightContentDescription != "" {
			buffer.WriteString("时长" + card.CoverRightContentDescription + ",")
		}
		if card.Desc != "" {
			buffer.WriteString(card.Desc)
		}
		card.TalkBack = buffer.String()
	}
}
