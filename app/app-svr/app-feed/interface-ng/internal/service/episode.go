package service

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
)

type episodePGC struct{}

func (s episodePGC) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildSmallCoverV2FromPGC(ctx, index, item, fanoutResult)
}

type episodeBangumi struct{}

func (s episodeBangumi) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildSmallCoverV2FromBangumi(ctx, index, item, fanoutResult)
}

type bangumiRcmd struct{}

func (s bangumiRcmd) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildSmallCoverV4FromBangumiRcmd(ctx, index, item, fanoutResult)
}

type episodeBangumiSingle struct{}

func (episodeBangumiSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildLargeCoverV1FromBangumi(ctx, index, item, fanoutResult)
}

type episodePGCSingle struct{}

func (s episodePGCSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildLargeCoverV1FromPGC(ctx, index, item, fanoutResult)
}

type ogv struct{}

func (ogv) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	return feedcard.BuildLargeCoverSingleV7(ctx, index, item, fanoutResult)
}
