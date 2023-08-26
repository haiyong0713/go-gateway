package dao

import (
	"context"
	"encoding/json"

	"go-gateway/app/web-svr/esports/service/internal/model"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

func (d *dao) GetGamesCache(ctx context.Context, gameIds []int64) (gamesInfoMap map[int64]*model.GameModel, missIds []int64, err error) {
	byteSlices, err := redis.ByteSlices(d.redis.Do(ctx, "HMGET", gameInfoMapCache, gameIds))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][Redis][GetGamesCache][HMGET][Error], err:%+v", err)
		return
	}
	gamesInfoMap = make(map[int64]*model.GameModel)
	missIds = make([]int64, 0)
	for index, bytes := range byteSlices {
		if len(bytes) == 0 {
			missIds = append(missIds, gameIds[index])
			continue
		}
		gameInfo := new(model.GameModel)
		if err = json.Unmarshal(bytes, &gameInfo); err != nil {
			missIds = append(missIds, gameIds[index])
			log.Errorc(ctx, "[Dao][Redis][GetGamesCache][Unmarshal][Error], err:%+v", err)
			continue
		}
		gamesInfoMap[gameInfo.ID] = gameInfo
	}
	return
}

func (d *dao) SetGamesCache(ctx context.Context, gamesInfoMap map[int64]*model.GameModel) (err error) {
	if len(gamesInfoMap) == 0 {
		return
	}
	args := redis.Args{}.Add(gameInfoMapCache)
	for gameId, gameInfo := range gamesInfoMap {
		cacheValue, errG := json.Marshal(gameInfo)
		if errG != nil {
			log.Errorc(ctx, "[Dao][Redis][SetGamesCache][Marshal][Error], err:%+v", err)
			err = errG
			return
		}
		args = args.Add(gameId).Add(cacheValue)
	}
	_, err = d.redis.Do(ctx, "HMSET", args...)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][GetGamesCache][HMSET][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) DeleteGameCache(ctx context.Context, gameId int64) (err error) {
	_, err = d.redis.Do(ctx, "hdel", gameInfoMapCache, gameId)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][DeleteGameCache][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) GetAllGamesCache(ctx context.Context) (gamesInfoMap map[int64]*model.GameModel, err error) {
	gamesInfoMap = make(map[int64]*model.GameModel)
	cacheRes, err := redis.StringMap(d.redis.Do(ctx, "hGetAll", gameInfoMapCache))
	if err != nil {
		log.Errorc(ctx, "[Dao][GetAllGamesCache][Error], err:%+v", err)
		return
	}

	for _, stringValue := range cacheRes {
		gameInfo := new(model.GameModel)
		err = json.Unmarshal([]byte(stringValue), &gameInfo)
		if err != nil {
			log.Errorc(ctx, "[Dao][GetAllGamesCache][Unmarshal][Error], err:%+v", err)
			return
		}
		gamesInfoMap[gameInfo.ID] = gameInfo
	}
	return
}
