package dao

import (
	"context"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/job/sql"
)

const (
	sql4FetchHotTeamList = `
SELECT id, title, logo
FROM es_teams
WHERE id IN (%v)
`
)

func FetchTeamInfoByLargeIDList(ctx context.Context, teamIDList []int64) (m map[int64]*v1.Team, err error) {
	m = make(map[int64]*v1.Team)
	listLen := len(teamIDList)
	listLenAfterSplit := len(teamIDList) / limit4HotData
	if d := len(teamIDList) % limit4HotData; d > 0 {
		listLenAfterSplit++
	}
	for i := 0; i < listLenAfterSplit; i++ {
		startIndex := limit4HotData * i
		endIndex := limit4HotData*i + limit4HotData
		if endIndex > listLen {
			endIndex = listLen
		}

		tmpM, tmpErr := FetchTeamInfoBySmallIDList(ctx, teamIDList[startIndex:endIndex])
		if tmpErr != nil {
			err = tmpErr

			return
		}

		for k, v := range tmpM {
			m[k] = v
		}
	}

	return
}

func FetchTeamInfoBySmallIDList(ctx context.Context, teamIDList []int64) (m map[int64]*v1.Team, err error) {
	var rows *xsql.Rows
	m = make(map[int64]*v1.Team, 0)
	rows, err = sql.GlobalDB.Query(ctx, fmt.Sprintf(sql4FetchHotTeamList, xstr.JoinInts(teamIDList)))
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "FetchTeamInfoBySmallIDList rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		var (
			tmpID       int64
			title, logo string
		)
		if scanErr := rows.Scan(&tmpID, &title, &logo); scanErr == nil && tmpID != 0 {
			tmpTeam := new(v1.Team)
			{
				tmpTeam.ID = tmpID
				tmpTeam.Title = title
				tmpTeam.Logo = logo
			}

			m[tmpID] = tmpTeam
		}
	}
	return
}
