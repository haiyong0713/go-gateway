package dao

import (
	"context"
	"go-gateway/app/web-svr/esports/interface/model"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/web-svr/esports/job/sql"
)

const (
	limit4HotData = 100

	sql4HotSeasonList = `
SELECT id, guess_version
FROM es_seasons
WHERE stime <= ?
	AND etime >= ?
    AND id > ?
ORDER BY id ASC
LIMIT ?
`
	_sqlTeamsInSeasonSQL = `SELECT t.id, 
       t.title, 
       t.region_id, 
       r.sid, 
       r.rank,
       t.logo
FROM   es_teams t, 
       es_team_in_seasons r 
WHERE  r.sid = ?
       AND r.tid = t.id 
       AND t.is_deleted = 0 
ORDER  BY r.rank DESC, 
          t.id ASC `
)

func FetchAllHotSeasonList(ctx context.Context) (m map[int64]int64, err error) {
	var startID int64
	m = make(map[int64]int64, 0)
	for {
		d, tmpErr := fetchHotSeasonListByStartID(ctx, startID, limit4HotData)
		if tmpErr != nil {
			err = tmpErr

			return
		}

		if len(d) > 0 {
			for k, v := range d {
				m[k] = v
			}

			startID = d[int64(len(d)-1)]
		}

		if len(d) < limit4HotData {
			break
		}
	}

	return
}

func fetchHotSeasonListByStartID(ctx context.Context, startID, limit int64) (m map[int64]int64, err error) {
	var rows *xsql.Rows
	m = make(map[int64]int64, 0)
	now := time.Now()
	startTime := now.Add(30 * 24 * time.Hour)
	endTime := now.Add(-30 * 24 * time.Hour)
	rows, err = sql.GlobalDB.Query(ctx, sql4HotSeasonList, startTime, endTime, startID, limit)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "fetchHotSeasonListByStartID rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		var seasonID, guessVersion int64
		if scanErr := rows.Scan(&seasonID, &guessVersion); scanErr == nil {
			m[seasonID] = guessVersion
		}
	}
	return
}

func (d *Dao) GetTeamsInSeasonFromDB(c context.Context, seasonId int64) (res []*model.TeamInSeason, err error) {
	res = make([]*model.TeamInSeason, 0)
	rows, err := d.db.Query(c, _sqlTeamsInSeasonSQL, seasonId)
	if err != nil {
		log.Errorc(c, "query teams_in_season from db error: %v", err)
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(c, "GetTeamsInSeasonFromDB rows error: %v", err)
			return
		}
	}()
	for rows.Next() {
		tmp := &model.TeamInSeason{}
		if err = rows.Scan(&tmp.TeamId, &tmp.TeamTitle, &tmp.RegionId, &tmp.SeasonId, &tmp.Rank, &tmp.Logo); err != nil {
			log.Errorc(c, "scan teams_in_season error: %v", err)
			return
		}
		res = append(res, tmp)
	}
	return

}
