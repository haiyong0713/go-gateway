package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) GetMatchCache(ctx context.Context, matchId int64) (matchModel *model.MatchModel, err error) {
	bytes, err := redis.Bytes(d.redis.Do(ctx, "HGET", matchInfoMapCache, matchId))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][Redis][GetMatchCache][HGET][Error], err:%+v", err)
		return
	}
	matchModel = new(model.MatchModel)
	if len(bytes) == 0 {
		return
	}
	if err = json.Unmarshal(bytes, &matchModel); err != nil {
		log.Errorc(ctx, "[Dao][Redis][GetMatchCache][Unmarshal][Error], err:%+v", err)
	}
	return
}

func (d *dao) SetMatchCache(ctx context.Context, matchModel model.MatchModel) (err error) {
	cacheValue, err := json.Marshal(matchModel)
	if err != nil {
		return
	}
	_, err = d.redis.Do(ctx, "HSET", matchInfoMapCache, matchModel.ID, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][SetMatchCache][HSet][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) DeleteMatchCache(ctx context.Context, matchId int64) (err error) {
	_, err = d.redis.Do(ctx, "hdel", matchInfoMapCache, matchId)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][DeleteMatchCache][Error], err:%+v", err)
		return
	}
	return
}
