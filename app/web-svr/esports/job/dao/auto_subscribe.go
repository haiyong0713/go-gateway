package dao

import (
	"context"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/esports/job/model"
	innerSql "go-gateway/app/web-svr/esports/job/sql"
)

const (
	sql4FetchAutoSubMids = `
SELECT id, mid
FROM auto_subscribe_season_detail_%v
WHERE team_id = %v AND is_deleted = 0
ORDER BY id asc
LIMIT %v
`
	sql4FetchAutoSubMidsWithLastID = `
SELECT id, mid
FROM auto_subscribe_season_detail_%v
WHERE team_id = %v AND is_deleted = 0 AND id > %v
ORDER BY id asc
LIMIT %v
`

	Limit4EveryQuery = 5000
)

func AutoSubMids(ctx context.Context, detail model.AutoSubscribeDetail, lastID int64) (mids []int64, newLastID int64, err error) {
	var rows *xsql.Rows
	if lastID > 0 {
		rows, err = innerSql.GlobalDB.Query(
			ctx,
			fmt.Sprintf(
				sql4FetchAutoSubMidsWithLastID,
				detail.SeasonID,
				detail.TeamId,
				lastID,
				Limit4EveryQuery))
	} else {
		rows, err = innerSql.GlobalDB.Query(
			ctx,
			fmt.Sprintf(
				sql4FetchAutoSubMids,
				detail.SeasonID,
				detail.TeamId,
				Limit4EveryQuery))
	}
	if err != nil {
		return
	}

	var id, mid int64
	mids = make([]int64, 0)
	for rows.Next() {
		scanErr := rows.Scan(&id, &mid)
		if scanErr == nil && id > 0 && mid > 0 {
			mids = append(mids, mid)
			newLastID = id
		}
	}

	err = rows.Err()
	_ = rows.Close()

	return
}
