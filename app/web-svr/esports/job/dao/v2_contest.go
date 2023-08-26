package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/web-svr/esports/job/sql"
)

const (
	sql4HotMatchList = `
SELECT id, home_id, away_id
FROM es_contests
WHERE sid = ?
    AND id > ?
ORDER BY id ASC
LIMIT ?
`
)

func FetchAllHotMatchList(ctx context.Context, seasonIDList []int64) (m map[int64][]int64, teamIDMap map[int64]int64, err error) {
	m = make(map[int64][]int64, 0)
	teamIDMap = make(map[int64]int64, 0)
	for _, v := range seasonIDList {
		var startID int64
		tmpList := make([]int64, 0)

		for {
			d, tmpTeamIDM, tmpErr := fetchHotMatchListByStartID(ctx, v, startID, limit4HotData)
			if tmpErr != nil {
				err = tmpErr

				return
			}

			if len(d) > 0 {
				tmpList = append(tmpList, d...)
				startID = d[len(d)-1]

				for k := range tmpTeamIDM {
					teamIDMap[k] = 1
				}
			}

			if len(d) < limit4HotData {
				break
			}
		}

		m[v] = tmpList
	}

	return
}

func fetchHotMatchListByStartID(ctx context.Context, seasonID, startID, limit int64) (list []int64,
	teamIDMap map[int64]int64, err error) {
	var rows *xsql.Rows
	list = make([]int64, 0)
	teamIDMap = make(map[int64]int64, 0)
	rows, err = sql.GlobalDB.Query(ctx, sql4HotMatchList, seasonID, startID, limit)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "fetchHotMatchListByStartID rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		var tmpID, homeTeamID, awayTeamID int64
		if scanErr := rows.Scan(&tmpID, &homeTeamID, &awayTeamID); scanErr == nil && tmpID != 0 {
			list = append(list, tmpID)
			teamIDMap[homeTeamID] = 1
			teamIDMap[awayTeamID] = 1
		}
	}
	return
}
