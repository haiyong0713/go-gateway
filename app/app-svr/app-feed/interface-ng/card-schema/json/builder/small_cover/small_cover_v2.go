package jsonsmallcover

import (
	"bytes"
	"math"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	feedcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	resourceV2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
)

// SmallCoverV2BuilderFactory is
type SmallCoverV2BuilderFactory interface {
	ReplaceContext(jsonbuilder.BuilderContext) SmallCoverV2BuilderFactory
	MarkAsSpecialCard(bool) SmallCoverV2BuilderFactory
	SpecialCard() bool

	DeriveArcPlayerBuilder() V2ArcPlayerBuilder
	DeriveBangumiSeasonBuilder() V2BangumiSeasonBuilder
	DerivePictureBuilder() V2PictureBuilder
	DeriveArticleBuilder() V2ArticleBuilder
	DeriveLiveRoomBuilder() V2LiveRoomBuilder
	DeriveEpBangumiBuilder() V2EpBangumiBuilder
	DeriveEpPGCBuilder() V2EpPGCBuilder
	DeriveSpecialSeasonBuilder() V2SpecialSeasonBuilder
	DeriveWebBuilder() V2WebBuilder
}

type smallCoverV2BuilderFactory struct {
	jsonbuilder.BuilderContext
	asSP bool
}

// NewSmallCoverV2Builder is
func NewSmallCoverV2Builder(ctx jsonbuilder.BuilderContext) SmallCoverV2BuilderFactory {
	return smallCoverV2BuilderFactory{BuilderContext: ctx}
}

func (b smallCoverV2BuilderFactory) ReplaceContext(ctx jsonbuilder.BuilderContext) SmallCoverV2BuilderFactory {
	b.BuilderContext = ctx
	return b
}

func (b smallCoverV2BuilderFactory) MarkAsSpecialCard(in bool) SmallCoverV2BuilderFactory {
	b.asSP = in
	return b
}

func (b smallCoverV2BuilderFactory) SpecialCard() bool {
	return b.asSP
}

func (b smallCoverV2BuilderFactory) DeriveArcPlayerBuilder() V2ArcPlayerBuilder {
	if !b.asSP {
		return v2ArchiveBuilder{parent: &b}
	}
	panic("unimpl")
}
func (b smallCoverV2BuilderFactory) DeriveBangumiSeasonBuilder() V2BangumiSeasonBuilder {
	if !b.asSP {
		return v2BangumiSeasonBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b smallCoverV2BuilderFactory) DerivePictureBuilder() V2PictureBuilder {
	if !b.asSP {
		return v2PictureBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b smallCoverV2BuilderFactory) DeriveArticleBuilder() V2ArticleBuilder {
	if !b.asSP {
		return v2ArticleBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b smallCoverV2BuilderFactory) DeriveLiveRoomBuilder() V2LiveRoomBuilder {
	if !b.asSP {
		return v2LiveRoomBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b smallCoverV2BuilderFactory) DeriveEpBangumiBuilder() V2EpBangumiBuilder {
	if !b.asSP {
		return v2EpBangumiBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b smallCoverV2BuilderFactory) DeriveEpPGCBuilder() V2EpPGCBuilder {
	if !b.asSP {
		return v2EpPGCBuilder{parent: &b}
	}
	panic("unimpl")
}

func (b smallCoverV2BuilderFactory) DeriveSpecialSeasonBuilder() V2SpecialSeasonBuilder {
	return v2SpecialSeasonBuilder{parent: &b}
}

func (b smallCoverV2BuilderFactory) DeriveWebBuilder() V2WebBuilder {
	return v2WebBuilder{parent: &b}
}

func SmallCoverV2FromSpecial(ctx cardschema.FeedContext, asc *resourceV2grpc.AppSpecialCard, item *ai.Item) func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		if asc.Title != "" {
			card.Title = asc.Title
		}
		if asc.Cover != "" {
			card.Cover = asc.Cover
		}
		if asc.Gifcover != "" && item.StaticCover == 0 {
			card.CoverGif = asc.Gifcover
		}
		if asc.Corner != "" {
			card.Badge = asc.Corner
			card.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, asc.Corner)
		}
		if asc.ReType == 0 || asc.ReType == 1 {
			card.URI = appcardmodel.FillURI(appcardmodel.OperateType[int(asc.ReType)], ctx.Device().Plat(),
				int(ctx.Device().Build()), asc.ReValue, nil)
		}
		if item.RcmdReason != nil {
			card.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
				item.RcmdReason.Content,
				jsonreasonstyle.CornerMarkFromAI(item),
				jsonreasonstyle.CorverMarkFromContext(ctx),
			)
		}
		if card.RcmdReasonStyle == nil && asc.Desc != "" {
			card.Desc = asc.Desc
		}
		if asc.PowerPicSun != "" && asc.PowerPicNight != "" && asc.Width > 0 && asc.Height > 0 {
			card.LeftCoverBadgeNewStyle = &jsoncard.ReasonStyle{
				IconURL:      asc.PowerPicSun,
				IconURLNight: asc.PowerPicNight,
				IconWidth:    int32(math.Floor(float64(asc.Width) / float64(asc.Height) * float64(21))),
				IconHeight:   21,
			}
		}
	}
}

func V2FilledByEpMaterials(arg *deliverygrpc.EpMaterial, item *ai.Item) func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
		if item.AllowGIF() && arg.GifCover != "" && item.StaticCover == 0 {
			card.CoverGif = arg.GifCover
		}
		if arg.Desc != "" {
			if card.DescButton != nil {
				card.DescButton.Text = arg.Desc
			} else {
				card.Desc = arg.Desc
			}
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

func V2FilledByMultiMaterials(arg *tunnelV2.Material, item *ai.Item, needGif bool) func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
		if needGif && item.AllowGIF() && arg.GifCover != "" && item.StaticCover == 0 {
			card.CoverGif = arg.GifCover
		}
		if arg.Desc != "" {
			if (item.Goto == "pgc" || item.Goto == "bangumi" || item.Goto == "av") && card.DescButton != nil {
				card.DescButton.Text = arg.Desc
			} else {
				card.Desc = arg.Desc
			}
			if item.Goto == "av" {
				card.Desc = arg.Desc
				card.RcmdReason = ""
				card.RcmdReasonStyleV2 = nil
				card.RcmdReasonStyle = nil
				if item.IconType != appcardmodel.AIStoryIconType {
					card.GotoIcon = nil
				}
			}
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

func V2ReplacedByRcmd(item *ai.Item) func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		if item.CustomizedTitle != "" {
			card.Title = item.CustomizedTitle
		}
		if appcardmodel.IsValidCover(item.CustomizedCover) {
			card.Cover = item.CustomizedCover
		}
		if item.CustomizedOGVDesc != "" && card.DescButton != nil {
			card.DescButton.Text = item.CustomizedOGVDesc
		}
	}
}

func SmallCoverV2AVCustomizedQuality(item *ai.Item, archive *arcgrpc.ArcPlayer) func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		if card.CardGoto != appcardmodel.CardGotoAv {
			return
		}
		cq := feedcard.CastArchiveCustomizedQuality(item, archive)
		cqCount := len(cq)
		//nolint:gomnd
		switch cqCount {
		case 1:
			card.CoverLeftText1 = cq[0].Text
			card.CoverLeftIcon1 = cq[0].Icon
			card.CoverLeftText2 = ""
			card.CoverLeftIcon2 = 0
		case 2:
			card.CoverLeftText1 = cq[0].Text
			card.CoverLeftIcon1 = cq[0].Icon
			card.CoverLeftText2 = cq[1].Text
			card.CoverLeftIcon2 = cq[1].Icon
		}
		card.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(card.CoverLeftIcon1,
			card.CoverLeftText1)
		card.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(card.CoverLeftIcon2,
			card.CoverLeftText2)
		if item.HideDuration == 1 {
			card.CoverRightText = ""
			card.CoverRightContentDescription = ""
		}
	}
}

func SmallCoverV2AVCustomizedDesc(item *ai.Item, archive *arcgrpc.ArcPlayer, tag *taggrpc.Tag, accountCard *accountgrpc.Card) func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		if card.RcmdReason != "" {
			return
		}
		if card.CardGoto != appcardmodel.CardGotoAv {
			return
		}
		if item.IconType == appcardmodel.AIUpIconType {
			upName := accountCard.GetName()
			if archive.Arc.Author.Name != "" {
				upName = archive.Arc.Author.Name
			}
			card.DescButton = &jsoncard.Button{
				Type:    appcardmodel.ButtonGrey,
				Text:    upName,
				URI:     appcardmodel.FillURI(appcardmodel.GotoMid, 0, 0, strconv.FormatInt(archive.Arc.Author.Mid, 10), nil),
				Event:   appcardmodel.EventChannelClick,
				EventV2: appcardmodel.EventV2ChannelClick,
			}
		}
		if tag == nil {
			return
		}
		customizedBtn, ok := feedcard.CastArchiveCustomizedDesc(item, archive, tag)
		if !ok {
			return
		}
		card.DescButton = &jsoncard.Button{
			Type:    appcardmodel.ButtonGrey,
			Text:    customizedBtn.Text,
			URI:     customizedBtn.URI,
			Event:   appcardmodel.EventChannelClick,
			EventV2: appcardmodel.EventV2ChannelClick,
		}
	}
}

func SmallCoverV2TalkBack() func(*jsoncard.SmallCoverV2) {
	return func(card *jsoncard.SmallCoverV2) {
		buffer := bytes.Buffer{}
		buffer.WriteString(appcardmodel.TalkBackCardType(card.Goto) + ",")
		buffer.WriteString(card.Title + ",")
		buffer.WriteString(card.CoverLeft1ContentDescription + ",")
		buffer.WriteString(card.CoverLeft2ContentDescription + ",")
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
