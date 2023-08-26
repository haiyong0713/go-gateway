package service

import (
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"

	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"

	"github.com/pkg/errors"
)

type picture struct{}

func (picture) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	//nolint:gomnd
	if len(picture.Imgs) < 3 {
		if ctx.VersionControl().Can("feed.usingOnePicV3") {
			return feedcard.BuildOnePicV3FromPicture(ctx, index, item, fanoutResult)
		}
		if ctx.VersionControl().Can("feed.usingOnePicV2") {
			return feedcard.BuildOnePicV2FromPicture(ctx, index, item, fanoutResult)
		}
		return feedcard.BuildSmallCoverV2FromPicture(ctx, index, item, fanoutResult)
	}
	if ctx.VersionControl().Can("feed.usingThreePicV3") {
		return feedcard.BuildThreePicV3FromPicture(ctx, index, item, fanoutResult)
	}
	return feedcard.BuildThreePicV2FromPicture(ctx, index, item, fanoutResult)
}

type pictureSingle struct{}

func (pictureSingle) Build(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *feedcard.FanoutResult) (cardschema.FeedCard, error) {
	picture, ok := fanoutResult.Dynamic.Picture[item.ID]
	if !ok {
		return nil, errors.Errorf("picture not exist")
	}
	//nolint:gomnd
	if len(picture.Imgs) < 3 {
		return feedcard.BuildOnePicV1FromPicture(ctx, index, item, fanoutResult)
	}
	return feedcard.BuildThreePicV1FromPicture(ctx, index, item, fanoutResult)
}
