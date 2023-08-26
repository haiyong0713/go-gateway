package olympic

import (
	"context"
	xsql "database/sql"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/olympic"
	"time"
)

const (
	_olympicContestsSql = "select id, data, state from  act_web_data where vid = ?"
	_olympicContestSql  = "select id, data, state from  act_web_data where id = ? and vid = ?"
	_olympicQueryConfig = "select id, data, state from act_web_data where vid = ?"

	_olympicContestCacheKey    = "activity:olympic:contest:detailCache:id:%d"
	_olympicContestCacheKeyTtl = 300

	_olympicQueryConfigCacheKey = "activity:olympic:queryWord:configs:sourceId:%d"
	_olympicQueryConfigCacheTtl = 300

	_defaultCacheTtl      = 300 * time.Second
	_defaultRefreshTicker = 30
)

func (d *Dao) refreshValidContestCacheTicker(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if d.conf.OlympicConf != nil && d.conf.OlympicConf.RefreshCacheSeconds != 0 {
		duration = time.Duration(d.conf.OlympicConf.RefreshCacheSeconds) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			d.refreshValidContestCache(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Dao) refreshValidContestCache(ctx context.Context) {
	queryConfigs, err := d.GetQueryConfigs(ctx, d.conf.OlympicConf.QuerySourceId, false)
	if err != nil {
		return
	}
	for _, queryConfig := range queryConfigs {
		_ = d.refreshContestLocalCache(ctx, queryConfig.MatchId)
	}
	return
}

func (d *Dao) refreshContestLocalCache(ctx context.Context, id int64) (err error) {
	contest, err := d.GetOlympicContest(ctx, id, d.conf.OlympicConf.ContestSourceId, true, false)
	if err != nil {
		return
	}
	err = d.contestCache.SetWithExpire(id, contest, _defaultCacheTtl)
	if err != nil {
		log.Errorc(ctx, "[refreshContestLocalCache][SetWithExpire][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetOlympicContests(ctx context.Context, sourceId int64) (contests []*olympic.OlympicContest, err error) {
	contests = make([]*olympic.OlympicContest, 0)
	rows, err := d.db.Query(ctx, _olympicContestsSql, sourceId)
	if err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "[GetOlympicContests][Query][Error], err:%+v", err)
		return
	}
	if err == xsql.ErrNoRows {
		err = nil
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		contestDBData := new(olympic.OlympicDBData)
		if err = rows.Scan(&contestDBData.Id, &contestDBData.Data, &contestDBData.State); err != nil {
			log.Errorc(ctx, "[GetOlympicContests][Scan][Error], err:%+v", err)
			return
		}
		if contestDBData.Data == "" {
			continue
		}
		olympicContest := new(olympic.OlympicContest)
		if errG := json.Unmarshal([]byte(contestDBData.Data), &olympicContest); errG != nil {
			log.Errorc(ctx, "[GetOlympicContests][Unmarshal][Error], err:%+v", err)
			continue
		}
		olympicContest.Id = contestDBData.Id
		contests = append(contests, olympicContest)
	}
	return
}

func (d *Dao) GetOlympicContest(ctx context.Context, id int64, sourceId int64, skipLocal bool, skipCache bool) (contest *olympic.OlympicContest, err error) {
	if !skipLocal {
		contest, err = d.getOlympicContestFromLocal(ctx, id)
		if err == nil && contest != nil {
			return
		}
	}
	if !skipCache {
		contest, err = d.getOlympicContestFromCache(ctx, id)
		if err != nil && err != redis.ErrNil {
			return
		}
		if err == nil && contest != nil {
			return
		}
	}
	contest, err = d.getOlympicContestFromDB(ctx, id, sourceId)
	if err != nil {
		return
	}
	_ = d.setOlympicContestCache(ctx, contest)
	return
}

func formatOlympicContestCache(id int64) string {
	return fmt.Sprintf(_olympicContestCacheKey, id)
}

func (d *Dao) getOlympicContestFromDB(ctx context.Context, id int64, sourceId int64) (contest *olympic.OlympicContest, err error) {
	row := d.db.QueryRow(ctx, _olympicContestSql, id, sourceId)
	contestDBData := new(olympic.OlympicDBData)
	if err = row.Scan(&contestDBData.Id, &contestDBData.Data, &contestDBData.State); err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "[GetOlympicContests][Scan][Error], err:%+v", err)
		return
	}
	if err == xsql.ErrNoRows {
		err = xecode.Errorf(xecode.RequestErr, "记录不存在")
		return
	}
	if contestDBData.Data == "" {
		log.Errorc(ctx, "[GetOlympicContests][Data][Error], err:%+v, id:%d, dbData:%+v", err, id, contestDBData)
		err = xecode.Errorf(xecode.RequestErr, "记录不存在")
		return
	}
	contest = new(olympic.OlympicContest)
	if errG := json.Unmarshal([]byte(contestDBData.Data), &contest); errG != nil {
		log.Errorc(ctx, "[GetOlympicContests][Unmarshal][Error], err:%+v", err)
		err = xecode.Errorf(xecode.RequestErr, "记录不存在")
		return
	}
	contest.Id = contestDBData.Id
	return
}

func (d *Dao) getOlympicContestFromCache(ctx context.Context, id int64) (contest *olympic.OlympicContest, err error) {
	redisKey := formatOlympicContestCache(id)
	redisValue, err := redis.Bytes(d.redis.Do(ctx, "get", redisKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorc(ctx, "[getOlympicContestFromCache][Get][Error], err:%+v", err)
		}
		return
	}
	contest = new(olympic.OlympicContest)
	if err = json.Unmarshal(redisValue, &contest); err != nil {
		log.Errorc(ctx, "[getOlympicContestFromCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) setOlympicContestCache(ctx context.Context, contest *olympic.OlympicContest) (err error) {
	redisKey := formatOlympicContestCache(contest.Id)
	cacheValue, err := json.Marshal(contest)
	if err != nil {
		log.Errorc(ctx, "[setOlympicContestCache][SetEx][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx, "setEx", redisKey, _olympicContestCacheKeyTtl, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[setOlympicContestCache][SetEx][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) getOlympicContestFromLocal(ctx context.Context, id int64) (contest *olympic.OlympicContest, err error) {
	cacheValue, err := d.contestCache.Get(id)
	if err != nil {
		return
	}
	contest, ok := cacheValue.(*olympic.OlympicContest)
	if !ok {
		log.Errorc(ctx, "[GetOlympicContestFromLocal][Asert][Error], cache:%+v", cacheValue)
		err = xecode.Errorf(xecode.RequestErr, "类型错误")
		return
	}
	return
}
