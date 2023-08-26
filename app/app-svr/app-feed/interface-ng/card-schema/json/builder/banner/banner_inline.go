package jsonbanner

import (
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"

	"github.com/pkg/errors"
)

type InlineBannerBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) InlineBannerBuilder
	SetRcmd(*ai.Item) InlineBannerBuilder
	SetBase(*jsoncard.Base) InlineBannerBuilder
	SetBanners([]*banner.Banner) InlineBannerBuilder
	SetVersion(string) InlineBannerBuilder
	SetArcPlayer(map[int64]*arcgrpc.ArcPlayer) InlineBannerBuilder
	SetEpisode(map[int32]*pgcinline.EpisodeCard) InlineBannerBuilder
	SetLiveRoom(map[int64]*live.Room) InlineBannerBuilder
	SetInline(*jsonlargecover.Inline) InlineBannerBuilder
	SetAuthorCard(map[int64]*accountgrpc.Card) InlineBannerBuilder

	Build() (*jsoncard.BannerV8, error)
}

type inlineBannerBuilder struct {
	jsonbuilder.BuilderContext
	base       *jsoncard.Base
	banners    []*banner.Banner
	version    string
	rcmd       *ai.Item
	arcPlayer  map[int64]*arcgrpc.ArcPlayer
	episode    map[int32]*pgcinline.EpisodeCard
	liveRoom   map[int64]*live.Room
	authorCard map[int64]*accountgrpc.Card
	inline     *jsonlargecover.Inline
}

func NewBannerInlineBuilder(ctx jsonbuilder.BuilderContext) InlineBannerBuilder {
	return inlineBannerBuilder{BuilderContext: ctx}
}

func (b inlineBannerBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) InlineBannerBuilder {
	b.BuilderContext = ctx
	return b
}

func (b inlineBannerBuilder) SetBase(base *jsoncard.Base) InlineBannerBuilder {
	b.base = base
	return b
}

func (b inlineBannerBuilder) SetBanners(banners []*banner.Banner) InlineBannerBuilder {
	b.banners = banners
	return b
}

func (b inlineBannerBuilder) SetVersion(in string) InlineBannerBuilder {
	b.version = in
	return b
}

func (b inlineBannerBuilder) SetRcmd(item *ai.Item) InlineBannerBuilder {
	b.rcmd = item
	return b
}

func (b inlineBannerBuilder) SetArcPlayer(arcs map[int64]*arcgrpc.ArcPlayer) InlineBannerBuilder {
	b.arcPlayer = arcs
	return b
}

func (b inlineBannerBuilder) SetEpisode(in map[int32]*pgcinline.EpisodeCard) InlineBannerBuilder {
	b.episode = in
	return b
}

func (b inlineBannerBuilder) SetLiveRoom(in map[int64]*live.Room) InlineBannerBuilder {
	b.liveRoom = in
	return b
}

func (b inlineBannerBuilder) SetInline(in *jsonlargecover.Inline) InlineBannerBuilder {
	b.inline = in
	return b
}

func (b inlineBannerBuilder) SetAuthorCard(in map[int64]*accountgrpc.Card) InlineBannerBuilder {
	b.authorCard = in
	return b
}

func (b inlineBannerBuilder) Build() (*jsoncard.BannerV8, error) {
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.rcmd.BannerInfo == nil {
		return nil, errors.Errorf("empty `BannerInfo`, trackid: %s", b.rcmd.TrackID)
	}
	if len(b.banners) == 0 {
		return nil, errors.Errorf("empty `Banners` field")
	}
	bannerItems := make([]*jsoncard.BannerItem, 0, len(b.banners))
	for _, v := range b.banners {
		bannerItem := &jsoncard.BannerItem{
			ResourceID: int64(v.ResourceID),
			ID:         v.ID,
			Index:      int64(v.Index),
		}
		switch {
		case card.IsAdBanner(v):
			bannerItem.Type = card.BannerTypeAd
			bannerItem.AdBanner = v
		case card.IsAdInlineBanner(v):
			bannerItem.Type = card.BannerTypeAdInline
			// 低版本降级成ad类型
			if !b.BuilderContext.VersionControl().Can("banner.enableAdInline") {
				bannerItem.Type = card.BannerTypeAd
			}
			bannerItem.AdBanner = v
		case card.IsStaticBanner(v):
			bannerItem.Type = card.BannerTypeStatic
			bannerItem.StaticBanner = v
		case card.IsInlineBanner(v):
			if err := b.setInlineBannerItem(v, bannerItem); err != nil {
				log.Error("Failed to set inline banner item: %+v", err)
				bannerItem.FailbackToStatic(v)
			}
		}
		bannerItems = append(bannerItems, bannerItem)
	}
	out := &jsoncard.BannerV8{
		Base:       b.base,
		Hash:       b.version,
		BannerItem: bannerItems,
	}
	return out, nil
}

func (b inlineBannerBuilder) buildInlineAv(id int64) (*jsoncard.LargeCoverInline, error) {
	archive, ok := b.arcPlayer[id]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(b.BuilderContext).SetParam(strconv.FormatInt(id, 10)).
		SetMetricRcmd(b.rcmd).SetTrackID(b.rcmd.TrackID).Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(b.BuilderContext)
	card, err := factory.DeriveArcPlayerBuilder().
		SetBase(base).
		SetRcmd(b.rcmd).
		SetArcPlayer(archive).
		SetAuthorCard(b.authorCard[archive.Arc.Author.Mid]).
		SetInline(b.inline).
		WithAfter(jsonlargecover.DbClickLike(b.BuilderContext, b.rcmd)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (b inlineBannerBuilder) buildInlinePGC(id int64) (*jsoncard.LargeCoverInline, error) {
	epCard, ok := b.episode[int32(id)]
	if !ok {
		return nil, errors.Errorf("ep card not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(b.BuilderContext).SetParam(strconv.FormatInt(id, 10)).
		SetMetricRcmd(b.rcmd).SetTrackID(b.rcmd.TrackID).Build()
	if err != nil {
		return nil, err
	}
	factory := jsonlargecover.NewLargeCoverInlineBuilder(b.BuilderContext)
	card, err := factory.DerivePgcBuilder().
		SetBase(base).
		SetRcmd(b.rcmd).
		SetEpisode(epCard).
		SetInline(b.inline).
		WithAfter(jsonlargecover.DbClickLike(b.BuilderContext, b.rcmd)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (b inlineBannerBuilder) buildInlineLive(id int64) (*jsoncard.LargeCoverInline, error) {
	live, ok := b.liveRoom[id]
	if !ok {
		return nil, errors.Errorf("live room not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(b.BuilderContext).SetParam(strconv.FormatInt(id, 10)).
		SetMetricRcmd(b.rcmd).SetTrackID(b.rcmd.TrackID).Build()
	if err != nil {
		return nil, err
	}
	factory := jsonlargecover.NewLargeCoverInlineBuilder(b.BuilderContext)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(b.rcmd).
		SetLiveRoom(live).
		SetAuthorCard(b.authorCard[live.UID]).
		SetInline(b.inline).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (b inlineBannerBuilder) setInlineBannerItem(banner *banner.Banner, item *jsoncard.BannerItem) error {
	id, err := strconv.ParseInt(banner.BannerMeta.InlineId, 10, 64)
	if err != nil {
		return errors.WithStack(err)
	}
	switch banner.BannerMeta.InlineType {
	case card.InlineTypeAv:
		inlineAv, err := b.buildInlineAv(id)
		if err != nil {
			return err
		}
		item.Type = card.BannerTypeInlineAv
		item.InlineAv = &jsoncard.EmbedBannerInline{
			EmbedInline: jsoncard.EmbedInline{
				LargeCoverInline: *inlineAv,
			},
		}
		card.BannerHide(item.InlineAv)
		card.SetBannerMeta(item.InlineAv, banner)
		b.setInlineMeta(item.InlineAv, banner)
	case card.InlineTypePGC:
		inlinePGC, err := b.buildInlinePGC(id)
		if err != nil {
			return err
		}
		item.Type = card.BannerTypeInlinePGC
		item.InlinePGC = &jsoncard.EmbedBannerInline{
			EmbedInline: jsoncard.EmbedInline{
				LargeCoverInline: *inlinePGC,
			},
		}
		card.BannerHide(item.InlinePGC)
		card.SetBannerMeta(item.InlinePGC, banner)
		b.setInlineMeta(item.InlinePGC, banner)
	case card.InlineTypeLive:
		inlineLive, err := b.buildInlineLive(id)
		if err != nil {
			return err
		}
		item.Type = card.BannerTypeInlineLive
		item.InlineLive = &jsoncard.EmbedBannerInline{
			EmbedInline: jsoncard.EmbedInline{
				LargeCoverInline: *inlineLive,
			},
		}
		card.BannerHide(item.InlineLive)
		card.SetBannerMeta(item.InlineLive, banner)
		b.setInlineMeta(item.InlineLive, banner)
	default:
		return errors.Errorf("Unrecognized inline type: %s", banner.BannerMeta.InlineType)
	}
	return nil
}

func (b inlineBannerBuilder) setInlineMeta(embedBannerInline *card.EmbedBannerInline, banner *banner.Banner) {
	embedBannerInline.ExtraURI = card.ExtraURIFromBanner(banner)
	embedBannerInline.Title = banner.Title
	embedBannerInline.Cover = banner.Image
	embedBannerInline.HideDanmuSwitch = card.AsBannerInlineDanmu(banner.InlineBarrageSwitch)
	embedBannerInline.DisableDanmu = card.AsBannerInlineDanmu(banner.InlineBarrageSwitch)
	embedBannerInline.ThreePointV2 = nil
	embedBannerInline.ThreePoint = nil
}
