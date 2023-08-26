package adapters

import (
	"context"
	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
)

const (
	sql4GetWidsBySource = `
SELECT wid
FROM likes
WHERE sid = ? AND state=1
ORDER BY id DESC
`
)

func getWidsBySid(ctx context.Context, sid int64) (res []int64, err error) {
	res = make([]int64, 0)
	var rows *sql.Rows
	rows, err = component.GlobalDB.Query(ctx, sql4GetWidsBySource, sid)
	if err != nil {
		if err == sql.ErrNoRows {
			err = xecode.Error(xecode.RequestErr, "未找到该数据源ID")
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		aid := int64(0)
		err = rows.Scan(&aid)
		if err != nil {
			return
		}
		res = append(res, aid)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}
