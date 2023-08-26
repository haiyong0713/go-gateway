package service

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
)

type inlineAv struct{}

func (inlineAv) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	if ctx.VersionControl().Can("feed.usingInline2") {
		return feedcard.BuildLargeCoverV6FromArchive(ctx, index, item, fanoutResult)
	}
	return feedcard.BuildLargeCoverV5FromArchive(ctx, index, item, fanoutResult)
}

type inlineAvV2 struct{}

func (inlineAvV2) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildLargeCoverV9FromArchive(ctx, index, item, fanoutResult)
}
