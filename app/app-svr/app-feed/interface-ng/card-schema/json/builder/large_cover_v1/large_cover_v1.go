package large_cover_v1

import (
	"math"

	"go-gateway/app/app-svr/app-card/interface/model"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
)

// LargeCoverV1BuilderFactory is
type LargeCoverV1BuilderFactory interface {
	ReplaceContext(jsonbuilder.BuilderContext) LargeCoverV1BuilderFactory

	DeriveArcPlayerBuilder() V1ArcPlayerBuilder
	DeriveEpBangumiBuilder() V1EpBangumiBuilder
	DeriveEpPGCBuilder() V1EpPGCBuilder
	DeriveLiveRoomBuilder() V1LiveRoomBuilder
	DeriveArticleBuilder() V1ArticleBuilder
}

type largeCoverV1BuilderFactory struct {
	jsonbuilder.BuilderContext
}

// NewLargeCoverV1Builder is
func NewLargeCoverV1Builder(ctx jsonbuilder.BuilderContext) LargeCoverV1BuilderFactory {
	return largeCoverV1BuilderFactory{BuilderContext: ctx}
}

func (b largeCoverV1BuilderFactory) ReplaceContext(ctx jsonbuilder.BuilderContext) LargeCoverV1BuilderFactory {
	b.BuilderContext = ctx
	return b
}

func (b largeCoverV1BuilderFactory) DeriveArcPlayerBuilder() V1ArcPlayerBuilder {
	return v1ArchiveBuilder{parent: &b}
}

func (b largeCoverV1BuilderFactory) DeriveEpBangumiBuilder() V1EpBangumiBuilder {
	return v1EpBangumiBuilder{parent: &b}
}

func (b largeCoverV1BuilderFactory) DeriveEpPGCBuilder() V1EpPGCBuilder {
	return v1EpPGCBuilder{parent: &b}
}

func (b largeCoverV1BuilderFactory) DeriveLiveRoomBuilder() V1LiveRoomBuilder {
	return v1LiveRoomBuilder{parent: &b}
}

func (b largeCoverV1BuilderFactory) DeriveArticleBuilder() V1ArticleBuilder {
	return v1ArticleBuilder{parent: &b}
}

func V1FilledByMultiMaterials(arg *tunnelV2.Material, item *ai.Item, needGif bool) func(v1 *jsoncard.LargeCoverV1) {
	return func(card *jsoncard.LargeCoverV1) {
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
		if arg.Desc != "" && item.Goto != string(model.CardGotoBangumi) {
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

func V1ReplacedByRcmd(item *ai.Item) func(*jsoncard.LargeCoverV1) {
	return func(card *jsoncard.LargeCoverV1) {
		if item.CustomizedTitle != "" {
			card.Title = item.CustomizedTitle
		}
		if appcardmodel.IsValidCover(item.CustomizedCover) {
			card.Cover = item.CustomizedCover
		}
	}
}
