package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) GetContestCache(ctx context.Context, contestId int64) (contestModel *model.ContestModel, err error) {
	contestKey := fmt.Sprintf(contestInfoCache, contestId)
	bytes, err := redis.Bytes(d.redis.Do(ctx, "get", contestKey))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][Redis][GetContestCache][Error], err:%+v", err)
		return
	}
	if err == redis.ErrNil {
		err = nil
	}
	contestModel = new(model.ContestModel)
	if len(bytes) == 0 {
		return
	}
	if err = json.Unmarshal(bytes, &contestModel); err != nil {
		log.Errorc(ctx, "[Dao][Unmarshal][GetContestCache][Error], err:%+v", err)
		return
	}
	return
}
func (d *dao) GetContestsCache(ctx context.Context, contestIds []int64) (contestModelMap map[int64]*model.ContestModel, missIds []int64, err error) {
	args := redis.Args{}
	for _, v := range contestIds {
		args = args.Add(fmt.Sprintf(contestInfoCache, v))
	}
	byteSlices, err := redis.ByteSlices(d.redis.Do(ctx, "mget", args...))
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][GetContestCache][Error], err:%+v", err)
		return
	}
	contestModelMap = make(map[int64]*model.ContestModel)
	missIds = make([]int64, 0)
	for index, bytes := range byteSlices {
		if len(bytes) == 0 {
			missIds = append(missIds, contestIds[index])
			continue
		}
		contestModel := new(model.ContestModel)
		if err = json.Unmarshal(bytes, &contestModel); err != nil {
			log.Errorc(ctx, "[Dao][Unmarshal][GetContestCache][Error], err:%+v, cache:%s", err, string(bytes))
			return
		}
		contestModelMap[contestModel.ID] = contestModel
	}
	return
}
func (d *dao) DeleteContestCache(ctx context.Context, contestId int64) (err error) {
	_, err = d.redis.Do(ctx, "DEL", fmt.Sprintf(contestInfoCache, contestId))
	if err != nil {
		log.Errorc(ctx, "[Dao][DeleteContestCache][DEL][Error], err:%+v", err)
	}
	return
}

func (d *dao) SetContestCache(ctx context.Context, contestModels []*model.ContestModel) (err error) {
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)

	args := redis.Args{}
	for _, v := range contestModels {
		args = args.Add(fmt.Sprintf(contestInfoCache, v.ID))
		cacheValue, errG := json.Marshal(v)
		if errG != nil {
			err = errG
			return
		}
		args = args.Add(cacheValue)
	}
	err = conn.Send("mset", args...)
	if err != nil {
		log.Errorc(ctx, "[Dao][SetContestCache][MSET][Error], err:%+v", err)
		return
	}
	for _, v := range contestModels {
		contestKey := fmt.Sprintf(contestInfoCache, v.ID)
		err = conn.Send("expire", contestKey, contestInfoCacheTTL)
		if err != nil {
			log.Errorc(ctx, "[Dao][SetContestCache][Expire][Error], err:%+v", err)
			return
		}
	}
	err = conn.Flush()
	if err != nil {
		log.Errorc(ctx, "[Dao][SetContestCache][Flush][Error], err:%+v", err)
		return
	}
	return
}
