package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) GetSeriesCacheById(ctx context.Context, seriesId int64) (seriesModel *model.ContestSeriesModel, err error) {
	redisKey := fmt.Sprintf(contestSeriesInfoCache, seriesId)
	bytes, err := redis.Bytes(d.redis.Do(ctx, "get", redisKey))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][Redis][GetMatchCache][HGET][Error], err:%+v", err)
		return
	}
	seriesModel = new(model.ContestSeriesModel)
	if len(bytes) == 0 {
		return
	}
	if err = json.Unmarshal(bytes, &seriesModel); err != nil {
		log.Errorc(ctx, "[Dao][Redis][GetSeriesCache][Unmarshal][Error], err:%+v", err)
	}
	return
}

func (d *dao) SetSeriesCacheById(ctx context.Context, seriesModel *model.ContestSeriesModel) (err error) {
	redisKey := fmt.Sprintf(contestSeriesInfoCache, seriesModel.ID)
	cacheValue, err := json.Marshal(seriesModel)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][SetSeriesCacheById][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx, "setEx", redisKey, contestSeriesInfoCacheTTL, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][SetSeriesCacheById][Set][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) GetSeasonSeriesListCache(ctx context.Context, seasonId int64) (seriesModels []*model.ContestSeriesModel, err error) {

	redisKey := fmt.Sprintf(seasonContestSeriesListCache, seasonId)
	bytes, err := redis.Bytes(d.redis.Do(ctx, "get", redisKey))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][Redis][GetMatchCache][GEt][Error], err:%+v", err)
		return
	}
	seriesModels = make([]*model.ContestSeriesModel, 0)
	if len(bytes) == 0 {
		return
	}
	if err = json.Unmarshal(bytes, &seriesModels); err != nil {
		log.Errorc(ctx, "[Dao][Redis][GetSeasonSeriesList][Unmarshal][Error], err:%+v", err)
	}
	return
}

func (d *dao) SetSeasonSeriesListCache(ctx context.Context, seriesModels []*model.ContestSeriesModel) (err error) {
	if len(seriesModels) == 0 {
		return
	}

	redisKey := fmt.Sprintf(seasonContestSeriesListCache, seriesModels[0].SeasonId)
	cacheValue, err := json.Marshal(seriesModels)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][SetSeasonSeriesList][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx, "setEx", redisKey, seasonContestSeriesListCacheTTL, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][SetSeasonSeriesList][Set][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) GetSeriesCacheByIds(ctx context.Context, seriesIds []int64) (seriesModelMap map[int64]*model.ContestSeriesModel, missIds []int64, err error) {
	seriesModelMap = make(map[int64]*model.ContestSeriesModel)
	args := redis.Args{}
	for _, v := range seriesIds {
		args = args.Add(fmt.Sprintf(contestSeriesInfoCache, v))
	}
	byteSlices, err := redis.ByteSlices(d.redis.Do(ctx, "mget", args...))
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][GetContestCache][Error], err:%+v", err)
		return
	}
	missIds = make([]int64, 0)
	for index, bytes := range byteSlices {
		if len(bytes) == 0 {
			missIds = append(missIds, seriesIds[index])
			continue
		}
		contestSeriesModel := new(model.ContestSeriesModel)
		if err = json.Unmarshal(bytes, &contestSeriesModel); err != nil {
			log.Errorc(ctx, "[Dao][Unmarshal][GetContestCache][Error], err:%+v, cache:%s", err, string(bytes))
			return
		}
		seriesModelMap[contestSeriesModel.ID] = contestSeriesModel
	}
	return
}

func (d *dao) SetSeriesCacheByIds(ctx context.Context, seriesModels []*model.ContestSeriesModel) (err error) {
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)
	contestKeys := make([]string, 0)
	args := redis.Args{}
	for _, v := range seriesModels {
		contestKeys = append(contestKeys, fmt.Sprintf(contestSeriesInfoCache, v.ID))
		args = args.Add(contestKeys)
		cacheValue, errG := json.Marshal(v)
		if errG != nil {
			err = errG
			return
		}
		args = args.Add(cacheValue)
	}
	err = conn.Send("mset", args...)
	if err != nil {
		log.Errorc(ctx, "[Dao][SetContestSeriesCache][MSET][Error], err:%+v", err)
		return
	}
	for _, v := range seriesModels {
		contestSeriesKey := fmt.Sprintf(contestInfoCache, v.ID)
		err = conn.Send("expire", contestSeriesKey, contestSeriesInfoCacheTTL)
		if err != nil {
			log.Errorc(ctx, "[Dao][SetContestSeriesCache][Expire][Error], err:%+v", err)
			return
		}
	}
	err = conn.Flush()
	if err != nil {
		log.Errorc(ctx, "[Dao][SetContestSeriesCache][Flush][Error], err:%+v", err)
		return
	}
	return
}
