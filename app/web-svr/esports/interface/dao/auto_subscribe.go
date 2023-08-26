package dao

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/model"
)

const (
	ContestStatusNotStart = iota + 1
)

const (
	sql4CountAutoSubscribe = `
SELECT team_id
FROM auto_subscribe_season_detail_%v FORCE INDEX (ix_mtId)
WHERE mid = %v
    AND is_deleted = 0
`

	sql4AutoSubscribeDetail = `
INSERT INTO auto_subscribe_season_detail_%v(mid, team_id)
%v
ON DUPLICATE KEY UPDATE mtime = now()
`

	sql4AutoSubscribeSeasons = `
SELECT season_id
FROM auto_subscribe_seasons
WHERE is_deleted = 0
`

	sql4SeasonTeamContestList = `
SELECT id, sid, home_id, away_id,contest_status,live_room
FROM es_contests
WHERE sid in (%v) AND status = 0
`
)

func AutoSubscribeDetail(ctx context.Context, mid int64, req *model.AutoSubRequest) (err error) {
	values := ""
	for _, v := range req.TeamIDList {
		if values == "" {
			values = fmt.Sprintf("VALUES(%v, %v)", mid, v)
		} else {
			tmpValue := fmt.Sprintf("(%v, %v)", mid, v)
			values = fmt.Sprintf("%v, %v", values, tmpValue)
		}
	}

	sql := fmt.Sprintf(sql4AutoSubscribeDetail, req.SeasonID, values)
	_, err = component.GlobalDBOfMaster.Exec(ctx, sql)

	return
}

func FetchAutoSubDetail(ctx context.Context, mid int64, req *model.AutoSubRequest) (m map[int64]bool, err error) {
	sql := fmt.Sprintf(sql4CountAutoSubscribe, req.SeasonID, mid)
	rows, err := component.GlobalDB.Query(ctx, sql)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		log.Errorc(ctx, "FetchAutoSubDetail component.GlobalDB.Query(%s) error(%+v)", sql, err)
		return
	}
	m = make(map[int64]bool, 0)
	for rows.Next() {
		var teamID int64
		if scanErr := rows.Scan(&teamID); scanErr != nil {
			err = scanErr

			return
		}

		m[teamID] = true
	}

	err = rows.Err()

	return
}

func FetchAutoSubSeasonList(ctx context.Context) (list []int64, err error) {
	rows, err := component.GlobalDB.Query(ctx, sql4AutoSubscribeSeasons)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		log.Errorc(ctx, "FetchAutoSubSeasonList component.GlobalDB.Query(%s) error(%+v)", sql4AutoSubscribeSeasons, err)
		return
	}
	list = make([]int64, 0)
	for rows.Next() {
		var seasonId int64
		if scanErr := rows.Scan(&seasonId); scanErr != nil {
			err = scanErr

			return
		}
		list = append(list, seasonId)
	}

	err = rows.Err()

	return
}

func FetchAutoSubSeasonTeamContestIDMap(ctx context.Context, seasonIDList []int64) (m map[string][]int64, err error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(sql4SeasonTeamContestList, xstr.JoinInts(seasonIDList)))
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		log.Errorc(ctx, "FetchAutoSubSeasonList component.GlobalDB.Query(%s) error(%+v)", sql4SeasonTeamContestList, err)
		return
	}
	tmpM := make(map[string]map[int64]bool, 0)
	for rows.Next() {
		var contestID, seasonID, homeID, awayID, contestStatus, liveRoom int64
		if scanErr := rows.Scan(&contestID, &seasonID, &homeID, &awayID, &contestStatus, &liveRoom); scanErr != nil {
			err = scanErr
			return
		}
		//  过滤已过期赛程不能一键订阅
		if contestStatus != ContestStatusNotStart || liveRoom == 0 {
			continue
		}
		homeKey := GenAutoSubUniqKey(seasonID, homeID)
		homeM := make(map[int64]bool, 0)
		if d, ok := tmpM[homeKey]; ok {
			homeM = d
		}
		homeM[contestID] = true
		tmpM[homeKey] = homeM

		awayKey := GenAutoSubUniqKey(seasonID, awayID)
		awayM := make(map[int64]bool, 0)
		if d, ok := tmpM[awayKey]; ok {
			awayM = d
		}
		awayM[contestID] = true
		tmpM[awayKey] = awayM
	}
	err = rows.Err()
	m = genAutoSubSeasonTeamContestIDMap(tmpM)
	return
}

func genAutoSubSeasonTeamContestIDMap(tmpM map[string]map[int64]bool) map[string][]int64 {
	m := make(map[string][]int64)
	for k, v := range tmpM {
		tmpList := make([]int64, 0)
		for contestID := range v {
			tmpList = append(tmpList, contestID)
		}
		m[k] = tmpList
	}
	return m
}

func GenAutoSubUniqKey(seasonID, teamID int64) string {
	return fmt.Sprintf("%v_%v", seasonID, teamID)
}
