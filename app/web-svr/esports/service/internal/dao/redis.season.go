package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func seasonCacheKey(seasonId int64) string {
	return fmt.Sprintf(seasonInfoCache, seasonId)
}

func (d *dao) GetSeasonInfoCache(ctx context.Context, seasonId int64) (seasonInfo *model.SeasonModel, err error) {
	redisKey := seasonCacheKey(seasonId)
	bytes, err := redis.Bytes(d.redis.Do(ctx, "get", redisKey))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][Redis][Get][Error], err:%s", err)
		return
	}
	if err == redis.ErrNil {
		err = nil
		return
	}
	if err = json.Unmarshal(bytes, &seasonInfo); err != nil {
		log.Errorc(ctx, "[Dao][Redis][Unmarshal][Error], err:%s", err)
		return
	}
	return
}
func (d *dao) SetSeasonInfoCache(ctx context.Context, seasonInfo *model.SeasonModel) (err error) {
	bytes, err := json.Marshal(seasonInfo)
	if err != nil {
		log.Errorc(ctx, "[Dao][SetSeasonInfoCache][Marshal][Error], err:%+v", err)
		return
	}
	redisKey := seasonCacheKey(seasonInfo.ID)

	if _, err = d.redis.Do(ctx, "setEx", redisKey, seasonInfoCacheTTL, bytes); err != nil {
		log.Errorc(ctx, "[Dao][SetSeasonInfoCache][SETEX][Error], err:%+v", err)
	}
	return
}

func (d *dao) SetSeasonsInfoCache(ctx context.Context, seasonModels []*model.SeasonModel) (err error) {
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)

	args := redis.Args{}
	for _, v := range seasonModels {
		args = args.Add(fmt.Sprintf(seasonInfoCache, v.ID))
		cacheValue, errG := json.Marshal(v)
		if errG != nil {
			err = errG
			return
		}
		args = args.Add(cacheValue)
	}
	err = conn.Send("mset", args...)
	if err != nil {
		log.Errorc(ctx, "[Dao][SetSeasonsInfoCache][MSET][Error], err:%+v", err)
		return
	}
	for _, v := range seasonModels {
		seasonKey := fmt.Sprintf(seasonInfoCache, v.ID)
		err = conn.Send("expire", seasonKey, seasonInfoCacheTTL)
		if err != nil {
			log.Errorc(ctx, "[Dao][SetSeasonsInfoCache][Expire][Error], err:%+v", err)
			return
		}
	}
	err = conn.Flush()
	if err != nil {
		log.Errorc(ctx, "[Dao][SetSeasonsInfoCache][Flush][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) DeleteSeasonCache(ctx context.Context, seasonId int64) (err error) {
	redisKey := seasonCacheKey(seasonId)
	_, err = d.redis.Do(ctx, "del", redisKey)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][DeleteSeasonCache][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) GetSeasonsInfoCache(ctx context.Context, seasonIds []int64) (seasonsInfo map[int64]*model.SeasonModel, missIds []int64, err error) {
	args := redis.Args{}
	for _, seasonId := range seasonIds {
		args = args.Add(seasonCacheKey(seasonId))
	}

	var bss [][]byte
	if bss, err = redis.ByteSlices(d.redis.Do(ctx, "MGET", args...)); err != nil {
		log.Errorc(ctx, "[Dao][Redis] d.redis.Do(ctx, MGET) error(%v)", err)
		return
	}
	seasonsInfo = make(map[int64]*model.SeasonModel)
	missIds = make([]int64, 0)
	for index, bs := range bss {
		season := new(model.SeasonModel)
		if len(bs) == 0 {
			missIds = append(missIds, seasonIds[index])
			continue
		}
		if err = json.Unmarshal(bs, season); err != nil {
			log.Error("GetTeamsCache json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		seasonsInfo[season.ID] = season
	}
	return
}
