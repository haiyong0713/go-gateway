package adapters

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
)

var VideoSourceDS = &videoSourceDataSource{}

type videoSourceDataSource struct {
}

func (m *videoSourceDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0)
	aids, err := getWidsBySid(ctx, sourceId)
	if err != nil {
		return
	}
	return getVoteVideoInfoByAids(ctx, aids)
}

func (m *videoSourceDataSource) NewEmptyItem() vote.DataSourceItem {
	return &video{}
}
