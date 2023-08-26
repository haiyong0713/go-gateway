package adapters

import (
	"context"
	"encoding/json"
	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
	"strconv"
	"strings"
)

const (
	sql4GetOperVideo = `
SELECT data
FROM act_web_data
WHERE id = ?`
)

var OperationVideoDS = &operationVideoDataSource{}

type operationVideoDataSource struct {
}

func (m *operationVideoDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0)
	var aidStr string
	err = component.GlobalDB.QueryRow(ctx, sql4GetOperVideo, sourceId).Scan(&aidStr)
	if err != nil {
		if err == sql.ErrNoRows {
			err = xecode.Error(xecode.RequestErr, "未找到该数据源ID")
		}
		return
	}
	operC := &operConfig{}
	err = json.Unmarshal([]byte(aidStr), operC)
	if err != nil {
		return
	}
	aidsStrs := strings.Split(operC.VideoIdsStr, "\n")
	if len(aidsStrs) == 0 {
		return
	}
	var aids []int64
	for _, aidStr := range aidsStrs {
		aid, err := strconv.ParseInt(aidStr, 10, 64)
		if err == nil {
			aids = append(aids, aid)
		}
	}
	return getVoteVideoInfoByAids(ctx, aids)
}

func (m *operationVideoDataSource) NewEmptyItem() vote.DataSourceItem {
	return &video{}
}
