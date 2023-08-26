package service

import (
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
)

type liveRoom struct{}

func (liveRoom) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildSmallCoverV2FromLiveRoom(ctx, index, item, fanoutResult)
}

type liveRoomSingle struct{}

func (liveRoomSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleInline == model.SingleInlineV1 {
		return feedcard.BuildLargeCoverSingleV8(ctx, index, item, fanoutResult)
	}
	return feedcard.BuildLargeCoverV1FromLiveRoom(ctx, index, item, fanoutResult)
}
