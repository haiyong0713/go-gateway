package guess

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/activity/interface/model/guess"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const (
	bizLimitKey4DBRestoreOfMainDetail = "restore_guess_main_details"

	bizNameOfMainDetail = "guess_main_details"

	cacheKey4GuessMainDetailList = "guess:main:detail:1016:%v:%v"
)

func (d *Dao) HotMainDetailListByMainIDList(ctx context.Context, mainIDList []int64, business int64) (m map[string]map[int64]*guess.MainRes, err error) {
	m = make(map[string]map[int64]*guess.MainRes, 0)
	tmpM := make(map[int64]*guess.MainRes, 0)
	tmpM, err = d.AvailableHotMainDetailMap(ctx, mainIDList, business)
	if err == nil {
		for _, v := range tmpM {
			key := v.GenHotMapKey()
			if _, ok := m[key]; !ok {
				m[key] = make(map[int64]*guess.MainRes, 0)
			}
			m[key][v.ID] = v
		}
	}

	return
}

func (d *Dao) DetailListByMainIDList(ctx context.Context, mainIDList []int64, business int64) (m map[int64]*guess.MainRes, err error) {
	m = make(map[int64]*guess.MainRes, 0)
	missedMainIDList := make([]int64, 0)
	canRestore := false

	m, canRestore, missedMainIDList, err = d.FetchDetailListFromCacheByIDList(ctx, mainIDList, business)
	if err != nil {
		return
	}

	if !canRestore {
		return
	}

	if canRestore && len(missedMainIDList) > 0 {
		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizLimitKey4DBRestoreOfMainDetail) {
			tool.AddDBBackSourceMetrics(bizNameOfMainDetail)

			missedM := make(map[int64]*guess.MainRes, 0)
			missedM, err = d.RawMDsResult(ctx, missedMainIDList, business)
			if err != nil {
				tool.AddDBErrMetrics(bizNameOfMainDetail)

				return
			}

			restoreM := make(map[int64]*guess.MainRes, 0)
			for _, v := range missedMainIDList {
				if d, ok := missedM[v]; ok {
					restoreM[v] = d.DeepCopy()
				} else {
					tmp := guess.NewMainRes()
					{
						tmp.ID = v
					}

					restoreM[v] = tmp
				}
			}

			_, _ = d.ResetGuessAggregationInfoMap(ctx, restoreM, business)

			for k, v := range missedM {
				m[k] = v
			}
		}
	}

	return
}

func (d *Dao) ResetGuessAggregationInfoMap(ctx context.Context, m map[int64]*guess.MainRes, business int64) (succeedNum, failedNum int64) {
	if len(m) == 0 {
		return
	}

	for _, v := range m {
		if err := d.ResetGuessAggregationInfoInCache(ctx, v, business); err != nil {
			failedNum++
		} else {
			succeedNum++
		}
	}

	return
}

func (d *Dao) ResetGuessAggregationInfoInCacheByMainID(ctx context.Context, mainID, business int64) (err error) {
	info := guess.NewMainRes()
	info, err = d.RawMDResult(ctx, mainID, business)
	if err != nil {
		return
	}

	for i := 0; i < 3; i++ {
		err = d.ResetGuessAggregationInfoInCache(ctx, info, business)
		if err == nil {
			break
		}
	}

	return
}

func (d *Dao) DeleteGuessAggregationInfoCache(ctx context.Context, mainID, business int64) (err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	for i := 0; i < 3; i++ {
		_, err = conn.Do("DEL", guessMainDetailListCacheKey(mainID, business))
		if err == nil {
			break
		}
	}

	return
}

func (d *Dao) ResetGuessAggregationInfoInCache(ctx context.Context, info *guess.MainRes, business int64) (err error) {
	bs := make([]byte, 0)
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	if info.ID == 0 {
		bs = []byte("")
	} else {
		bs, _ = json.Marshal(info)
	}

	status := tool.StatusOfSucceed
	_, err = conn.Do("SETEX", guessMainDetailListCacheKey(info.ID, business), tool.CalculateExpiredSeconds(0), bs)
	if err != nil {
		status = tool.StatusOfFailed
	}

	tool.IncrCacheResetMetric(bizLimitKey4DBRestoreOfMainDetail, status)

	return
}

func guessMainDetailListCacheKey(mainID, business int64) string {
	return fmt.Sprintf(cacheKey4GuessMainDetailList, mainID, business)
}

func (d *Dao) FetchDetailListFromCacheByIDList(ctx context.Context, mainIDList []int64, business int64) (
	m map[int64]*guess.MainRes, canRestore bool, missedMainIDList []int64, err error) {
	m = make(map[int64]*guess.MainRes, 0)
	missedMainIDList = make([]int64, 0)
	noRecordMainIDM := make(map[int64]bool, 0)
	bsList := make([][]byte, 0)
	args := redis.Args{}

	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	for _, mainID := range mainIDList {
		args = args.Add(guessMainDetailListCacheKey(mainID, business))
	}

	bsList, err = redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			canRestore = true
			err = nil
			missedMainIDList = mainIDList[:]
		}

		return
	}

	for _, bs := range bsList {
		if bs == nil {
			continue
		}

		mainRes := new(guess.MainRes)
		if unmarshalErr := json.Unmarshal(bs, &mainRes); unmarshalErr == nil {
			m[mainRes.ID] = mainRes
		} else {
			noRecordMainIDM[mainRes.ID] = true
		}
	}

	for _, mainID := range mainIDList {
		if _, ok := m[mainID]; !ok {
			if _, ok := noRecordMainIDM[mainID]; !ok {
				missedMainIDList = append(missedMainIDList, mainID)
			}
		}
	}

	if len(missedMainIDList) > 0 {
		canRestore = true
	}

	return
}
