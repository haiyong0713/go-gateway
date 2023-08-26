package service

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
)

//nolint:unused
type tunnel struct{}

//nolint:unused
func (tunnel) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildSmallCoverV4FromTunnel(ctx, index, item, fanoutResult)
}
