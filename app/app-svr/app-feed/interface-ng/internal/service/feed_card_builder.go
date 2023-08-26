package service

import (
	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"

	"github.com/pkg/errors"
)

var (
	GlobalCardBuilderResolver = &cardBuilderResolver{
		resolver: map[string]map[string]feedcard.CardBuilder{},
	}
	columnMap = map[int8]string{
		int8(appcardmodel.ColumnDefault):    _doubleColumn,
		int8(appcardmodel.ColumnSvrSingle):  _singleColumn,
		int8(appcardmodel.ColumnUserSingle): _singleColumn,
		int8(appcardmodel.ColumnSvrDouble):  _doubleColumn,
		int8(appcardmodel.ColumnUserDouble): _doubleColumn,
	}
)

const (
	_singleColumn = "single"
	_doubleColumn = "double"
	_ipadColumn   = "ipad"
)

func init() {
	GlobalCardBuilderResolver.register("av", _doubleColumn, av{})
	GlobalCardBuilderResolver.register("av", _singleColumn, avSingle{})
	GlobalCardBuilderResolver.register("picture", _doubleColumn, picture{})
	GlobalCardBuilderResolver.register("article_s", _doubleColumn, article{})
	GlobalCardBuilderResolver.register("article_s", _singleColumn, articleSingle{})
	GlobalCardBuilderResolver.register("live", _doubleColumn, liveRoom{})
	GlobalCardBuilderResolver.register("live", _singleColumn, liveRoomSingle{})
	GlobalCardBuilderResolver.register("banner", _doubleColumn, banner{})
	GlobalCardBuilderResolver.register("banner", _singleColumn, bannerSingle{})
	GlobalCardBuilderResolver.register("banner", _ipadColumn, bannerIPad{})
	GlobalCardBuilderResolver.register("inline_av", _doubleColumn, inlineAv{})
	GlobalCardBuilderResolver.register("inline_av_v2", _doubleColumn, inlineAvV2{})
	GlobalCardBuilderResolver.register("inline_live", _doubleColumn, inlineLiveRoom{})
	GlobalCardBuilderResolver.register("inline_pgc", _doubleColumn, inlinePgc{})
	GlobalCardBuilderResolver.register("bangumi", _doubleColumn, episodeBangumi{})
	GlobalCardBuilderResolver.register("bangumi", _singleColumn, episodeBangumiSingle{})
	GlobalCardBuilderResolver.register("pgc", _doubleColumn, episodePGC{})
	GlobalCardBuilderResolver.register("pgc", _singleColumn, episodePGCSingle{})
	GlobalCardBuilderResolver.register("bangumi_rcmd", _doubleColumn, bangumiRcmd{})
	GlobalCardBuilderResolver.register("tunnel", _doubleColumn, v7Tunnel{})
	GlobalCardBuilderResolver.register("follow_mode", _doubleColumn, selectV2{})
	GlobalCardBuilderResolver.register("follow_mode", _singleColumn, selectV2{})
	GlobalCardBuilderResolver.register("vip_renew", _doubleColumn, vip{})
	GlobalCardBuilderResolver.register("ai_story", _doubleColumn, story{})
	GlobalCardBuilderResolver.register("ad_av", _doubleColumn, adAv{})
	GlobalCardBuilderResolver.register("ad_web_s", _doubleColumn, adWebS{})
	GlobalCardBuilderResolver.register("ad_web", _doubleColumn, adWeb{})
	GlobalCardBuilderResolver.register("ad_player", _doubleColumn, adPlayer{})
	GlobalCardBuilderResolver.register("ad_inline_live", _doubleColumn, adInlineLive{})
	GlobalCardBuilderResolver.register("ai_story", _singleColumn, storySingle{})
	GlobalCardBuilderResolver.register("picture", _singleColumn, pictureSingle{})
	GlobalCardBuilderResolver.register("ad_av", _singleColumn, adAvSingle{})
	GlobalCardBuilderResolver.register("ad_web_s", _singleColumn, adWebSSingle{})
	GlobalCardBuilderResolver.register("ad_web", _singleColumn, adWebSingle{})
	GlobalCardBuilderResolver.register("ogv", _singleColumn, ogv{})
}

type cardBuilderResolver struct {
	resolver map[string]map[string]feedcard.CardBuilder
}

//nolint:unused
func (cbr *cardBuilderResolver) registerSingle(gotoType string, builder feedcard.CardBuilder) {
	cbr.register(gotoType, _singleColumn, builder)
}

func (cbr *cardBuilderResolver) register(gotoType string, columnType string, builder feedcard.CardBuilder) {
	builderMap, ok := cbr.resolver[columnType]
	if !ok {
		cbr.resolver[columnType] = map[string]feedcard.CardBuilder{gotoType: builder}
		return
	}
	if _, ok := builderMap[gotoType]; ok {
		panic(errors.Errorf("builder conflicated: %q", gotoType))
	}
	builderMap[gotoType] = builder
}

func (cbr cardBuilderResolver) getBuilder(ctx cardschema.FeedContext, gotoType string) (feedcard.CardBuilder, bool) {
	if ctx.VersionControl().Can("feed.usingIpadColumn") {
		builderMap, ok := cbr.resolver[_ipadColumn]
		if !ok {
			return nil, false
		}
		builder, ok := builderMap[gotoType]
		if !ok {
			return nil, false
		}
		return builder, true
	}
	builderMap, ok := cbr.resolver[columnMap[ctx.IndexParam().Column()]]
	if !ok {
		return nil, false
	}
	builder, ok := builderMap[gotoType]
	if !ok {
		return nil, false
	}
	return builder, true
}

func setFinallyIdx(ctx cardschema.FeedContext, cards []cardschema.FeedCard) {
	start := ctx.IndexParam().Idx()
	if start < 1 {
		start = ctx.AtTime().Unix()
	}
	cardSize := int64(len(cards))
	for i, card := range cards {
		index := start - (int64(i) + 1)
		if ctx.IndexParam().Pull() {
			index = start + (cardSize - int64(i))
		}
		if err := jsonbuilder.NewBaseUpdater(ctx, card.Get()).
			UpdateIndex(index).
			Update(); err != nil {
			log.Error("Failed to update base `idx`: %+v", err)
			continue
		}
	}
}
