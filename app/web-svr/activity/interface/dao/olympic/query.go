package olympic

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/olympic"
)

func (d *Dao) GetQueryConfigs(ctx context.Context, sourceId int64, skipCache bool) (queryConfigs []*olympic.OlympicQueryConfig, err error) {
	if !skipCache {
		queryConfigs, err = d.getQueryConfigsFromCache(ctx, sourceId)
		if err != nil && err != redis.ErrNil {
			return
		}
		if err == nil && queryConfigs != nil {
			return
		}
	}
	queryConfigs, err = d.GetQueryConfigsFromDB(ctx, sourceId)
	if err != nil {
		return
	}
	_ = d.setQueryConfigsCache(ctx, sourceId, queryConfigs)
	return
}

func (d *Dao) getQueryConfigsFromCache(ctx context.Context, sourceId int64) (queryConfigs []*olympic.OlympicQueryConfig, err error) {
	queryConfigs = make([]*olympic.OlympicQueryConfig, 0)
	redisKey := fmt.Sprintf(_olympicQueryConfigCacheKey, sourceId)
	cacheValue, err := redis.Bytes(d.redis.Do(ctx, "get", redisKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorc(ctx, "[getQueryConfigsFromCache][Get][Error], err:%+v", err)
		}
		return
	}
	if err = json.Unmarshal(cacheValue, &queryConfigs); err != nil {
		log.Errorc(ctx, "[getQueryConfigsFromCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) setQueryConfigsCache(ctx context.Context, sourceId int64, queryConfigs []*olympic.OlympicQueryConfig) (err error) {
	redisKey := fmt.Sprintf(_olympicQueryConfigCacheKey, sourceId)
	cacheValue, err := json.Marshal(queryConfigs)
	if err != nil {
		log.Errorc(ctx, "[setQueryConfigsCache][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx, "setEx", redisKey, _olympicQueryConfigCacheTtl, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[setQueryConfigsCache][setEx][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetQueryConfigsFromDB(ctx context.Context, sourceId int64) (queryConfigs []*olympic.OlympicQueryConfig, err error) {
	queryConfigs = make([]*olympic.OlympicQueryConfig, 0)
	rows, err := d.db.Query(ctx, _olympicQueryConfig, sourceId)
	if err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "[GetQueryConfigs][Query][Error], err:%+v", err)
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
		olympicQueryConfig := new(olympic.OlympicQueryConfig)
		if errG := json.Unmarshal([]byte(contestDBData.Data), &olympicQueryConfig); errG != nil {
			log.Errorc(ctx, "[GetOlympicContests][Unmarshal][Error], err:%+v", err)
			continue
		}
		olympicQueryConfig.State = contestDBData.State
		queryConfigs = append(queryConfigs, olympicQueryConfig)
	}
	return
}
