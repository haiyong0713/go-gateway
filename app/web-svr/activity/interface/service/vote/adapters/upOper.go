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
	sql4GetOperUp = `
SELECT data
FROM act_web_data
WHERE id = ?`
)

var OperationUpDS = &operationUpDataSource{}

type operUpConfig struct {
	MidStr string `json:"mids"`
}

type operationUpDataSource struct {
}

func (m *operationUpDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0)
	var midStr string
	err = component.GlobalDB.QueryRow(ctx, sql4GetOperUp, sourceId).Scan(&midStr)
	if err != nil {
		if err == sql.ErrNoRows {
			err = xecode.Error(xecode.RequestErr, "未找到该数据源ID")
		}
		return
	}
	operC := &operUpConfig{}
	err = json.Unmarshal([]byte(midStr), operC)
	if err != nil {
		return
	}
	midsStrs := strings.Split(operC.MidStr, "\n")
	if len(midsStrs) == 0 {
		return
	}
	var mids []int64
	for _, midStr := range midsStrs {
		aid, err := strconv.ParseInt(midStr, 10, 64)
		if err == nil {
			mids = append(mids, aid)
		}
	}
	res, err = getVoteUpInfoByMids(ctx, mids)
	return
}

func (m *operationUpDataSource) NewEmptyItem() vote.DataSourceItem {
	return &up{}
}
