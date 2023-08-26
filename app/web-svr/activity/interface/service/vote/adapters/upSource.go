package adapters

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
)

var UpSourceDS = &upSourceDataSource{}

type upSourceDataSource struct {
}

func (m *upSourceDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0)
	mids, err := getWidsBySid(ctx, sourceId)
	if err != nil {
		return
	}
	return getVoteUpInfoByMids(ctx, mids)
}

func (m *upSourceDataSource) NewEmptyItem() vote.DataSourceItem {
	return &up{}
}
