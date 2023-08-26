package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const (
	bizLimitKeyOfContest4DBRestore = "restore_contest"
	bizName4ContestOfResetCache    = "cache_contest"
	cacheKey4SeasonMatchIDList     = "season:%v:matchID:list"

	limitKey2FetchMatchIDListUnderSeason = "season_matchID_list"

	secondsOfSevenDays = 7 * 86400

	sql4RecentContestIDList = `
SELECT id
FROM es_contests
WHERE stime > ?
`
	sql4ContestIDListBySeasonID = `
SELECT id
FROM es_contests
WHERE sid = ?
    AND status = 0
`
)

func currentDateUnix() int64 {
	year, month, day := time.Now().Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
}

func cacheKey4SeasonMatchIDListBySeasonID(seasonID int64) string {
	return fmt.Sprintf(cacheKey4SeasonMatchIDList, seasonID)
}

func FetchContestIDListBySeasonID(ctx context.Context, seasonID int64) (list []int64, err error) {

	return
}

func DeleteSeasonMatchIDListCacheBySeasonID(ctx context.Context, seasonID int64) (err error) {
	cacheKey := cacheKey4SeasonMatchIDListBySeasonID(seasonID)
	err = component.GlobalMemcached.Delete(ctx, cacheKey)

	return
}

func FetchContestIDListBySeasonIDFromCache(ctx context.Context, seasonID int64) (list []int64, err error) {
	list = make([]int64, 0)
	cacheKey := cacheKey4SeasonMatchIDListBySeasonID(seasonID)
	err = component.GlobalMemcached.Get(ctx, cacheKey).Scan(&list)

	return
}

func FetchContestIDListBySeasonIDFromDB(ctx context.Context, seasonID int64) (list []int64, err error) {
	list = make([]int64, 0)
	var rows *xsql.Rows

	rows, err = component.GlobalDBOfMaster.Query(ctx, sql4ContestIDListBySeasonID, seasonID)
	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return
		}

		list = append(list, id)
	}

	err = rows.Err()

	return
}

func (d *Dao) RecentContestIDList(ctx context.Context) (list []int64, err error) {
	list = make([]int64, 0)
	var rows *xsql.Rows

	rows, err = d.db.Query(ctx, sql4RecentContestIDList, currentDateUnix()-secondsOfSevenDays)
	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return
		}

		list = append(list, id)
	}

	err = rows.Err()

	return
}

func (d *Dao) ContestListByIDList(ctx context.Context, list []int64) (m map[int64]*model.Contest, err error) {
	missedList := make([]int64, 0)

	m, missedList, err = d.ContestListFromCacheByIDList(list)
	if err != nil {
		return
	}

	if len(missedList) > 0 {
		tool.AddDBBackSourceMetricsByKeyList(bizLimitKeyOfContest4DBRestore, missedList)
		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizLimitKeyOfContest4DBRestore) {
			missedM, dbErr := d.RawEpContests(ctx, missedList)
			if dbErr == nil {
				if len(missedM) > 0 {
					for k, v := range missedM {
						m[k] = v
					}

					_ = d.ResetContestCacheByList(missedM)
				} else {
					tool.AddDBNoResultMetricsByKeyList(bizLimitKeyOfContest4DBRestore, missedList)
				}
			} else {
				tool.AddDBErrMetricsByKeyList(bizLimitKeyOfContest4DBRestore, missedList)
			}
		}
	}

	return
}

func (d *Dao) ContestListFromCacheByIDList(list []int64) (m map[int64]*model.Contest, missedList []int64, err error) {
	m = make(map[int64]*model.Contest, 0)
	missedList = make([]int64, 0)

	if len(list) == 0 {
		return
	}

	args := redis.Args{}
	for _, v := range list {
		args = args.Add(cacheKey4Contest(v))
	}

	conn := d.redis.Get(context.Background())
	defer func() {
		_ = conn.Close()
	}()

	bsList, mGetErr := redis.ByteSlices(conn.Do("MGET", args...))
	if mGetErr != nil {
		if mGetErr != redis.ErrNil {
			err = mGetErr

			return
		}
	}

	for _, v := range bsList {
		if v == nil {
			continue
		}

		tmp := new(model.Contest)
		if jsonErr := json.Unmarshal(v, tmp); jsonErr == nil {
			m[tmp.ID] = tmp
		} else {
			// TODO: do not add this record in missed list
		}
	}

	for _, v := range list {
		if _, ok := m[v]; !ok && v > 0 {
			missedList = append(missedList, v)
		}
	}

	return
}

func (d *Dao) ResetContestCacheByList(list map[int64]*model.Contest) (failedList []int64) {
	failedList = make([]int64, 0)
	if len(list) == 0 {
		return
	}

	conn := d.redis.Get(context.Background())
	defer func() {
		_ = conn.Close()
	}()
	for _, v := range list {
		bs, _ := json.Marshal(v)
		if _, cacheErr := conn.Do("SETEX", cacheKey4Contest(v.ID), tool.CalculateExpiredSeconds(0), bs); cacheErr != nil {
			failedList = append(failedList, v.ID)
			tool.Metric4CacheResetFailed.WithLabelValues([]string{bizName4ContestOfResetCache, tool.CacheOfRemote}...).Inc()
		}
	}

	return
}

func (d *Dao) ResetContestCacheByIDList(list []int64) (failedList []int64, err error) {
	failedList = make([]int64, 0)
	m, dbErr := d.RawEpContests(context.Background(), list)
	if dbErr != nil {
		err = dbErr

		return
	}

	if len(m) == 0 {
		return
	}

	failedList = d.ResetContestCacheByList(m)

	return
}
