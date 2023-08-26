package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const (
	bizLimitKeyOfSeason4DBRestore = "restore_season"
	bizName4SeasonOfResetCache    = "cache_season"
	bizName4TeamsInSeason         = "teams_in_season"
)

func (d *Dao) SeasonListByIDList(list []int64) (m map[int64]*model.Season, err error) {
	missedList := make([]int64, 0)

	m, missedList, err = d.SeasonListFromCacheByIDList(list)
	if err != nil {
		return
	}

	if len(missedList) > 0 {
		tool.AddDBBackSourceMetricsByKeyList(bizLimitKeyOfSeason4DBRestore, missedList)
		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizLimitKeyOfSeason4DBRestore) {
			missedM, dbErr := d.RawEpSeasons(context.Background(), missedList)
			if dbErr == nil {
				if len(missedM) > 0 {
					for k, v := range missedM {
						m[k] = v
					}

					_ = d.ResetSeasonCacheByList(missedM)
				} else {
					tool.AddDBNoResultMetricsByKeyList(bizLimitKeyOfSeason4DBRestore, missedList)
				}
			} else {
				tool.AddDBErrMetricsByKeyList(bizLimitKeyOfSeason4DBRestore, missedList)
			}
		}
	}

	return
}

func (d *Dao) SeasonListFromCacheByIDList(list []int64) (m map[int64]*model.Season, missedList []int64, err error) {
	m = make(map[int64]*model.Season, 0)
	missedList = make([]int64, 0)

	if len(list) == 0 {
		return
	}

	args := redis.Args{}
	for _, v := range list {
		args = args.Add(cacheKey4Season(v))
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

		tmp := new(model.Season)
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

func (d *Dao) ResetSeasonCacheByList(list map[int64]*model.Season) (failedList []int64) {
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
		if _, cacheErr := conn.Do("SETEX", cacheKey4Season(v.ID), tool.CalculateExpiredSeconds(0), bs); cacheErr != nil {
			failedList = append(failedList, v.ID)
			tool.Metric4CacheResetFailed.WithLabelValues([]string{bizName4SeasonOfResetCache, tool.CacheOfRemote}...).Inc()
		}
	}

	return
}

func (d *Dao) ResetSeasonCacheByIDList(list []int64) (failedList []int64, err error) {
	failedList = make([]int64, 0)
	m, dbErr := d.RawEpSeasons(context.Background(), list)
	if dbErr != nil {
		err = dbErr

		return
	}

	if len(m) == 0 {
		return
	}

	failedList = d.ResetSeasonCacheByList(m)

	return
}
