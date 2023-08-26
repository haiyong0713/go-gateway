package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

type EsContestsCache struct {
	ContestIds []int64 `json:"contest_ids"`
	Total      int     `json:"total"`
}

func (d *dao) GetEsContestIdsCache(ctx context.Context, md5Str string) (contestIds []int64, total int, err error) {
	esContestsCache := new(EsContestsCache)
	esContestsCache.ContestIds = make([]int64, 0)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get", fmt.Sprintf(esContestCache, md5Str)))
	if err != nil {
		log.Errorc(ctx, "[Redis][GetEsContestIdsCache][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &esContestsCache); err != nil {
		log.Errorc(ctx, "[Redis][GetEsContestIdsCache][Unmarshal][Error], err:%+v", err)
		return
	}
	contestIds = esContestsCache.ContestIds
	total = esContestsCache.Total
	return
}

func (d *dao) SetEsContestIdsCache(ctx context.Context, md5Str string, contestIds []int64, total int) (err error) {
	esContestsCache := &EsContestsCache{
		ContestIds: contestIds,
		Total:      total,
	}
	cacheValue, err := json.Marshal(esContestsCache)
	if err != nil {
		log.Errorc(ctx, "[Redis][GetEsContestIdsCache][Marshal][Error], err:%+v", err)
		return
	}
	_, err = redis.Bytes(
		d.redis.Do(
			ctx,
			"setEx",
			fmt.Sprintf(esContestCache, md5Str),
			esContestCacheTTL,
			cacheValue,
		),
	)
	if err != nil {
		log.Errorc(ctx, "[Redis][GetEsContestIdsCache][Error], err:%+v", err)
		return
	}
	return
}
