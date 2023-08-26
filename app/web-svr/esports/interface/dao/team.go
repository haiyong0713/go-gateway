package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const (
	bizLimitKeyOfTeam4DBRestore = "restore_team"
	bizName4TeamOfResetCache    = "cache_team"
)

func (d *Dao) TeamListByIDList(ctx context.Context, list []int64) (m map[int64]*model.Team, err error) {
	missedList := make([]int64, 0)

	m, missedList, err = d.TeamListFromCacheByIDList(list)
	if err != nil {
		return
	}

	if len(missedList) > 0 {
		tool.AddDBBackSourceMetricsByKeyList(bizLimitKeyOfTeam4DBRestore, missedList)

		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizLimitKeyOfTeam4DBRestore) {
			missedM, dbErr := d.RawEpTeams(ctx, missedList)
			if dbErr == nil {
				if len(missedM) > 0 {
					for k, v := range missedM {
						m[k] = v
					}

					_ = d.ResetTeamCacheByList(missedM)
				} else {
					tool.AddDBNoResultMetricsByKeyList(bizLimitKeyOfTeam4DBRestore, missedList)
				}
			} else {
				tool.AddDBNoResultMetricsByKeyList(bizLimitKeyOfTeam4DBRestore, missedList)
			}
		}
	}

	return
}

func (d *Dao) TeamListFromCacheByIDList(list []int64) (m map[int64]*model.Team, missedList []int64, err error) {
	m = make(map[int64]*model.Team, 0)
	missedList = make([]int64, 0)
	if len(list) == 0 {
		return
	}

	args := redis.Args{}
	for _, v := range list {
		args = args.Add(cacheKey4Team(v))
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

		tmp := new(model.Team)
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

func (d *Dao) ResetTeamCacheByList(list map[int64]*model.Team) (failedList []int64) {
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
		if _, cacheErr := conn.Do("SETEX", cacheKey4Team(v.ID), tool.CalculateExpiredSeconds(0), bs); cacheErr != nil {
			failedList = append(failedList, v.ID)
			tool.Metric4CacheResetFailed.WithLabelValues([]string{bizName4TeamOfResetCache, tool.CacheOfRemote}...).Inc()
		}
	}

	return
}

func (d *Dao) ResetTeamCacheByIDList(list []int64) (failedList []int64, err error) {
	failedList = make([]int64, 0)
	m, dbErr := d.RawEpTeams(context.Background(), list)
	if dbErr != nil {
		err = dbErr

		return
	}

	if len(m) == 0 {
		return
	}

	failedList = d.ResetTeamCacheByList(m)

	return
}

func (d *Dao) ResetTeamsInSeasonBySeasonIds(list []int64) (failedList []int64, err error) {
	failedList = make([]int64, 0)
	teamInSeasons, err := d.GetTeamsInSeasonFromDB(context.Background(), list)
	if err != nil {
		return failedList, err
	}

	for seasonId, teamsInSeason := range teamInSeasons {
		if err = d.AddTeamsInSeasonToCache(context.Background(), seasonId, teamsInSeason); err != nil {
			failedList = append(failedList, seasonId)
		}
	}
	return failedList, err
}
