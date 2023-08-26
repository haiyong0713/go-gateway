package service

import (
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"

	"github.com/pkg/errors"
)

type av struct{}

func (av) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	_, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}
	return feedcard.BuildSmallCoverV2FromArchive(ctx, index, item, fanoutResult)
}

type avSingle struct{}

func (avSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	_, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}
	if item.SingleInline == model.SingleInlineV1 {
		return feedcard.BuildLargeCoverSingleV9(ctx, index, item, fanoutResult)
	}
	return feedcard.BuildLargeCoverV1FromArchive(ctx, index, item, fanoutResult)
}
