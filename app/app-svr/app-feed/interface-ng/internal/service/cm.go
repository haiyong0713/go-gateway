package service

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
)

type adAv struct{}

func (adAv) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV2AdAv(ctx, index, item, fanoutResult)
}

type adWebS struct{}

func (adWebS) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV2AdWebS(ctx, index, item, fanoutResult)
}

type adWeb struct{}

func (adWeb) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV2AdWeb(ctx, index, item, fanoutResult)
}

type adPlayer struct{}

func (adPlayer) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV2AdPlayer(ctx, index, item, fanoutResult)
}

type adInlineLive struct{}

func (adInlineLive) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV2AdInlineLive(ctx, item, fanoutResult)
}

type adWebSSingle struct{}

func (adWebSSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV1AdWebS(ctx, index, item, fanoutResult)
}

type adWebSingle struct{}

func (adWebSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV1AdWeb(ctx, index, item, fanoutResult)
}

type adAvSingle struct{}

func (adAvSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildCmV1AdAv(ctx, index, item, fanoutResult)
}
