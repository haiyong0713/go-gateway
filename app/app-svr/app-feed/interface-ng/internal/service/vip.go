package service

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
)

type vip struct{}

func (vip) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildSmallCoverV7FromVip(ctx, index, item, fanoutResult)
}
